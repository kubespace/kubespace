package serializers

import (
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
)

type WorkspaceSerializer struct {
	Name         string `json:"name" form:"name"`
	Type         string `json:"type" form:"type"`
	Description  string `json:"description" form:"description"`
	ApiUrl       string `json:"api_url" form:"api_url"`
	CodeUrl      string `json:"code_url" form:"code_url"`
	CodeType     string `json:"code_type" form:"code_type"`
	CodeSecretId uint   `json:"code_secret_id" form:"code_secret_id"`
}

type WorkspaceUpdateSerializer struct {
	Description  string `json:"description" form:"description"`
	CodeSecretId uint   `json:"code_secret_id" form:"code_secret_id"`
}

type WorkspaceListSerializer struct {
	WithPipeline bool   `json:"with_pipeline" form:"with_pipeline"`
	Type         string `json:"type" form:"type"`
}

type WorkspaceReleaseSerializer struct {
	WorkspaceId uint   `json:"workspace_id" form:"workspace_id"`
	Version     string `json:"version" form:"version"`
}

type PipelineSerializer struct {
	ID          uint                      `json:"id"`
	WorkspaceId uint                      `json:"workspace_id"`
	Name        string                    `json:"name"`
	Sources     types.PipelineSources     `json:"sources"`
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
	ID           uint                   `json:"id"`
	Name         string                 `json:"name"`
	TriggerMode  string                 `json:"trigger_mode"`
	CustomParams map[string]interface{} `json:"custom_params"`
	Jobs         types.PipelineJobs     `json:"jobs"`
}

type PipelineListSerializer struct {
	WorkspaceId uint `json:"workspace_id" form:"workspace_id"`
}

type PipelineBuildListSerializer struct {
	PipelineId      uint   `json:"pipeline_id" form:"pipeline_id"`
	LastBuildNumber int    `json:"last_build_number" form:"last_build_number"`
	Status          string `json:"status" form:"status"`
	Limit           int    `json:"limit" form:"limit" default:"20"`
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
	StageRunId   uint                              `json:"stage_run_id"`
	CustomParams map[string]interface{}            `json:"custom_params"`
	JobParams    map[string]map[string]interface{} `json:"job_params"`
}

type PipelineStageRetrySerializer struct {
	StageRunId uint `json:"stage_run_id"`
}

type PipelineResourceSerializer struct {
	WorkspaceId uint   `json:"workspace_id" form:"workspace_id"`
	Global      bool   `json:"global" form:"global"`
	Name        string `json:"name" form:"name"`
	Type        string `json:"type" form:"type"`
	Value       string `json:"value" form:"value"`
	SecretId    uint   `json:"secret_id" form:"secret_id"`
	Description string `json:"description" form:"description"`
}
