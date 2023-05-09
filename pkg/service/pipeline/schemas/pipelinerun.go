package schemas

import (
	"github.com/kubespace/kubespace/pkg/model/types"
)

// JobCallbackParams 任务执行完成回调参数
type JobCallbackParams struct {
	JobId  uint   `json:"job_id"`
	Status string `json:"status"`
}

// AddReleaseVersionParams 发布阶段执行时添加版本
type AddReleaseVersionParams struct {
	WorkspaceId uint   `json:"workspace_id"`
	JobId       uint   `json:"job_id"`
	Version     string `json:"version"`
}

// PipelineStageCancelParams 流水线构建阶段取消
type PipelineStageCancelParams struct {
	StageRunId uint `json:"stage_run_id"`
}

// PipelineStageReexecParams 流水线构建阶段重新执行
type PipelineStageReexecParams struct {
	StageRunId uint `json:"stage_run_id"`
}

// PipelineBuildParams 流水线构建参数
type PipelineBuildParams struct {
	*types.PipelineBuildConfig `json:",inline"`
	PipelineId                 uint `json:"pipeline_id"`
}

type PipelineParams struct {
	ID          uint                  `json:"id"`
	WorkspaceId uint                  `json:"workspace_id"`
	Name        string                `json:"name"`
	Sources     types.PipelineSources `json:"sources"`
	Triggers    []*PipelineTrigger    `json:"triggers"`
	Stages      []*PipelineStage      `json:"stages"`
}

type PipelineTrigger struct {
	Id uint `json:"id"`
	// 触发类型，
	Type string `json:"type"`
	// 定时配置
	Cron string `json:"cron,omitempty"`
}

type PipelineStage struct {
	ID           uint                   `json:"id"`
	Name         string                 `json:"name"`
	TriggerMode  string                 `json:"trigger_mode"`
	CustomParams map[string]interface{} `json:"custom_params"`
	Jobs         types.PipelineJobs     `json:"jobs"`
}
