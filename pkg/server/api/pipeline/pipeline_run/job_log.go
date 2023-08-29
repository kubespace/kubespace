package pipeline_run

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
)

type jobLogHandler struct {
	models *model.Models
}

func JobLogHandler(conf *config.ServerConfig) api.Handler {
	return &jobLogHandler{models: conf.Models}
}

func (h *jobLogHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	workspaceId, _ := utils.ParseUint(c.Query("workspace_id"))
	return true, &api.AuthPerm{
		Scope:   types.ScopePipeline,
		ScopeId: workspaceId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *jobLogHandler) Handle(c *api.Context) *utils.Response {
	jobRunId, err := utils.ParseUint(c.Param("jobRunId"))
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	jobLog, err := h.models.PipelineRunManager.GetJobRunLog(jobRunId, true)
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}
	if jobLog == nil {
		return c.ResponseOK("")
	}
	return c.ResponseOK(jobLog.Logs)
}
