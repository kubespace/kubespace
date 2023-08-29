package version

import (
	"bytes"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	projectservice "github.com/kubespace/kubespace/pkg/service/project"
	"github.com/kubespace/kubespace/pkg/utils"
	"helm.sh/helm/v3/pkg/chart/loader"
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

func (h *getHandler) Handle(c *api.Context) *utils.Response {
	appVersionId, _ := utils.ParseUint(c.Param("id"))
	appVersion, _ := h.models.AppVersionManager.GetById(appVersionId)
	app, _ := h.models.AppManager.GetAppWithVersion(appVersion.ScopeId)
	appCharts, err := h.models.AppVersionManager.GetChart(appVersion.ChartPath)
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}
	charts, err := loader.LoadArchive(bytes.NewReader(appCharts.Content))
	if err != nil {
		return c.ResponseError(errors.New(code.GetError, err))
	}
	res := map[string]interface{}{
		"id":              appVersion.ID,
		"name":            app.Name,
		"description":     app.Description,
		"package_name":    appVersion.PackageName,
		"package_version": appVersion.PackageVersion,
		"type":            app.Type,
		"values":          appVersion.Values,
		"templates":       charts.Templates,
	}
	return c.ResponseOK(res)
}
