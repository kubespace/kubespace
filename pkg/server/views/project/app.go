package project

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/service/project"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ProjectApp struct {
	Views      []*views.View
	AppService *project.AppService
	models     *model.Models
}

func NewProjectApp(config *config.ServerConfig) *ProjectApp {
	app := &ProjectApp{
		AppService: config.ServiceFactory.Project.AppService,
		models:     config.Models,
	}
	app.Views = []*views.View{
		views.NewView(http.MethodGet, "", app.listApps),
		views.NewView(http.MethodGet, "/versions", app.listAppVersions),
		views.NewView(http.MethodGet, "/version/:id", app.getAppVersion),
		views.NewView(http.MethodGet, "/status", app.listAppStatus),
		views.NewView(http.MethodGet, "/status_sse", app.statusSSE),
		views.NewView(http.MethodGet, "/download", app.downloadChart),
		views.NewView(http.MethodGet, "/:id", app.getApp),
		views.NewView(http.MethodGet, "/version/:id/chartfiles", app.getChartFiles),
		views.NewView(http.MethodPost, "", app.create),
		views.NewView(http.MethodPost, "/install", app.install),
		views.NewView(http.MethodPost, "/destroy", app.destroy),
		views.NewView(http.MethodPost, "/import_storeapp", app.importStoreapp),
		views.NewView(http.MethodPost, "/import_custom_app", app.importCustomApp),
		views.NewView(http.MethodPost, "/duplicate_app", app.duplicateApp),
		views.NewView(http.MethodDelete, "/version/:id", app.deleteAppVersion),
		views.NewView(http.MethodDelete, "/:id", app.deleteApp),
	}
	return app
}

func (a *ProjectApp) create(c *views.Context) *utils.Response {
	var ser serializers.CreateAppSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	app, resp := a.AppService.CreateProjectApp(c.User, ser)
	var opScope, scopeName, opDetail, opNamespace string
	if ser.Scope == types.AppVersionScopeProjectApp {
		opDetail = fmt.Sprintf("应用%s创建新版本：%s-%s", app.Name, ser.Name, ser.Version)
		projectObj, err := a.models.ProjectManager.Get(uint(ser.ScopeId))
		if err != nil {
			return &utils.Response{Code: code.GetError, Msg: err.Error()}
		}
		scopeName = projectObj.Name
		opScope = types.ScopeProject
		opNamespace = projectObj.Namespace
	}
	ser.ChartFiles = nil
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationCreate,
		OperateDetail:        opDetail,
		Scope:                opScope,
		ScopeId:              ser.ScopeId,
		ScopeName:            scopeName,
		Namespace:            opNamespace,
		ResourceId:           app.ID,
		ResourceType:         types.AuditResourceAppVersion,
		ResourceName:         ser.Name + "-" + ser.Version,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: ser,
	})
	return resp
}

func (a *ProjectApp) listApps(c *views.Context) *utils.Response {
	var ser serializers.AppListSerializer
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	data, err := a.AppService.ListApp(ser.Scope, ser.ScopeId)
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success, Data: data}
}

func (a *ProjectApp) listAppStatus(c *views.Context) *utils.Response {
	var ser serializers.AppListSerializer
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
	app, err := a.models.AppManager.Get(uint(appId))
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if app == nil {
		return &utils.Response{Code: code.GetError, Msg: fmt.Sprintf("获取应用失败：未找到该应用id=%d", appId)}
	}

	var opDetail string
	var opScopeName, opScope, opNamespace string
	var resType = types.AuditResourceApp
	if app.Scope == types.AppVersionScopeProjectApp {
		if projectObj, err := a.models.ProjectManager.Get(app.ScopeId); err != nil {
			return &utils.Response{Code: code.DBError, Msg: fmt.Sprintf("获取工作空间id=%d失败：%s", app.ScopeId, err.Error())}
		} else {
			opScope = types.ScopeProject
			opScopeName = projectObj.Name
			opNamespace = projectObj.Namespace
		}
		opDetail = fmt.Sprintf("删除应用%s，以及所有版本", app.Name)
	} else if app.Scope == types.AppVersionScopeComponent {
		if clusterObj, err := a.models.ClusterManager.Get(app.ScopeId); err != nil {
			return &utils.Response{Code: code.DBError, Msg: fmt.Sprintf("获取集群id=%d失败：%s", app.ScopeId, err.Error())}
		} else {
			opScope = types.ScopeCluster
			opScopeName = clusterObj.Name1
		}
		opNamespace = app.Namespace
		opDetail = fmt.Sprintf("删除集群组件%s", app.Name)
		resType = types.AuditResourceClusterComponent
	}
	err = a.models.AppManager.DeleteApp(uint(appId))
	resp := &utils.Response{Code: code.Success}
	if err != nil {
		resp = &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationDelete,
		OperateDetail:        opDetail,
		Scope:                opScope,
		ScopeId:              app.ScopeId,
		ScopeName:            opScopeName,
		Namespace:            opNamespace,
		ResourceId:           app.ID,
		ResourceType:         resType,
		ResourceName:         app.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: nil,
	})
	return &utils.Response{Code: code.Success}
}

func (a *ProjectApp) listAppVersions(c *views.Context) *utils.Response {
	var ser serializers.AppVersionListSerializer
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

func (a *ProjectApp) deleteAppVersion(c *views.Context) *utils.Response {
	appVersionId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	appVersion, err := a.models.AppVersionManager.GetById(uint(appVersionId))
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: fmt.Sprintf("获取应用版本失败：%s", err.Error())}
	}
	var opScope, opScopeName, opDetail, opNamespace string
	var opScopeId uint
	if appVersion.Scope == types.AppVersionScopeProjectApp {
		appObj, err := a.models.AppManager.Get(appVersion.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.GetError, Msg: fmt.Sprintf("获取应用失败：%s", err.Error())}
		}
		projectObj, err := a.models.ProjectManager.Get(appObj.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.GetError, Msg: fmt.Sprintf("获取应用所在工作空间失败：%s", err.Error())}
		}
		opScope = types.ScopeProject
		opScopeId = projectObj.ID
		opScopeName = projectObj.Name
		opNamespace = projectObj.Namespace
		opDetail = fmt.Sprintf("删除应用%s所属版本：%s-%s", appObj.Name, appVersion.PackageName, appVersion.PackageVersion)
	} else if appVersion.Scope == types.AppVersionScopeStoreApp {
		appStoreObj, err := a.models.AppStoreManager.GetById(appVersion.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.GetError, Msg: fmt.Sprintf("获取应用商店应用失败：%s", err.Error())}
		}
		opScope = types.ScopeAppStore
		opScopeId = appStoreObj.ID
		opScopeName = appStoreObj.Name
		opDetail = fmt.Sprintf("应用商店删除应用版本：%s-%s", appVersion.PackageName, appVersion.PackageVersion)
	} else if appVersion.Scope == types.AppVersionScopeComponent {
		appObj, err := a.models.AppManager.Get(appVersion.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.GetError, Msg: fmt.Sprintf("获取应用id=%d失败：%s", appVersion.ScopeId, err.Error())}
		}
		clusterObj, err := a.models.ClusterManager.Get(appObj.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: fmt.Sprintf("获取集群id=%d失败：%s", appObj.ScopeId, err.Error())}
		}
		opScope = types.ScopeCluster
		opScopeId = clusterObj.ID
		opNamespace = appObj.Namespace
		opScopeName = clusterObj.Name1

		opDetail = fmt.Sprintf("删除集群组件版本：%s-%s", appVersion.PackageName, appVersion.PackageVersion)
	}
	resp := &utils.Response{Code: code.Success}
	if err = a.models.AppVersionManager.Delete(uint(appVersionId)); err != nil {
		resp = &utils.Response{Code: code.DBError, Msg: "删除应用版本失败：" + err.Error()}
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationDelete,
		OperateDetail:        opDetail,
		Scope:                opScope,
		ScopeId:              opScopeId,
		ScopeName:            opScopeName,
		Namespace:            opNamespace,
		ResourceId:           appVersion.ID,
		ResourceType:         types.AuditResourceAppVersion,
		ResourceName:         appVersion.PackageName + "-" + appVersion.PackageVersion,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: nil,
	})
	return resp
}

func (a *ProjectApp) install(c *views.Context) *utils.Response {
	var ser serializers.InstallAppSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	versionApp, err := a.models.AppVersionManager.GetById(ser.AppVersionId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	app, err := a.models.AppManager.Get(ser.AppId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	var opDetail, opScope, opScopeName, opNamespace, opResType string
	var opScopeId uint
	opStr := "安装"
	operation := types.AuditOperationInstall
	if ser.Upgrade {
		opStr = "升级"
		operation = types.AuditOperationUpgrade
	}
	if app.Scope == types.AppVersionScopeProjectApp {
		projectObj, err := a.models.ProjectManager.Get(app.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.GetError, Msg: fmt.Sprintf("获取应用所在工作空间失败：%s", err.Error())}
		}
		opScope = types.ScopeProject
		opScopeId = projectObj.ID
		opNamespace = projectObj.Namespace
		opScopeName = projectObj.Name
		opResType = types.AuditResourceApp
		opDetail = fmt.Sprintf("%s应用%s版本：%s-%s", opStr, app.Name, versionApp.PackageName, versionApp.PackageVersion)
	} else {
		clusterObj, err := a.models.ClusterManager.Get(app.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: fmt.Sprintf("获取集群id=%d失败：%s", app.ScopeId, err.Error())}
		}
		opScope = types.ScopeCluster
		opScopeId = clusterObj.ID
		opNamespace = app.Namespace
		opScopeName = clusterObj.Name1
		opResType = types.AuditResourceClusterComponent
		opDetail = fmt.Sprintf("%s集群组件：%s-%s", opStr, versionApp.PackageName, versionApp.PackageVersion)
	}
	resp := a.AppService.InstallApp(c.User, ser)
	c.CreateAudit(&types.AuditOperate{
		Operation:            operation,
		OperateDetail:        opDetail,
		Scope:                opScope,
		ScopeId:              opScopeId,
		ScopeName:            opScopeName,
		Namespace:            opNamespace,
		ResourceId:           app.ID,
		ResourceType:         opResType,
		ResourceName:         app.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: nil,
	})
	return resp
}

func (a *ProjectApp) destroy(c *views.Context) *utils.Response {
	var ser serializers.InstallAppSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	app, err := a.models.AppManager.Get(ser.AppId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	var opDetail, opScope, opScopeName, opNamespace, opResType string
	var opScopeId uint
	if app.Scope == types.AppVersionScopeProjectApp {
		projectObj, err := a.models.ProjectManager.Get(app.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.GetError, Msg: fmt.Sprintf("获取应用所在工作空间失败：%s", err.Error())}
		}
		opScope = types.ScopeProject
		opScopeId = projectObj.ID
		opNamespace = projectObj.Namespace
		opScopeName = projectObj.Name
		opResType = types.AuditResourceApp
		opDetail = fmt.Sprintf("销毁应用：%s", app.Name)
	} else {
		clusterObj, err := a.models.ClusterManager.Get(app.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: fmt.Sprintf("获取集群id=%d失败：%s", app.ScopeId, err.Error())}
		}
		opScope = types.ScopeCluster
		opScopeId = clusterObj.ID
		opNamespace = app.Namespace
		opScopeName = clusterObj.Name1
		opResType = types.AuditResourceClusterComponent
		opDetail = fmt.Sprintf("销毁集群组件：%s", app.Name)
	}
	resp := a.AppService.DestroyApp(c.User, ser)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationDestroy,
		OperateDetail:        opDetail,
		Scope:                opScope,
		ScopeId:              opScopeId,
		ScopeName:            opScopeName,
		Namespace:            opNamespace,
		ResourceId:           app.ID,
		ResourceType:         opResType,
		ResourceName:         app.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: nil,
	})
	return resp
}

func (a *ProjectApp) importStoreapp(c *views.Context) *utils.Response {
	var ser serializers.ImportStoreAppSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	app, appVersion, resp := a.AppService.ImportStoreApp(ser, c.User)
	if app == nil {
		return resp
	}
	var opDetail, opScope, opScopeName, opNamespace, opResType string
	var opScopeId uint
	if app.Scope == types.AppVersionScopeProjectApp {
		projectObj, err := a.models.ProjectManager.Get(app.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.GetError, Msg: fmt.Sprintf("获取应用所在工作空间失败：%s", err.Error())}
		}
		opScope = types.ScopeProject
		opScopeId = projectObj.ID
		opNamespace = projectObj.Namespace
		opScopeName = projectObj.Name
		opResType = types.AuditResourceApp
		opDetail = fmt.Sprintf("从应用商店导入应用：%s-%s", appVersion.PackageName, appVersion.PackageVersion)
	} else {
		clusterObj, err := a.models.ClusterManager.Get(app.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: fmt.Sprintf("获取集群id=%d失败：%s", app.ScopeId, err.Error())}
		}
		opScope = types.ScopeCluster
		opScopeId = clusterObj.ID
		opNamespace = app.Namespace
		opScopeName = clusterObj.Name1
		opResType = types.AuditResourceClusterComponent
		opDetail = fmt.Sprintf("从应用商店导入集群组件：%s-%s", appVersion.PackageName, appVersion.PackageVersion)
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationImport,
		OperateDetail:        opDetail,
		Scope:                opScope,
		ScopeId:              opScopeId,
		ScopeName:            opScopeName,
		Namespace:            opNamespace,
		ResourceId:           app.ID,
		ResourceType:         opResType,
		ResourceName:         app.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: nil,
	})
	return resp
}

func (a *ProjectApp) importCustomApp(c *views.Context) *utils.Response {
	var ser serializers.ImportCustomAppSerializer
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
	app, _, resp := a.AppService.ImportCustomApp(c.User, ser, chartIn)
	if app == nil {
		return resp
	}
	var opDetail, opScope, opScopeName, opNamespace, opResType string
	var opScopeId uint
	if app.Scope == types.AppVersionScopeProjectApp {
		projectObj, err := a.models.ProjectManager.Get(app.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.GetError, Msg: fmt.Sprintf("获取应用所在工作空间失败：%s", err.Error())}
		}
		opScope = types.ScopeProject
		opScopeId = projectObj.ID
		opNamespace = projectObj.Namespace
		opScopeName = projectObj.Name
		opResType = types.AuditResourceApp
		opDetail = fmt.Sprintf("导入自定义应用：%s-%s", ser.Name, ser.PackageVersion)
	} else {
		clusterObj, err := a.models.ClusterManager.Get(app.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: fmt.Sprintf("获取集群id=%d失败：%s", app.ScopeId, err.Error())}
		}
		opScope = types.ScopeCluster
		opScopeId = clusterObj.ID
		opNamespace = app.Namespace
		opScopeName = clusterObj.Name1
		opResType = types.AuditResourceClusterComponent
		opDetail = fmt.Sprintf("导入自定义集群组件：%s-%s", ser.Name, ser.PackageVersion)
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationImport,
		OperateDetail:        opDetail,
		Scope:                opScope,
		ScopeId:              opScopeId,
		ScopeName:            opScopeName,
		Namespace:            opNamespace,
		ResourceId:           app.ID,
		ResourceType:         opResType,
		ResourceName:         app.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: nil,
	})
	return resp
}

func (a *ProjectApp) duplicateApp(c *views.Context) *utils.Response {
	var ser serializers.DuplicateAppSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	app, appVersion, resp := a.AppService.DuplicateApp(&ser, c.User)
	if app == nil {
		return resp
	}
	var opDetail, operation string
	projectObj, err := a.models.ProjectManager.Get(app.ScopeId)
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: fmt.Sprintf("获取应用所在工作空间失败：%s", err.Error())}
	}
	if ser.Scope == types.AppVersionScopeProjectApp {
		destProject, err := a.models.ProjectManager.Get(ser.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.GetError, Msg: fmt.Sprintf("获取目标工作空间失败：%s", err.Error())}
		}
		operation = types.AuditOperationClone
		opDetail = fmt.Sprintf("克隆应用版本：%s-%s到目标工作空间：%s", appVersion.PackageName, appVersion.PackageVersion, destProject.Name)
	} else {
		operation = types.AuditOperationRelease
		opDetail = fmt.Sprintf("发布应用版本：%s-%s到应用商店", appVersion.PackageName, appVersion.PackageVersion)
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            operation,
		OperateDetail:        opDetail,
		Scope:                types.ScopeProject,
		ScopeId:              projectObj.ID,
		ScopeName:            projectObj.Name,
		Namespace:            projectObj.Namespace,
		ResourceId:           app.ID,
		ResourceType:         types.AuditResourceApp,
		ResourceName:         app.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: nil,
	})
	return resp
}

func (a *ProjectApp) statusSSE(c *views.Context) *utils.Response {
	var ser serializers.AppListSerializer
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	w := c.Writer
	clientGone := w.CloseNotify()
	c.SSEvent("message", "{}")
	w.Flush()
	//c.Stream()
	tick := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-clientGone:
			klog.Info("app status client gone")
			return nil
		case <-tick.C:
			tick.Stop()
			res := a.AppService.ListAppStatus(ser)
			c.SSEvent("message", res)
			w.Flush()
			tick.Reset(5 * time.Second)
		}
	}
}

func (a *ProjectApp) downloadChart(c *views.Context) *utils.Response {
	chartPath := c.Query("path")
	if chartPath == "" {
		c.JSON(http.StatusNotFound, &utils.Response{Code: code.GetError, Msg: "not found chart params"})
		return nil
	}

	appChart, err := a.models.AppVersionManager.GetChart(chartPath)
	if err != nil {
		c.JSON(http.StatusNotFound, &utils.Response{Code: code.GetError, Msg: err.Error()})
		return nil
	}
	chartName := chartPath
	s := strings.Split(chartPath, "/")
	if len(s) >= 2 {
		chartName = s[len(s)-1]
	}
	fileContentDisposition := "attachment;filename=\"" + chartName + "\""
	c.Header("Content-Disposition", fileContentDisposition)
	c.Data(http.StatusOK, "application/x-tar", appChart.Content)
	return nil
}

func (a *ProjectApp) getChartFiles(c *views.Context) *utils.Response {
	appVersionId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	app, appVersion, chartFiles, err := a.AppService.GetAppChartFiles(uint(appVersionId))
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: err.Error()}
	}
	data := map[string]interface{}{
		"chart_files": chartFiles,
		"app":         app,
		"app_version": appVersion,
	}
	return &utils.Response{Code: code.Success, Data: data}
}
