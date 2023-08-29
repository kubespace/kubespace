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

type resolveChartHandler struct {
	models          *model.Models
	appStoreService *projectservice.AppStoreService
}

func ResolveChartHandler(conf *config.ServerConfig) api.Handler {
	return &resolveChartHandler{
		models:          conf.Models,
		appStoreService: conf.ServiceFactory.Project.AppStoreService,
	}
}

func (h *resolveChartHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, nil, nil
}

func (h *resolveChartHandler) Handle(c *api.Context) *utils.Response {
	file, err := c.FormFile("file")
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, "get chart file error: "+err.Error()))
	}
	chartIn, err := file.Open()
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, "get chart file error: "+err.Error()))
	}

	return h.appStoreService.ResolveChart(chartIn)
}
