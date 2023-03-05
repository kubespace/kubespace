package pipeline_job

import (
	"github.com/gin-gonic/gin"
	"github.com/kubespace/kubespace/pkg/spacelet/pipeline_job/plugins"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"net/http"
)

type JobExecutor struct {
	plugins *plugins.Plugins
	client  *utils.HttpClient
}

func NewJobExecutor(dataDir string, client *utils.HttpClient) *JobExecutor {
	return &JobExecutor{
		plugins: plugins.NewPlugins(dataDir),
		client:  client,
	}
}

type JobRunParams struct {
	JobId  uint                   `json:"job_id" form:"job_id"`
	Plugin string                 `json:"plugin" form:"plugin"`
	Params map[string]interface{} `json:"params" form:"params"`
}

func (j *JobExecutor) Execute(c *gin.Context) {
	var params JobRunParams
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusOK, &utils.Response{Code: code.ParamsError, Msg: err.Error()})
		return
	}
	pluginParams := &plugins.PluginParams{
		JobId:     params.JobId,
		PluginKey: params.Plugin,
		Params:    params.Params,
	}
	c.JSON(http.StatusOK, j.plugins.Execute(pluginParams))
}

type JobStatusParams struct {
	JobId   uint `json:"job_id" form:"job_id" url:"job_id"`
	WithLog bool `json:"with_log" form:"with_log" url:"with_log"`
}

func (j *JobExecutor) Status(c *gin.Context) {
	var params JobStatusParams
	if err := c.BindQuery(&params); err != nil {
		c.JSON(http.StatusOK, &utils.Response{Code: code.ParamsError, Msg: err.Error()})
		return
	}
	statusLog, err := j.plugins.GetStatusLog(params.JobId, params.WithLog)
	if err != nil {
		c.JSON(http.StatusOK, &utils.Response{Code: code.GetError, Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, &utils.Response{Code: code.Success, Data: statusLog})
}

type JobCleanParams struct {
	JobId uint `json:"job_id" form:"job_id" url:"job_id"`
}

func (j *JobExecutor) Cleanup(c *gin.Context) {
	var params JobStatusParams
	if err := c.BindQuery(&params); err != nil {
		c.JSON(http.StatusOK, &utils.Response{Code: code.ParamsError, Msg: err.Error()})
		return
	}
	err := j.plugins.Cleanup(params.JobId)
	if err != nil {
		c.JSON(http.StatusOK, &utils.Response{Code: code.GetError, Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, &utils.Response{Code: code.Success})
}

func (j *JobExecutor) Cancel(c *gin.Context) {}
