package project_views

import (
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/project"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"net/http"
	"strconv"
)

type ProjectApp struct {
	Views      []*views.View
	AppService *project.AppService
	models     *model.Models
}

func NewProjectApp(models *model.Models, appService *project.AppService) *ProjectApp {
	app := &ProjectApp{
		AppService: appService,
		models:     models,
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "", app.listApps),
		views.NewView(http.MethodGet, "/versions", app.listAppVersions),
		views.NewView(http.MethodGet, "/version/:id", app.getAppVersion),
		views.NewView(http.MethodGet, "/status", app.listAppStatus),
		views.NewView(http.MethodGet, "/:id", app.getApp),
		views.NewView(http.MethodPost, "", app.create),
		views.NewView(http.MethodPost, "/install", app.install),
		views.NewView(http.MethodPost, "/destroy", app.destroy),
		views.NewView(http.MethodPost, "/import_storeapp", app.importStoreapp),
		views.NewView(http.MethodDelete, "/:id", app.deleteApp),
	}
	app.Views = vs
	return app
}

func (a *ProjectApp) create(c *views.Context) *utils.Response {
	var ser serializers.ProjectCreateAppSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return a.AppService.CreateProjectApp(c.User, ser)
}

func (a *ProjectApp) listApps(c *views.Context) *utils.Response {
	var ser serializers.ProjectAppListSerializer
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return a.AppService.ListApp(ser)
}

func (a *ProjectApp) listAppStatus(c *views.Context) *utils.Response {
	var ser serializers.ProjectAppListSerializer
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return a.AppService.ListAppStatus(ser)
}

func (a *ProjectApp) getApp(c *views.Context) *utils.Response {
	appId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return a.AppService.GetApp(uint(appId))
}

func (a *ProjectApp) deleteApp(c *views.Context) *utils.Response {
	appId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	err = a.models.ProjectAppManager.DeleteProjectApp(uint(appId))
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success}
}

func (a *ProjectApp) listAppVersions(c *views.Context) *utils.Response {
	var ser serializers.ProjectAppVersionListSerializer
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return a.AppService.ListAppVersions(ser)
}

func (a *ProjectApp) getAppVersion(c *views.Context) *utils.Response {
	appVersionId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return a.AppService.GetAppVersion(uint(appVersionId))
}

func (a *ProjectApp) install(c *views.Context) *utils.Response {
	var ser serializers.ProjectInstallAppSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return a.AppService.InstallApp(c.User, ser)
}

func (a *ProjectApp) destroy(c *views.Context) *utils.Response {
	var ser serializers.ProjectInstallAppSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return a.AppService.DestroyApp(c.User, ser)
}

func (a *ProjectApp) importStoreapp(c *views.Context) *utils.Response {
	var ser serializers.ImportStoreAppSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return a.AppService.ImportStoreApp(ser, c.User)
}
