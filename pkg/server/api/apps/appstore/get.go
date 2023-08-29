package appstore

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	projectservice "github.com/kubespace/kubespace/pkg/service/project"
	"github.com/kubespace/kubespace/pkg/utils"
)

type getHandler struct {
	models     *model.Models
	appStoreService *projectservice.AppStoreService
}

func GetHandler(conf *config.ServerConfig) api.Handler {
	return &getHandler{
		models:     conf.Models,
		appStoreService: conf.ServiceFactory.Project.AppStoreService,
	}
}

func (h *getHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, nil, nil
}

func (h *getHandler) Handle(c *api.Context) *utils.Response {
	var form getStoreAppForm
	if err := c.ShouldBindQuery(&form); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	storeAppId, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	return h.appStoreService.GetStoreApp(storeAppId, form.WithVersions)
}
