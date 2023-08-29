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

type PipelineBody struct {
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
