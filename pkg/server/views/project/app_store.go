package project

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/service/project"
	"github.com/kubespace/kubespace/pkg/utils"
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
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
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
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, "get chart file error: "+err.Error()))
	}
	chartIn, err := file.Open()
	if err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, "open chart file error: "+err.Error()))
	}

	var iconIn io.Reader
	icon, err := c.FormFile("icon")
	if err == nil {
		iconIn, _ = icon.Open()
	}
	app, appVersion, err := s.AppStoreService.CreateStoreApp(c.User, ser, chartIn, iconIn)
	resp := c.GenerateResponse(err, nil)
	if app == nil {
		return resp
	}
	var resId uint
	if appVersion != nil {
		resId = appVersion.ID
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationCreate,
		OperateDetail:        fmt.Sprintf("应用商店导入新应用版本：%s-%s", ser.Name, ser.PackageVersion),
		Scope:                types.ScopeAppStore,
		ScopeId:              app.ID,
		ScopeName:            app.Name,
		ResourceId:           resId,
		ResourceType:         types.AuditResourceAppVersion,
		ResourceName:         ser.Name + "-" + ser.PackageVersion,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: ser,
	})
	return resp
}

func (s *AppStore) deleteVersion(c *views.Context) *utils.Response {
	appId, err := strconv.ParseUint(c.Param("appId"), 10, 64)
	if err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	versionId, err := strconv.ParseUint(c.Param("versionId"), 10, 64)
	if err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	app, err := s.models.AppStoreManager.GetById(uint(appId))
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DataNotExists, "获取应用商店应用id=%d失败："+err.Error()))
	}
	appVersion, err := s.models.AppVersionManager.GetById(uint(versionId))
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DataNotExists, fmt.Sprintf("获取应用商店应用%s版本id=%d失败：%s", app.Name, versionId, err.Error())))
	}
	resp := s.AppStoreService.DeleteVersion(uint(appId), uint(versionId), c.User)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationDelete,
		OperateDetail:        fmt.Sprintf("应用商店删除应用版本：%s-%s", appVersion.PackageName, appVersion.PackageVersion),
		Scope:                types.ScopeAppStore,
		ScopeId:              app.ID,
		ScopeName:            app.Name,
		ResourceId:           appVersion.ID,
		ResourceType:         types.AuditResourceAppVersion,
		ResourceName:         appVersion.PackageName + "-" + appVersion.PackageVersion,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: nil,
	})
	return resp
}

func (s *AppStore) resolveChart(c *views.Context) *utils.Response {
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
