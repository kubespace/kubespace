package schemas

type ListGitReposParams struct {
	SecretId uint   `json:"secret_id" form:"secret_id"`
	GitType  string `json:"git_type" form:"git_type"`
	ApiUrl   string `json:"api_url" form:"api_url"`
}

type ListRepoBranchParams struct {
	PipelineId uint `json:"pipeline_id" form:"pipeline_id"`
}
