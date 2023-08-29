package resource

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
)

type listHandler struct {
	models *model.Models
}

func ListHandler(conf *config.ServerConfig) api.Handler {
	return &listHandler{models: conf.Models}
}

func (h *listHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	workspaceId, err := utils.ParseUint(c.Param("workspaceId"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopePipeline,
		ScopeId: workspaceId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *listHandler) Handle(c *api.Context) *utils.Response {
	workspaceId, err := utils.ParseUint(c.Param("workspaceId"))
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	resources, err := h.models.PipelineResourceManager.List(workspaceId)
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}
	return c.ResponseOK(resources)
}
