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

type chartFilesHandler struct {
	models     *model.Models
	appService *projectservice.AppService
}

func ChartFilesHandler(conf *config.ServerConfig) api.Handler {
	return &chartFilesHandler{
		models:     conf.Models,
		appService: conf.ServiceFactory.Project.AppService,
	}
}

func (h *chartFilesHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	appVersionId, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	appVersion, err := h.models.AppVersionManager.GetById(appVersionId)
	if err != nil {
		return true, nil, errors.New(code.DataNotExists, err)
	}
	app, err := h.models.AppManager.GetById(appVersion.ScopeId)
	if err != nil {
		return true, nil, errors.New(code.DataNotExists, err)
	}
	return true, &api.AuthPerm{
		Scope:   app.Scope,
		ScopeId: app.ScopeId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *chartFilesHandler) Handle(c *api.Context) *utils.Response {
	appVersionId, _ := utils.ParseUint(c.Param("id"))
	app, appVersion, chartFiles, err := h.appService.GetAppChartFiles(appVersionId)
	if err != nil {
		return c.ResponseError(err)
	}
	data := map[string]interface{}{
		"chart_files": chartFiles,
		"app":         app,
		"app_version": appVersion,
	}
	return c.ResponseOK(data)
}
