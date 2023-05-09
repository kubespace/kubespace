package pipeline_trigger

import (
	"database/sql"
	"fmt"
	"github.com/kubespace/kubespace/pkg/model/types"
	pipelineservice "github.com/kubespace/kubespace/pkg/service/pipeline"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
	"time"
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
		// 正在处理该流水线触发配置
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
	if triggerObj, err := p.models.PipelineTriggerManager.Get(trigger.ID); err != nil {
		return err
	} else if triggerObj == nil {
		return fmt.Errorf("not found trigger object id=%d", trigger.ID)
	} else {
		trigger = *triggerObj
	}
	if trigger.NextTriggerTime == nil || trigger.NextTriggerTime.Time.After(time.Now()) {
		return nil
	}

	pipeline, err := p.models.PipelineManager.Get(trigger.PipelineId)
	if err != nil {
		return err
	}
	pipelineWorkspace, err := p.models.PipelineWorkspaceManager.Get(pipeline.WorkspaceId)
	if err != nil {
		return err
	}
	klog.Infof("start trigger pipeline id=%d name=%s workspace name=%s", pipeline.ID, pipeline.Name, pipelineWorkspace.Name)
	if trigger.Type == types.PipelineTriggerTypeCode {
		return p.triggerCodePipeline(pipelineWorkspace, pipeline, &trigger)
	}
	return nil
}

// 代码空间流水线触发事件，是否有最新的代码提交，如果有，则生成一个新的流水线构建事件
func (p *PipelineTriggerController) triggerCodePipeline(
	workspace *types.PipelineWorkspace,
	pipeline *types.Pipeline,
	trigger *types.PipelineTrigger) error {
	codeCache, err := p.models.PipelineCodeCacheManager.GetByWorkspaceId(workspace.ID)
	if err != nil {
		return err
	}
	if codeCache == nil {
		klog.Infof("not found code cache with workspace id=%d", workspace.ID)
		return nil
	}
	if codeCache.CommitCache == nil || codeCache.CommitCache.BranchLatestCommit == nil {
		klog.Infof("code cache is empty and retry next time")
		return nil
	}
	commitCache := codeCache.CommitCache.BranchLatestCommit

	// 是否是第一次触发，第一次不生成触发事件
	first := false
	// 流水线触发代码分支配置
	triggerCodeConfig := trigger.Config.Code
	if triggerCodeConfig == nil {
		triggerCodeConfig = &types.PipelineTriggerConfigCode{}
	}
	if triggerCodeConfig.BranchLatestCommit == nil {
		// 如果还没有触发分支的记录，则是第一次初始化，不进行事件触发
		triggerCodeConfig.BranchLatestCommit = make(map[string]*types.PipelineBuildCodeBranch)
		first = true
	}
	updated := false
	for branch, commit := range commitCache {
		currCommit, ok := triggerCodeConfig.BranchLatestCommit[branch]
		if !ok || currCommit.CommitId != commit.CommitId {
			// 如果没有记录或者当前记录的commitId与代码库不一致，则更新该commitId

			if !first && pipelineservice.MatchBranchSource(pipeline.Sources, branch) {
				// 如果不是第一次初始化，且匹配当前流水线代码源规则，则生成触发事件
				if err := p.models.PipelineTriggerEventManager.Create(&types.PipelineTriggerEvent{
					PipelineId:  pipeline.ID,
					From:        types.PipelineTriggerEventFromTrigger,
					TriggerId:   trigger.ID,
					Status:      types.PipelineTriggerEventStatusNew,
					EventConfig: types.PipelineBuildConfig{CodeBranch: commitCache[branch]},
					TriggerUser: commit.Author,
					CreateTime:  time.Now(),
					UpdateTime:  time.Now(),
				}); err != nil {
					klog.Errorf("create pipeline trigger event error: %s", err.Error())
					continue
				}
			}
			triggerCodeConfig.BranchLatestCommit[branch] = commitCache[branch]
			updated = true
		}
	}
	if updated || trigger.NextTriggerTime != nil {
		// 更新triggerConfig到数据库
		return p.models.PipelineTriggerManager.Update(trigger.ID, &types.PipelineTrigger{
			Config:          types.PipelineTriggerConfig{Code: triggerCodeConfig},
			UpdateTime:      time.Now(),
			NextTriggerTime: &sql.NullTime{},
		})
	}
	return nil
}

// 定时触发
func (p *PipelineTriggerController) cronTrigger(
	workspace *types.PipelineWorkspace,
	pipeline *types.Pipeline,
	trigger *types.PipelineTrigger) (err error) {
	nextTriggerTime, err := utils.NextTriggerTime(trigger.Config.Cron.Cron)
	if err != nil {
		return err
	}
	if workspace.Type == types.WorkspaceTypeCode {
		err = p.cronTriggerCodePipeline(workspace, pipeline, trigger)
	} else if workspace.Type == types.WorkspaceTypeCustom {
		err = p.cronTriggerCustomPipeline(workspace, pipeline, trigger)
	}
	if err != nil {
		return err
	}
	// 修改下次触发时间
	return p.models.PipelineTriggerManager.Update(trigger.ID, &types.PipelineTrigger{
		UpdateTime:      time.Now(),
		NextTriggerTime: &sql.NullTime{Time: nextTriggerTime},
	})
}

// 代码流水线定时触发，获取流水线所有分支最新的一个提交作为触发源
func (p *PipelineTriggerController) cronTriggerCodePipeline(
	workspace *types.PipelineWorkspace,
	pipeline *types.Pipeline,
	trigger *types.PipelineTrigger) error {
	codeCache, err := p.models.PipelineCodeCacheManager.GetByWorkspaceId(workspace.ID)
	if err != nil {
		return err
	}
	if codeCache == nil {
		klog.Infof("not found code cache with workspace id=%d", workspace.ID)
		return nil
	}
	if codeCache.CommitCache == nil || codeCache.CommitCache.BranchLatestCommit == nil {
		klog.Infof("code cache is empty and retry next time")
		return nil
	}
	commitCache := codeCache.CommitCache.BranchLatestCommit
	var latestCommit *types.PipelineBuildCodeBranch
	for branch, commit := range commitCache {
		if pipelineservice.MatchBranchSource(pipeline.Sources, branch) {
			if latestCommit == nil || latestCommit.CommitTime.Before(commit.CommitTime) {
				latestCommit = commitCache[branch]
			}
		}
	}
	if latestCommit == nil {
		klog.Infof("not found pipeline branch sources commits")
		return nil
	}
	if err = p.models.PipelineTriggerEventManager.Create(&types.PipelineTriggerEvent{
		PipelineId:  pipeline.ID,
		From:        types.PipelineTriggerEventFromTrigger,
		TriggerId:   trigger.ID,
		Status:      types.PipelineTriggerEventStatusNew,
		EventConfig: types.PipelineBuildConfig{CodeBranch: latestCommit},
		TriggerUser: latestCommit.Author,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}); err != nil {
		klog.Errorf("create pipeline trigger event error: %s", err.Error())
		return err
	}
	return nil
}

// 自定义流水线定时触发，获取源流水线最新构建作为触发源
func (p *PipelineTriggerController) cronTriggerCustomPipeline(
	workspace *types.PipelineWorkspace,
	pipeline *types.Pipeline,
	trigger *types.PipelineTrigger) error {
	var buildSources []*types.PipelineBuildCustomSource
	for _, source := range pipeline.Sources {
		pipelineBuilds, err := p.models.PipelineRunManager.ListPipelineRun(source.Pipeline, 0, types.PipelineStatusOK, 1)
		if err != nil {
			return err
		}
		if len(pipelineBuilds) <= 0 {
			// 没有成功的构建记录
			continue
		}
		buildSources = append(buildSources, &types.PipelineBuildCustomSource{
			WorkspaceId:         ,
			WorkspaceName:       "",
			PipelineId:          0,
			PipelineName:        "",
			BuildReleaseVersion: "",
			BuildId:             0,
			BuildNumber:         0,
			BuildOperator:       "",
			IsBuild:             false,
		})
	}
	return nil
}
