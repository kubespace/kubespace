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

type getHandler struct {
	models     *model.Models
	appService *projectservice.AppService
}

func GetHandler(conf *config.ServerConfig) api.Handler {
	return &getHandler{
		models:     conf.Models,
		appService: conf.ServiceFactory.Project.AppService,
	}
}

func (h *getHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	appId, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	app, err := h.models.AppManager.GetById(appId)
	if err != nil {
		return true, nil, errors.New(code.DataNotExists, err)
	}
	return true, &api.AuthPerm{
		Scope:   app.Scope,
		ScopeId: app.ScopeId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *getHandler) Handle(c *api.Context) *utils.Response {
	appId, _ := utils.ParseUint(c.Param("id"))
	return h.appService.GetApp(appId)
}
