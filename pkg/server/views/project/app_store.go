package project

import (
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/service/project"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"io"
	"k8s.io/klog/v2"
	"net/http"
	"strconv"
)

type AppStore struct {
	Views           []*views.View
	AppStoreService *project.AppStoreService
	models          *model.Models
}

func NewAppStore(config *config.ServerConfig) *AppStore {
	app := &AppStore{
		AppStoreService: config.ServiceFactory.Project.AppStoreService,
		models:          config.Models,
	}
	app.Views = []*views.View{
		views.NewView(http.MethodGet, "", app.list),
		views.NewView(http.MethodGet, "/:id", app.get),
		views.NewView(http.MethodPost, "/resolve", app.resolveChart),
		views.NewView(http.MethodPost, "/create", app.create),
		views.NewView(http.MethodDelete, "/:appId/:versionId", app.deleteVersion),
	}
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
