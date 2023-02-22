package pipeline_job

import "github.com/gin-gonic/gin"

type JobExecutor struct {
	DataDir string
}

func NewJobExecutor(dataDir string) *JobExecutor {
	return &JobExecutor{
		DataDir: dataDir,
	}
}

func (j *JobExecutor) Execute(c *gin.Context) {

}

func (j *JobExecutor) Status(c *gin.Context) {

}

func (j *JobExecutor) Log(c *gin.Context) {

}

func (j *JobExecutor) Cleanup(c *gin.Context) {

}

func (j *JobExecutor) Cancel(c *gin.Context) {

}
