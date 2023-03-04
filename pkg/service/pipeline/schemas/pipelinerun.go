package schemas

type JobCallbackParams struct {
	JobId  uint   `json:"job_id"`
	Status string `json:"status"`
}
