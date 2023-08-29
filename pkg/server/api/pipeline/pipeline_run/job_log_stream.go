package pipeline_run

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
	"time"
)

type jobLogStreamHandler struct {
	models *model.Models
}

func JobLogStreamHandler(conf *config.ServerConfig) api.Handler {
	return &jobLogStreamHandler{models: conf.Models}
}

func (h *jobLogStreamHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	workspaceId, _ := utils.ParseUint(c.Query("workspace_id"))
	return true, &api.AuthPerm{
		Scope:   types.ScopePipeline,
		ScopeId: workspaceId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *jobLogStreamHandler) Handle(c *api.Context) *utils.Response {
	jobRunId, err := utils.ParseUint(c.Param("jobRunId"))
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	lastJobLog, err := h.models.PipelineRunManager.GetJobRunLog(jobRunId, true)
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	c.SSEvent("message", "\n")
	if lastJobLog != nil {
		c.SSEvent("message", lastJobLog.Logs)
	}
	c.Writer.Flush()

	tick := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-c.Writer.CloseNotify():
			return nil
		case <-tick.C:
			currJobLog, err := h.models.PipelineRunManager.GetJobRunLog(jobRunId, false)
			if err != nil {
				klog.Errorf("get job id=%s log error: %s", jobRunId, err.Error())
				c.SSEvent("message", "get job log error: "+err.Error())
				c.Writer.Flush()
				continue
			}
			if currJobLog == nil {
				continue
			}
			if lastJobLog == nil || currJobLog.UpdateTime != lastJobLog.UpdateTime {
				lastJobLog, err = h.models.PipelineRunManager.GetJobRunLog(jobRunId, true)
				if err != nil {
					klog.Errorf("get job id=%s log error: %s", jobRunId, err.Error())
				} else {
					c.SSEvent("message", lastJobLog.Logs)
					c.Writer.Flush()
				}
			}

		}
	}
}
