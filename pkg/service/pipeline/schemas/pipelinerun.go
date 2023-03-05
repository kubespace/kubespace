package schemas

type JobCallbackParams struct {
	JobId  uint   `json:"job_id"`
	Status string `json:"status"`
}

type AddReleaseVersionParams struct {
	WorkspaceId uint   `json:"workspace_id"`
	JobId       uint   `json:"job_id"`
	Version     string `json:"version"`
}
