package project_views

import (
	"bytes"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/project"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"io"
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
		views.NewView(http.MethodGet, "/versions", app.listAppVersions),
		views.NewView(http.MethodPost, "/resolve", app.resolveChart),
		views.NewView(http.MethodPost, "/create", app.create),
	}
	app.Views = vs
	return app
}

func (s *AppStore) list(c *views.Context) *utils.Response {
	apps, err := s.models.AppStoreManager.ListStoreApps()
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: "获取应用商店列表失败:" + err.Error()}
	}
	return &utils.Response{Code: code.Success, Data: apps}
}

func (s *AppStore) listAppVersions(c *views.Context) *utils.Response {
	var ser serializers.ProjectAppVersionListSerializer
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return s.AppStoreService.ListAppVersions(ser)
}

func (s *AppStore) create(c *views.Context) *utils.Response {
	var ser serializers.AppStoreCreateSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}

	file, err := c.FormFile("file")
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: "get chart file error: " + err.Error()}
	}
	chartIn, err := file.Open()
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: "get chart file error: " + err.Error()}
	}
	return s.AppStoreService.CreateStoreApp(c.User, ser, chartIn)
}

func (s *AppStore) resolveChart(c *views.Context) *utils.Response {
	file, err := c.FormFile("file")
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: "get chart file error: " + err.Error()}
	}
	chartIn, err := file.Open()
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: "get chart file error: " + err.Error()}
	}
	buf := bytes.NewBuffer(nil)
	io.Copy(buf, chartIn)
	buf.Bytes()
	return s.AppStoreService.ResolveChart(chartIn)
}
