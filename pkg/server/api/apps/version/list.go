package version

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

type listAppVersionsForm struct {
	Scope   string `json:"scope" form:"scope"`
	ScopeId uint   `json:"scope_id" form:"scope_id"`
}

func (h *listHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	var form listAppVersionsForm
	if err := c.ShouldBindQuery(&form); err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	if form.Scope == types.ScopeAppStore {
		return true, nil, nil
	}
	app, err := h.models.AppManager.GetById(form.ScopeId)
	if err != nil {
		return true, nil, errors.New(code.DataNotExists, err)
	}
	return true, &api.AuthPerm{
		Scope:   app.Scope,
		ScopeId: app.ScopeId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *listHandler) Handle(c *api.Context) *utils.Response {
	var form listAppVersionsForm
	if err := c.ShouldBindQuery(&form); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	appVersions, err := h.models.AppVersionManager.List(form.Scope, form.ScopeId)
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}
	return c.ResponseOK(appVersions)
}
