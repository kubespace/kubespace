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

type getHandler struct {
	models *model.Models
}

func GetHandler(conf *config.ServerConfig) api.Handler {
	return &getHandler{models: conf.Models}
}

func (h *getHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	id, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopePipeline,
		ScopeId: id,
		Role:    types.RoleViewer,
	}, nil
}

func (h *getHandler) Handle(c *api.Context) *utils.Response {
	workspaceId, _ := utils.ParseUint(c.Param("id"))
	workspace, err := h.models.PipelineWorkspaceManager.Get(workspaceId)
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, "获取流水线空间失败："+err.Error()))
	}
	return c.ResponseOK(workspace)
}
