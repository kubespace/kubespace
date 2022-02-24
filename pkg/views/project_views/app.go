package project_views

import (
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

func NewProjectApp(models *model.Models) *ProjectApp {
	app := &ProjectApp{
		AppService: project.NewAppService(models),
	}
	vs := []*views.View{
		//views.NewView(http.MethodGet, "", app.list),
		views.NewView(http.MethodPost, "", app.create),
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

func (a *ProjectApp) install(c *views.Context) *utils.Response {
	var ser serializers.ProjectInstallAppSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return a.AppService.InstallApp(c.User, ser)
}
