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

type existReleaseHandler struct {
	models *model.Models
}

func ExistReleaseHandler(conf *config.ServerConfig) api.Handler {
	return &existReleaseHandler{models: conf.Models}
}

type existReleaseForm struct {
	WorkspaceId uint   `json:"workspace_id" form:"workspace_id"`
	Version     string `json:"version" form:"version"`
}

func (h *existReleaseHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	var form existReleaseForm
	if err := c.ShouldBindQuery(&form); err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopePipeline,
		ScopeId: form.WorkspaceId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *existReleaseHandler) Handle(c *api.Context) *utils.Response {
	var form existReleaseForm
	if err := c.ShouldBind(&form); err != nil {
		return c.ResponseError(errors.New(code.ParseError, err))
	}
	exists, err := h.models.PipelineReleaseManager.ExistsRelease(form.WorkspaceId, form.Version)
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}
	return c.ResponseOK(map[string]interface{}{"exists": exists})
}
