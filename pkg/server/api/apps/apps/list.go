package apps

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	projectservice "github.com/kubespace/kubespace/pkg/service/project"
	"github.com/kubespace/kubespace/pkg/utils"
)

type listHandler struct {
	models     *model.Models
	appService *projectservice.AppService
}

func ListHandler(conf *config.ServerConfig) api.Handler {
	return &listHandler{
		models:     conf.Models,
		appService: conf.ServiceFactory.Project.AppService,
	}
}

type listAppForm struct {
	Scope   string `json:"scope" form:"scope"`
	ScopeId uint   `json:"scope_id" form:"scope_id"`
}

func (h *listHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	var form listAppForm
	if err := c.ShouldBindQuery(&form); err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   form.Scope,
		ScopeId: form.ScopeId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *listHandler) Handle(c *api.Context) *utils.Response {
	var form listAppForm
	if err := c.ShouldBindQuery(&form); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	data, err := h.appService.ListApp(form.Scope, form.ScopeId)
	if err != nil {
		return c.ResponseError(err)
	}
	return c.ResponseOK(data)
}
