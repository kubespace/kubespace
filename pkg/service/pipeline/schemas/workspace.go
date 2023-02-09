package schemas

type ListGitReposParams struct {
	SecretId uint   `json:"secret_id" form:"secret_id"`
	GitTye   string `json:"git_type" form:"git_type"`
	ApiUrl   string `json:"api_url" form:"api_url"`
}
