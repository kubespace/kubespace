package pipelinerun

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/controller/pipelinerun/job_run"
	"github.com/kubespace/kubespace/pkg/core/lock"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/manager/pipeline"
	"github.com/kubespace/kubespace/pkg/model/types"
	"strconv"
)

// CancelHandler 流水线构建取消处理
type CancelHandler struct {
	models *model.Models
	jobRun *job_run.JobRun
	// 流水线取消时对其进行加锁，保证只有一个进行处理
	lock lock.Lock
}

func NewCancelHandler(models *model.Models, jobRun *job_run.JobRun) *CancelHandler {
	return &CancelHandler{
		models: models,
		jobRun: jobRun,
		lock:   lock.NewMemLock(),
	}
}

// Check 检查流水线构建状态是否取消
func (p *CancelHandler) Check(object interface{}) bool {
	pipelineRun, ok := object.(types.PipelineRun)
	if !ok {
		return false
	}
	if pipelineRun.Status != types.PipelineStatusCancel {
		return false
	}
	if locked, _ := p.lock.Locked(strconv.Itoa(int(pipelineRun.ID))); locked {
		// 该流水线构建已存在锁，正在被执行
		return false
	}
	return true
}

func (p *CancelHandler) Handle(object interface{}) (err error) {
	pipelineRun := object.(types.PipelineRun)
	// 对流水线构建执行加锁，保证只有一个goroutinue执行
	if ok, _ := p.lock.Acquire(strconv.Itoa(int(pipelineRun.ID))); !ok {
		return nil
	}
	// 执行完成释放锁
	defer p.lock.Release(strconv.Itoa(int(pipelineRun.ID)))
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
		var cancelJobs []*types.PipelineRunJob
		for _, jobRun := range stageRun.Jobs {
			if jobRun.Status == types.PipelineStatusOK || jobRun.Status == types.PipelineStatusError {
				continue
			}
			if err = p.jobRun.Cancel(jobRun.ID); err != nil {
				return err
			}
			jobRun.Status = types.PipelineStatusCancel
			cancelJobs = append(cancelJobs, jobRun)
		}
		// 更新阶段以及流水线构建状态为canceled已取消
		_, _, err = p.models.PipelineRunManager.UpdatePipelineStageRun(&pipeline.UpdateStageObj{
			StageRunId:     stageRun.ID,
			StageRunStatus: types.PipelineStatusCanceled,
			StageRunJobs:   cancelJobs,
		})
		return err
	}

	return nil
}
