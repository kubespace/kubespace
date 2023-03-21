package pipeline_trigger

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kubespace/kubespace/pkg/model/types"
	pipelineservice "github.com/kubespace/kubespace/pkg/service/pipeline"
	utilgit "github.com/kubespace/kubespace/pkg/third/git"
	"k8s.io/klog/v2"
)

func (p *PipelineTriggerController) triggerLockKey(id uint) string {
	return fmt.Sprintf("pipeline_trigger_controller:trigger:%d", id)
}

// 流水线触发配置检查
func (p *PipelineTriggerController) triggerCheck(obj interface{}) bool {
	trigger, ok := obj.(types.PipelineTrigger)
	if !ok {
		return false
	}
	if locked, _ := p.lock.Locked(p.triggerLockKey(trigger.ID)); locked {
		// 正在处理该流水线配置
		return false
	}
	return true
}

// 流水线触发处理
func (p *PipelineTriggerController) triggerHandle(obj interface{}) error {
	trigger := obj.(types.PipelineTrigger)
	// 对流水线配置处理加锁，保证只有一个goroutinue执行
	if ok, _ := p.lock.Acquire(p.triggerLockKey(trigger.ID)); !ok {
		return nil
	}
	// 执行完成释放锁
	defer p.lock.Release(p.triggerLockKey(trigger.ID))

	pipeline, err := p.models.PipelineManager.Get(trigger.PipelineId)
	if err != nil {
		return err
	}
	pipelineWorkspace, err := p.models.PipelineWorkspaceManager.Get(pipeline.WorkspaceId)
	if err != nil {
		return err
	}
	if trigger.Type == types.PipelineTriggerTypeCode {
		return p.triggerCodePipeline(pipelineWorkspace, pipeline, &trigger)
	}
	return nil
}

// 触发代码空间流水线
func (p *PipelineTriggerController) triggerCodePipeline(
	workspace *types.PipelineWorkspace,
	pipeline *types.Pipeline,
	trigger *types.PipelineTrigger) error {
	// 流水线触发代码分支配置
	var triggerConfig types.PipelineTriggerConfigCode
	if err := json.Unmarshal(trigger.Config, &triggerConfig); err != nil {
		return err
	}
	if workspace.Code == nil {
		return fmt.Errorf("pipeline workspace id=%d name=%s no secret", workspace.ID, workspace.Name)
	}
	secret, err := p.models.SettingsSecretManager.Get(workspace.Code.SecretId)
	if err != nil {
		return fmt.Errorf("获取代码密钥失败：" + err.Error())
	}
	gitcli, err := utilgit.NewClient(workspace.Code.Type, workspace.Code.ApiUrl, secret.GetSecret())
	if err != nil {
		return err
	}
	// 获取所有分支以及分支的最新提交commitId
	branches, err := gitcli.ListRepoBranches(context.Background(), workspace.Code.CloneUrl)
	if err != nil {
		return err
	}
	first := false
	if triggerConfig.BranchLatestCommit == nil {
		// 如果还没有触发分支的记录，则是第一次初始化，不进行事件触发
		triggerConfig.BranchLatestCommit = make(map[string]*types.PipelineTriggerConfigCodeBranch)
		first = true
	}
	for _, branch := range branches {
		currCommit, ok := triggerConfig.BranchLatestCommit[branch.Name]
		if !ok || currCommit.CommitId != branch.Ref {
			// 如果没有记录或者当前记录的commitId与代码库不一致，则更新该commitId
			latestCommit, err := gitcli.GetBranchLatestCommit(context.Background(), workspace.Code.CloneUrl, branch.Name)
			if err != nil {
				klog.Errorf("get code %s branch=%s latest commit info error: %s", workspace.Code.CloneUrl, branch.Name, err.Error())
				continue
			}
			configBranch := &types.PipelineTriggerConfigCodeBranch{
				Branch:     branch.Name,
				CommitId:   latestCommit.CommitId,
				Author:     latestCommit.Author,
				Message:    latestCommit.Message,
				CommitTime: latestCommit.CommitTime,
			}
			triggerConfig.BranchLatestCommit[branch.Name] = configBranch
			if !first && pipelineservice.MatchBranchSource(pipeline.Sources, branch.Name) {
				// 如果不是第一次初始化，且匹配当前流水线代码源规则，则发送触发事件
			}
		}
	}
	// 更新triggerConfig到数据库
	return nil
}
