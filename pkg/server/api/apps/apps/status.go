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

type statusHandler struct {
	models     *model.Models
	appService *projectservice.AppService
}

func StatusHandler(conf *config.ServerConfig) api.Handler {
	return &statusHandler{
		models:     conf.Models,
		appService: conf.ServiceFactory.Project.AppService,
	}
}

type appStatusForm struct {
	Scope         string `json:"scope" form:"scope"`
	ScopeId       uint   `json:"scope_id" form:"scope_id"`
	Name          string `json:"name" form:"name"`
}

func (h *statusHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	var form appStatusForm
	if err := c.ShouldBindQuery(&form); err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   form.Scope,
		ScopeId: form.ScopeId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *statusHandler) Handle(c *api.Context) *utils.Response {
	var form appStatusForm
	if err := c.ShouldBindQuery(&form); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	status, err := h.appService.GetAppStatus(form.Scope, form.ScopeId, form.Name)
	return c.Response(err, status)
}
