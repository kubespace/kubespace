package pipespace

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
)

type latestReleaseHandler struct {
	models *model.Models
}

func LatestReleaseHandler(conf *config.ServerConfig) api.Handler {
	return &latestReleaseHandler{models: conf.Models}
}

type latestReleaseForm struct {
	WorkspaceId uint `json:"workspace_id" form:"workspace_id"`
}

func (h *latestReleaseHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	var form latestReleaseForm
	if err := c.ShouldBindQuery(&form); err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopePipeline,
		ScopeId: form.WorkspaceId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *latestReleaseHandler) Handle(c *api.Context) *utils.Response {
	var form latestReleaseForm
	if err := c.ShouldBind(&form); err != nil {
		return c.ResponseError(errors.New(code.ParseError, err))
	}
	rel, err := h.models.PipelineReleaseManager.GetLatestRelease(form.WorkspaceId)
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}
	return c.ResponseOK(rel)
}
