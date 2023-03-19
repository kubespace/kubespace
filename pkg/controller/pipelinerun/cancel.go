package pipelinerun

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/model/manager/pipeline"
	"github.com/kubespace/kubespace/pkg/model/types"
)

func (p *PipelineRunController) cancelLockKey(id uint) string {
	return fmt.Sprintf("pipeline_run_controller:build:cancel:%d", id)
}

// Check 检查流水线构建状态是否取消
func (p *PipelineRunController) cancelCheck(object interface{}) bool {
	pipelineRun, ok := object.(types.PipelineRun)
	if !ok {
		return false
	}
	if pipelineRun.Status != types.PipelineStatusCancel {
		return false
	}
	if locked, _ := p.lock.Locked(p.cancelLockKey(pipelineRun.ID)); locked {
		// 该流水线构建已存在锁，正在被执行
		return false
	}
	return true
}

func (p *PipelineRunController) cancel(object interface{}) (err error) {
	pipelineRun := object.(types.PipelineRun)
	// 对流水线构建执行加锁，保证只有一个goroutinue执行
	if ok, _ := p.lock.Acquire(p.cancelLockKey(pipelineRun.ID)); !ok {
		return nil
	}
	// 执行完成释放锁
	defer p.lock.Release(p.cancelLockKey(pipelineRun.ID))
	if latestPipelineRun, err := p.models.PipelineRunManager.Get(pipelineRun.ID); err != nil {
		return err
	} else {
		pipelineRun = *latestPipelineRun
	}

	if pipelineRun.Status != types.PipelineStatusCancel {
		return fmt.Errorf("pipeline run id=%d status=%s, do not cancel", pipelineRun.ID, pipelineRun.Status)
	}
	stages, err := p.models.PipelineRunManager.StagesRun(pipelineRun.ID)
	for _, stageRun := range stages {
		if stageRun.Status != types.PipelineStatusCancel {
			continue
		}
		for _, jobRun := range stageRun.Jobs {
			// 已经执行完成的任务不取消
			if jobRun.Status == types.PipelineStatusOK || jobRun.Status == types.PipelineStatusError {
				continue
			}
			// 取消任务执行
			if err = p.jobRun.Cancel(jobRun.ID); err != nil {
				return err
			}
			// 取消之后，任务未取消完成，状态修改为cancel取消中，任务退出后，在build流程会将状态修改为canceled
			jobRun.Status = types.PipelineStatusCancel
			// 及时更新到数据库
			if _, _, err = p.models.PipelineRunManager.UpdatePipelineStageRun(&pipeline.UpdateStageObj{
				StageRunId:   stageRun.ID,
				StageRunJobs: []*types.PipelineRunJob{jobRun},
			}); err != nil {
				return err
			}
		}
		// 更新阶段状态为canceled已取消，controller不会处理canceled状态
		_, _, err = p.models.PipelineRunManager.UpdatePipelineStageRun(&pipeline.UpdateStageObj{
			StageRunId:     stageRun.ID,
			StageRunStatus: types.PipelineStatusCanceled,
		})
		return err
	}

	return nil
}
