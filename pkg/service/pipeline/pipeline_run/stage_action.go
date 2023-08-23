package pipeline_run

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model/manager/pipeline"
	"github.com/kubespace/kubespace/pkg/model/types"
	"time"
)

const (
	// StageActionManualExec 手动触发执行
	StageActionManualExec = "manual_exec"

	// StageActionCancel 取消执行中阶段
	StageActionCancel = "cancel"

	// StageActionCancelReexec 重新执行取消阶段
	StageActionCancelReexec = "cancel_reexec"

	// StageActionReexec 重新执行已执行完成阶段
	StageActionReexec = "reexec"

	// StageActionErrorRetry 失败阶段重试
	StageActionErrorRetry = "error_retry"
)

// StageActionParams 阶段执行时参数
type StageActionParams struct {
	CustomParams map[string]interface{}
	JobParams    map[uint]map[string]interface{}
}

func (r *PipelineRunService) StageAction(
	action string,
	stageRunId uint,
	params StageActionParams) (*types.PipelineRun, *types.PipelineRunStage, error) {
	stageRun, err := r.models.PipelineRunManager.GetStageRun(stageRunId)
	if err != nil {
		return nil, nil, errors.New(code.DataNotExists, "获取执行阶段数据失败："+err.Error())
	}
	switch action {
	case StageActionManualExec:
		if stageRun.Status != types.PipelineStatusPause {
			msg := fmt.Sprintf("current stage run id=%v status is %v, not pause", stageRun.ID, stageRun.Status)
			return nil, nil, errors.New(code.StatusError, msg)
		}
		return r.executeStage(stageRun, &params)
	case StageActionCancelReexec:
		if stageRun.Status != types.PipelineStatusCanceled {
			msg := fmt.Sprintf("current stage run id=%v status is %v, not canceled", stageRun.ID, stageRun.Status)
			return nil, nil, errors.New(code.StatusError, msg)
		}
		return r.executeStage(stageRun, &params)
	case StageActionErrorRetry:
		if stageRun.Status != types.PipelineStatusError {
			msg := fmt.Sprintf("current stage run id=%v status is %v, not error", stageRun.ID, stageRun.Status)
			return nil, nil, errors.New(code.StatusError, msg)
		}
		return r.executeStage(stageRun, &params)
	case StageActionCancel:
		if stageRun.Status != types.PipelineStatusDoing {
			msg := fmt.Sprintf("current stage run id=%v status is %v, not running", stageRun.ID, stageRun.Status)
			return nil, nil, errors.New(code.StatusError, msg)
		}
		// 更新当前阶段状态为取消中，controller监听到取消中阶段后，给阶段中所有任务下发取消指令
		return r.models.PipelineRunManager.UpdatePipelineStageRun(&pipeline.UpdateStageObj{
			StageRunId:     stageRun.ID,
			StageRunStatus: types.PipelineStatusCancel,
		})
	case StageActionReexec:
		return r.models.PipelineRunManager.ReexecStage(stageRun.PipelineRunId, stageRunId, params.CustomParams, params.JobParams)
	default:
		return nil, nil, errors.New(code.ParamsError, fmt.Sprintf("param action=%s is unsupported", action))
	}
}

// 将阶段状态置为doing，controller监听到之后，执行该阶段
func (r *PipelineRunService) executeStage(
	stageRun *types.PipelineRunStage,
	params *StageActionParams) (*types.PipelineRun, *types.PipelineRunStage, error) {
	for i, job := range stageRun.Jobs {
		if jobParams, ok := params.JobParams[job.ID]; ok {
			stageRun.Jobs[i].Params = jobParams
		}
	}
	now := time.Now()
	return r.models.PipelineRunManager.UpdatePipelineStageRun(&pipeline.UpdateStageObj{
		StageRunId:     stageRun.ID,
		StageRunStatus: types.PipelineStatusDoing,
		StageExecTime:  &now,
		StageRunJobs:   stageRun.Jobs,
		CustomParams:   params.CustomParams,
	})
}
