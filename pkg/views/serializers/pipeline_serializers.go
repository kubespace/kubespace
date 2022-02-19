package serializers

import (
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
)

type WorkspaceSerializer struct {
	Name         string `json:"name" form:"name"`
	Type         string `json:"type" form:"type"`
	Description  string `json:"description" form:"description"`
	CodeUrl      string `json:"code_url" form:"code_url"`
	CodeType     string `json:"code_type" form:"code_type"`
	CodeSecretId uint   `json:"code_secret_id" form:"code_secret_id"`
}

type PipelineSerializer struct {
	ID          uint                      `json:"id"`
	WorkspaceId uint                      `json:"workspace_id"`
	Name        string                    `json:"name"`
	Triggers    types.PipelineTriggers    `json:"triggers"`
	Stages      []PipelineStageSerializer `json:"stages"`
}

type PipelineTrigger struct {
	Type        string                      `json:"type"`
	Expressions []PipelineTriggerExpression `json:"expressions"`
}

type PipelineTriggerExpression struct {
	Key      string `json:"key"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type PipelineStageSerializer struct {
	ID          uint               `json:"id"`
	Name        string             `json:"name"`
	TriggerMode string             `json:"trigger_mode"`
	Jobs        types.PipelineJobs `json:"jobs"`
}

type PipelineListSerializer struct {
	WorkspaceId uint `json:"workspace_id" form:"workspace_id"`
}

type PipelineBuildListSerializer struct {
	PipelineId uint `json:"pipeline_id" form:"pipeline_id"`
}

type PipelineBuildSerializer struct {
	PipelineId uint                   `json:"pipeline_id"`
	Params     map[string]interface{} `json:"params"`
}

type PipelineCallbackSerializer struct {
	JobId  uint            `json:"job_id"`
	Result *utils.Response `json:"result"`
}

type PipelineStageManualSerializer struct {
	StageRunId uint `json:"stage_run_id"`
}

type PipelineStageRetrySerializer struct {
	StageRunId uint `json:"stage_run_id"`
}