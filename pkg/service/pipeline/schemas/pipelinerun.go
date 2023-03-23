package schemas

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
	PipelineId uint                   `json:"pipeline_id"`
	Params     map[string]interface{} `json:"params"`
}

//type PipelineBuildCode
