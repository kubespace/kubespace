package project_views

import (
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/project"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"net/http"
)

type ProjectApp struct {
	Views      []*views.View
	AppService *project.AppService
}

func NewProjectApp(kr *kube_resource.KubeResources, models *model.Models) *ProjectApp {
	app := &ProjectApp{
		AppService: project.NewAppService(kr, models),
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "", app.listApps),
		views.NewView(http.MethodGet, "/versions", app.listAppVersions),
		views.NewView(http.MethodGet, "/:id", app.getApp),
		views.NewView(http.MethodPost, "", app.create),
		views.NewView(http.MethodPost, "/install", app.install),
		views.NewView(http.MethodPost, "/destroy", app.destroy),
		//views.NewView(http.MethodDelete, "/:id", app.delete),
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

func (a *ProjectApp) getApp(c *views.Context) *utils.Response {
	//projectId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	var ser serializers.ProjectAppListSerializer
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return a.AppService.ListApp(ser)
}

func (a *ProjectApp) listAppVersions(c *views.Context) *utils.Response {
	var ser serializers.ProjectAppVersionListSerializer
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return a.AppService.ListAppVersions(ser)
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
