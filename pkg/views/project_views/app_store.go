package project_views

import (
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/project"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"io"
	"k8s.io/klog"
	"net/http"
	"strconv"
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
		views.NewView(http.MethodGet, "/:id", app.get),
		views.NewView(http.MethodPost, "/resolve", app.resolveChart),
		views.NewView(http.MethodPost, "/create", app.create),
		views.NewView(http.MethodDelete, "/:appId/:versionId", app.deleteVersion),
	}
	app.Views = vs
	return app
}

func (s *AppStore) list(c *views.Context) *utils.Response {
	var ser serializers.GetStoreAppSerializer
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return s.AppStoreService.ListStoreApp(&ser)
}

func (s *AppStore) get(c *views.Context) *utils.Response {
	var ser serializers.GetStoreAppSerializer
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	appId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return s.AppStoreService.GetStoreApp(uint(appId), ser)
}

func (s *AppStore) create(c *views.Context) *utils.Response {
	var ser serializers.AppStoreCreateSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	klog.Info(ser)

	file, err := c.FormFile("file")
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: "get chart file error: " + err.Error()}
	}
	chartIn, err := file.Open()
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: "get chart file error: " + err.Error()}
	}

	var iconIn io.Reader
	icon, err := c.FormFile("icon")
	if err == nil {
		iconIn, _ = icon.Open()
		klog.Info(icon)
	}
	return s.AppStoreService.CreateStoreApp(c.User, ser, chartIn, iconIn)
}

func (s *AppStore) deleteVersion(c *views.Context) *utils.Response {
	appId, err := strconv.ParseUint(c.Param("appId"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	versionId, err := strconv.ParseUint(c.Param("versionId"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return s.AppStoreService.DeleteVersion(uint(appId), uint(versionId), c.User)
}

func (s *AppStore) resolveChart(c *views.Context) *utils.Response {
	//a, _ := ioutil.ReadAll(c.Request)
	klog.Info(c.Request.Form)
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
