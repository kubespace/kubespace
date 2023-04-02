package pipeline_trigger

import (
	"context"
	"fmt"
	"github.com/kubespace/kubespace/pkg/model/manager/pipeline"
	"github.com/kubespace/kubespace/pkg/model/types"
	utilgit "github.com/kubespace/kubespace/pkg/third/git"
	"k8s.io/klog/v2"
	"time"
)

func (p *PipelineTriggerController) codeCacheLockKey(id uint) string {
	return fmt.Sprintf("pipeline_trigger_controller:event:%d", id)
}

// 代码分支缓存检查
func (p *PipelineTriggerController) codeCacheCheck(obj interface{}) bool {
	cache, ok := obj.(types.PipelineCodeCache)
	if !ok {
		return false
	}
	if locked, _ := p.lock.Locked(p.codeCacheLockKey(cache.ID)); locked {
		// 正在缓存该代码分支
		return false
	}
	if cache.Status != types.PipelineCodeCacheStatusOpen {
		return false
	}
	return true
}

// 对流水线代码空间分支获取最新提交，并进行缓存
func (p *PipelineTriggerController) codeCacheHandle(obj interface{}) error {
	cache := obj.(types.PipelineCodeCache)
	// 对流水线配置处理加锁，保证只有一个goroutinue执行
	if ok, _ := p.lock.Acquire(p.codeCacheLockKey(cache.ID)); !ok {
		return nil
	}
	// 执行完成释放锁
	defer p.lock.Release(p.codeCacheLockKey(cache.ID))
	workspace, err := p.models.PipelineWorkspaceManager.Get(cache.WorkspaceId)
	if err != nil {
		klog.Errorf("get workspace id=%d error: %s", cache.WorkspaceId, err.Error())
		return err
	}
	klog.V(1).Infof("start cache code branch workspace id=%d name=%s", workspace.ID, workspace.Name)

	if workspace.Code == nil {
		klog.Errorf("pipeline workspace id=%d name=%s no secret", workspace.ID, workspace.Name)
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

	// 流水线触发代码分支配置
	commitCache := cache.CommitCache
	if commitCache == nil {
		commitCache = &types.CodeBranchCommitCache{
			BranchLatestCommit: make(map[string]*types.PipelineBuildCodeBranch),
		}
	}
	updated := false
	for _, branch := range branches {
		currCommit, ok := commitCache.BranchLatestCommit[branch.Name]
		if !ok || currCommit.CommitId != branch.CommitId {
			klog.Infof("branch=%s updated, curr commit id=%s, remote commit id=%s", branch, currCommit.CommitId, branch.CommitId)
			// 如果没有记录或者当前记录的commitId与代码库不一致，则更新该commitId
			latestCommit, err := gitcli.GetBranchLatestCommit(context.Background(), workspace.Code.CloneUrl, branch.Name)
			if err != nil {
				klog.Errorf("get code %s branch=%s latest commit info error: %s", workspace.Code.CloneUrl, branch.Name, err.Error())
				continue
			}
			branchCommit := &types.PipelineBuildCodeBranch{
				Branch:     branch.Name,
				CommitId:   latestCommit.CommitId,
				Author:     latestCommit.Author,
				Message:    latestCommit.Message,
				CommitTime: latestCommit.CommitTime,
			}
			commitCache.BranchLatestCommit[branch.Name] = branchCommit
			updated = true
		}
	}
	if updated {
		// 更新branch commit cache
		if err = p.models.PipelineCodeCacheManager.Update(cache.ID, &types.PipelineCodeCache{
			CommitCache: commitCache,
			UpdateTime:  time.Now(),
		}); err != nil {
			return err
		}
		triggerTime := time.Now()
		// 更新所有的触发
		return p.models.PipelineTriggerManager.UpdateTriggerTime(triggerTime, &pipeline.PipelineTriggerCondition{
			WorkspaceId: cache.WorkspaceId,
		})
	}
	return nil
}
