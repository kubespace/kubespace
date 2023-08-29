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

type listHandler struct {
	models          *model.Models
	appStoreService *projectservice.AppStoreService
}

func ListHandler(conf *config.ServerConfig) api.Handler {
	return &listHandler{
		models:          conf.Models,
		appStoreService: conf.ServiceFactory.Project.AppStoreService,
	}
}

type getStoreAppForm struct {
	WithVersions bool `json:"with_versions" form:"with_versions"`
}

func (h *listHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, nil, nil
}

func (h *listHandler) Handle(c *api.Context) *utils.Response {
	var form getStoreAppForm
	if err := c.ShouldBindQuery(&form); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	return h.appStoreService.ListStoreApp(form.WithVersions)
}
