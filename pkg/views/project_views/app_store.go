package project_views

import (
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/project"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views"
	"net/http"
)

type AppStore struct {
	Views           []*views.View
	AppStoreService *project.AppStoreService
	models          *model.Models
}

func NewAppStore(models *model.Models, appStoreService *project.AppStoreService) *AppStore {
	app := &AppStore{
		AppStoreService: appStoreService,
		models:          models,
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "", app.list),
	}
	app.Views = vs
	return app
}

func (s *AppStore) list(c *views.Context) *utils.Response {
	file, err := c.FormFile("file")
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: "get chart file error: " + err.Error()}
	}
	chartIn, err := file.Open()
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: "get chart file error: " + err.Error()}
	}
	return s.AppStoreService.ResolveChart(chartIn)
}
