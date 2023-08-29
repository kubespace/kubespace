package project

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	kubetypes "github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/model/manager/project"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/service/cluster"
	"github.com/kubespace/kubespace/pkg/third/helm"
	"github.com/kubespace/kubespace/pkg/utils"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog/v2"
	"os"
	"sigs.k8s.io/yaml"
	"strings"
	"time"
)

type AppService struct {
	*AppBaseService
	kubeClient *cluster.KubeClient
}

func NewAppService(kubeClient *cluster.KubeClient, appBaseService *AppBaseService) *AppService {
	return &AppService{
		kubeClient:     kubeClient,
		AppBaseService: appBaseService,
	}
}

func (a *AppService) WriteFile(fileName, content string) error {
	f, err := os.Create(fileName)
	defer f.Close()
	if err != nil {
		return err
	}
	if _, err = f.Write([]byte(content)); err != nil {
		return err
	}
	return nil
}

type CreateAppForm struct {
	Scope              string                 `json:"scope" form:"scope"`
	ScopeId            uint                   `json:"scope_id" form:"scope_id"`
	Name               string                 `json:"name" form:"name"`
	From               string                 `json:"from" form:"from"`
	Type               string                 `json:"type" form:"type"`
	Description        string                 `json:"description" form:"description"`
	VersionDescription string                 `json:"version_description" form:"version_description"`
	Version            string                 `json:"version" form:"version"`
	Values             string                 `json:"values" form:"values"`
	ChartFiles         map[string]interface{} `json:"chart_files"`
	User               string                 `json:"user"`
}

func (a *AppService) CreateProjectApp(form *CreateAppForm) (*types.App, error) {
	app, err := a.models.AppManager.GetByName(form.Scope, form.ScopeId, form.Name)
	if err != nil {
		return nil, errors.New(code.DBError, err)
	}
	if app != nil {
		sameVersion, err := a.models.AppVersionManager.GetByPackageNameVersion(form.Scope, app.ID, form.Name, form.Version)
		if err != nil {
			return nil, errors.New(code.DBError, err)
		}
		if sameVersion != nil {
			return nil, errors.New(code.ParamsError, "该应用已存在相同版本，请重新输入版本号")
		}
		app.UpdateUser = form.User
		app.UpdateTime = time.Now()
		app.Description = form.Description
	} else {
		if form.Type != types.AppTypeOrdinaryApp && form.Type != types.AppTypeMiddleware {
			return nil, errors.New(code.ParamsError, "应用类型参数错误")
		}
		app = &types.App{
			Scope:       form.Scope,
			ScopeId:     form.ScopeId,
			Name:        form.Name,
			Status:      types.AppStatusUninstall,
			Type:        form.Type,
			Description: form.Description,
			CreateUser:  form.User,
			UpdateUser:  form.User,
			CreateTime:  time.Now(),
			UpdateTime:  time.Now(),
		}
	}
	chartGen := &helm.ChartGeneration{
		NeedModifyVersion: false,
		PackageVersion:    form.Version,
		AppVersion:        form.Version,
		Files:             form.ChartFiles,
		Base64Encoded:     false,
	}
	if form.From == types.AppVersionFromImport {
		chartGen.NeedModifyVersion = true
		chartGen.Base64Encoded = true
	}
	chartDir, chartPath, err := chartGen.GenerateChart()
	if chartDir != "" {
		defer os.RemoveAll(chartDir)
	}
	if err != nil {
		return nil, err
	}
	appVersion := &types.AppVersion{
		PackageName:    form.Name,
		PackageVersion: form.Version,
		AppVersion:     form.Version,
		Values:         form.Values,
		Description:    form.VersionDescription,
		From:           form.From,
		CreateUser:     form.User,
		CreateTime:     time.Now(),
		UpdateTime:     time.Now(),
	}
	_, err = a.models.AppManager.CreateApp(chartPath, app, appVersion)
	if err != nil {
		return nil, errors.New(code.DBError, err)
	}
	return app, nil
}

type InstallAppForm struct {
	AppId        uint   `json:"app_id" form:"app_id"`
	Values       string `json:"values" form:"values"`
	AppVersionId uint   `json:"app_version_id" form:"app_version_id"`
	Upgrade      bool   `json:"upgrade" form:"upgrade"`
	User         string `json:"user" form:"user"`
}

func (a *AppService) InstallApp(installForm *InstallAppForm) (*types.App, *types.AppVersion, error) {
	versionApp, err := a.models.AppVersionManager.GetById(installForm.AppVersionId)
	if err != nil {
		return nil, nil, errors.New(code.DataNotExists, "get app version error: "+err.Error())
	}
	if versionApp.ScopeId != installForm.AppId {
		return nil, nil, errors.New(code.ParamsError, "当前应用不存在该版本，请重新选择")
	}
	app, err := a.models.AppManager.GetById(installForm.AppId)
	if err != nil {
		return nil, nil, errors.New(code.DataNotExists, "get app error: "+err.Error())
	}
	var clusterId string
	var namespace string
	if app.Scope == types.ScopeProject {
		projectObj, err := a.models.ProjectManager.Get(app.ScopeId)
		if err != nil {
			return app, versionApp, errors.New(code.DataNotExists, "get project error: "+err.Error())
		}
		clusterId = projectObj.ClusterId
		namespace = projectObj.Namespace
	} else {
		clusterId = fmt.Sprintf("%d", app.ScopeId)
		namespace = app.Namespace
	}
	appChart, err := a.models.AppVersionManager.GetChart(versionApp.ChartPath)
	if err != nil {
		return app, versionApp, errors.New(code.DataNotExists, "not found chart path "+versionApp.ChartPath)
	}
	installParams := map[string]interface{}{
		"name":        app.Name,
		"namespace":   namespace,
		"chart_bytes": appChart.Content,
		"values":      installForm.Values,
	}
	var resp *utils.Response
	if installForm.Upgrade {
		resp = a.kubeClient.Update(clusterId, kubetypes.HelmType, installParams)
	} else {
		resp = a.kubeClient.Create(clusterId, kubetypes.HelmType, installParams)
	}
	if !resp.IsSuccess() {
		return app, versionApp, errors.New(resp.Code, resp.Msg)
	}
	app.AppVersionId = installForm.AppVersionId
	app.UpdateUser = installForm.User
	app.Status = types.AppStatusNotReady
	if err = a.models.AppManager.UpdateApp(app, "status", "app_version_id", "update_user", "update_time"); err != nil {
		return app, versionApp, errors.New(code.DBError, err)
	}
	versionApp.Values = installForm.Values
	if err = a.models.AppVersionManager.UpdateAppVersion(versionApp, "values"); err != nil {
		return app, versionApp, errors.New(code.DBError, err)
	}
	if _, err = a.models.AppManager.CreateRevision(versionApp, app); err != nil {
		klog.Errorf("create project app id=%s, name=%s revision error: %s", app.ID, app.Name, err)
	}
	return app, versionApp, nil
}

func (a *AppService) DestroyApp(appId uint, user string) (*types.App, error) {
	app, err := a.models.AppManager.GetAppWithVersion(appId)
	if err != nil {
		return nil, errors.New(code.DataNotExists, err)
	}
	var clusterId string
	var namespace string
	if app.Scope == types.ScopeProject {
		projectObj, err := a.models.ProjectManager.Get(app.ScopeId)
		if err != nil {
			return app, errors.New(code.DataNotExists, "get project error: "+err.Error())
		}
		clusterId = projectObj.ClusterId
		namespace = projectObj.Namespace
	} else {
		clusterId = fmt.Sprintf("%d", app.ScopeId)
		namespace = app.Namespace
	}
	destroyParams := map[string]interface{}{
		"namespace": namespace,
		"name":      app.Name,
	}
	resp := a.kubeClient.Delete(clusterId, kubetypes.HelmType, destroyParams)
	if !resp.IsSuccess() {
		return app, errors.New(resp.Code, resp.Msg)
	}
	app.UpdateUser = user
	app.Status = types.AppStatusUninstall
	if err = a.models.AppManager.UpdateApp(app, "status", "update_user", "update_time"); err != nil {
		return app, errors.New(code.DBError, err)
	}
	return app, nil
}

type AppRuntimeStatus struct {
	Name          string                       `json:"name"`
	RuntimeStatus string                       `json:"runtime_status"`
	Objects       []*unstructured.Unstructured `json:"objects"`
	PodsNum       int                          `json:"pods_num"`
	ReadyPodsNum  int                          `json:"ready_pods_num"`
}

func (a *AppService) updateAppStatus(scope string, scopeId uint, apps []*types.App) (map[string]*AppRuntimeStatus, error) {
	var clusterId string
	var namespaceApps = map[string][]string{}
	if scope == types.ScopeProject {
		projectObj, err := a.models.ProjectManager.Get(scopeId)
		if err != nil {
			return nil, errors.New(code.DataNotExists, err)
		}
		clusterId = projectObj.ClusterId
		for _, app := range apps {
			namespaceApps[projectObj.Namespace] = append(namespaceApps[projectObj.Namespace], app.Name)
		}
	} else {
		clusterId = fmt.Sprintf("%d", scopeId)
		for _, app := range apps {
			namespaceApps[app.Namespace] = append(namespaceApps[app.Namespace], app.Name)
		}
	}

	nameStatusMap := map[string]*AppRuntimeStatus{}
	for namespace, appNames := range namespaceApps {
		statusParams := map[string]interface{}{
			"namespace":   namespace,
			"names":       appNames,
			"with_status": true,
		}
		res := a.kubeClient.List(clusterId, kubetypes.HelmType, statusParams)
		if res.IsSuccess() {
			var appStatuses []*AppRuntimeStatus
			if err := utils.ConvertTypeByJson(res.Data, &appStatuses); err != nil {
				klog.Error("unmarshal app status error: ", err.Error())
			}
			for _, status := range appStatuses {
				nameStatusMap[status.Name] = status
			}
		} else {
			klog.Errorf("get app status error: %+v", res)
			return nil, nil
		}
	}
	return nameStatusMap, nil
}

func (a *AppService) ListApp(scope string, scopeId uint) ([]*types.App, error) {
	apps, err := a.models.AppManager.ListApps(project.ListAppCondition{
		Scope:       scope,
		ScopeId:     scopeId,
		WithVersion: true,
	})
	if err != nil {
		return nil, errors.New(code.DBError, err)
	}
	nameStatusMap, err := a.updateAppStatus(scope, scopeId, apps)
	if err != nil {
		return nil, err
	}
	if nameStatusMap != nil {
		for idx, app := range apps {
			if _, ok := nameStatusMap[app.Name]; ok {
				apps[idx].Status = nameStatusMap[app.Name].RuntimeStatus
				apps[idx].PodsNum = nameStatusMap[app.Name].PodsNum
				apps[idx].ReadyPodsNum = nameStatusMap[app.Name].ReadyPodsNum
			} else {
				apps[idx].Status = types.AppStatusUninstall
			}
		}
	}
	return apps, nil
}

func (a *AppService) GetApp(appId uint) *utils.Response {
	app, err := a.models.AppManager.GetAppWithVersion(appId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	var clusterId string
	var namespace string
	if app.Scope == types.ScopeProject {
		projectObj, err := a.models.ProjectManager.Get(app.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
		clusterId = projectObj.ClusterId
		namespace = projectObj.Namespace
	} else {
		clusterId = fmt.Sprintf("%d", app.ScopeId)
		namespace = app.Namespace
	}
	clusterObj, err := a.models.ClusterManager.GetByName(clusterId)
	if err != nil {
		klog.Errorf("get app %s cluster error: %s", appId, err.Error())
	}
	data := map[string]interface{}{
		"id":              app.ID,
		"name":            app.Name,
		"status":          app.Status,
		"cluster_id":      clusterId,
		"cluster":         clusterObj,
		"namespace":       namespace,
		"app_version_id":  app.AppVersionId,
		"app_version":     app.AppVersion.AppVersion,
		"type":            app.Type,
		"from":            app.AppVersion.From,
		"update_user":     app.UpdateUser,
		"create_time":     app.CreateTime,
		"update_time":     app.UpdateTime,
		"package_name":    app.AppVersion.PackageName,
		"package_version": app.AppVersion.PackageVersion,
	}
	appCharts, err := a.models.AppVersionManager.GetChart(app.AppVersion.ChartPath)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	chart, err := loader.LoadArchive(bytes.NewReader(appCharts.Content))
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: err.Error()}
	}
	actionConfig := new(action.Configuration)
	clientInstall := action.NewInstall(actionConfig)
	clientInstall.ReleaseName = app.Name
	clientInstall.Namespace = namespace
	clientInstall.ClientOnly = true
	clientInstall.DryRun = true
	values := map[string]interface{}{}
	yaml.Unmarshal([]byte(app.AppVersion.Values), &values)
	releaseDetail, err := clientInstall.Run(chart, values)
	if err != nil {
		klog.Errorf("install release error: %s", err)
		return &utils.Response{Code: code.HelmError, Msg: err.Error()}
	}
	data["manifest"] = releaseDetail.Manifest
	if app.Status == types.AppStatusUninstall {
		objects := a.GetReleaseObjects(releaseDetail)
		data["release"] = map[string]interface{}{
			"objects":       objects,
			"name":          releaseDetail.Name,
			"namespace":     releaseDetail.Namespace,
			"version":       releaseDetail.Version,
			"status":        releaseDetail.Info.Status,
			"chart_name":    releaseDetail.Chart.Name() + "-" + releaseDetail.Chart.Metadata.Version,
			"chart_version": releaseDetail.Chart.Metadata.Version,
			"app_version":   releaseDetail.Chart.AppVersion(),
			"last_deployed": releaseDetail.Info.LastDeployed,
		}
	} else {
		statusParams := map[string]interface{}{
			"namespace":     namespace,
			"name":          app.Name,
			"with_resource": true,
		}
		//nameStatusMap := map[string]*AppRuntimeStatus{}
		res := a.kubeClient.Get(clusterId, kubetypes.HelmType, statusParams)
		if res.IsSuccess() {
			var appStatuses AppRuntimeStatus
			if err := utils.ConvertTypeByJson(res.Data, &appStatuses); err != nil {
				klog.Error("unmarshal app status error: ", err.Error())
			} else {
				data["status"] = appStatuses.RuntimeStatus
			}
		} else {
			klog.Error("get app status error: ", res.Msg)
		}
		data["release"] = res.Data
	}

	return &utils.Response{Code: code.Success, Data: data}
}

func (a *AppService) GetReleaseObjects(release *release.Release) []*unstructured.Unstructured {
	var objects []*unstructured.Unstructured
	yamlList := strings.SplitAfter(release.Manifest, "\n---")
	for _, objectYaml := range yamlList {
		unstructuredObj := &unstructured.Unstructured{}
		yamlBytes := []byte(objectYaml)
		decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewBuffer(yamlBytes), len(yamlBytes))
		if err := decoder.Decode(unstructuredObj); err != nil {
			klog.Error("decode k8sObject objectYaml: ", objectYaml)
			klog.Error("decode k8sObject error: ", err)
			continue
		} else {
			objects = append(objects, unstructuredObj)
		}
	}
	return objects
}

func (a *AppService) GetAppStatus(scope string, scopeId uint, name string) (map[string]*AppRuntimeStatus, error) {
	apps, err := a.models.AppManager.ListApps(project.ListAppCondition{
		Scope:       scope,
		ScopeId:     scopeId,
		WithVersion: false,
		Name:        name,
	})
	if err != nil {
		return nil, errors.New(code.DBError, err)
	}
	nameStatusMap, err := a.updateAppStatus(scope, scopeId, apps)
	if err != nil {
		return nil, err
	}
	if nameStatusMap == nil {
		return nameStatusMap, nil
	}
	for _, app := range apps {
		status := ""
		if _, ok := nameStatusMap[app.Name]; ok {
			status = nameStatusMap[app.Name].RuntimeStatus
		} else {
			status = types.AppStatusUninstall
		}
		if status != "" && app.Status != status {
			app.Status = status
			err = a.models.AppManager.UpdateApp(app, "status")
			if err != nil {
				klog.Error("update project app status error: ", err.Error())
			}
		}
	}

	return nameStatusMap, nil
}

func (a *AppService) GetAppVersion(appVersionId uint) *utils.Response {
	appVersion, err := a.models.AppVersionManager.GetById(appVersionId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	app, err := a.models.AppManager.GetAppWithVersion(appVersion.ScopeId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	appCharts, err := a.models.AppVersionManager.GetChart(appVersion.ChartPath)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	charts, err := loader.LoadArchive(bytes.NewReader(appCharts.Content))
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: err.Error()}
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
	return &utils.Response{Code: code.Success, Data: res}
}

type ImportStoreAppForm struct {
	Scope        string `json:"scope" form:"scope"`
	ScopeId      uint   `json:"scope_id" form:"scope_id"`
	Namespace    string `json:"namespace" form:"namespace"`
	StoreAppId   uint   `json:"store_app_id" form:"store_app_id"`
	AppVersionId uint   `json:"app_version_id" form:"app_version_id"`
	User         string `json:"user" form:"user"`
}

// ImportStoreApp 从应用商店导入应用
func (a *AppService) ImportStoreApp(importForm *ImportStoreAppForm) (*types.App, *types.AppVersion, error) {
	storeApp, err := a.models.AppStoreManager.GetById(importForm.StoreAppId)
	if err != nil {
		return nil, nil, errors.New(code.DataNotExists, "获取商店应用失败: "+err.Error())
	}
	storeAppVersion, err := a.models.AppVersionManager.GetById(importForm.AppVersionId)
	if err != nil {
		return nil, nil, errors.New(code.DataNotExists, "获取商店应用版本失败: "+err.Error())
	}
	app, err := a.models.AppManager.GetByName(importForm.Scope, importForm.ScopeId, storeApp.Name)
	if err != nil {
		return nil, nil, errors.New(code.DBError, err.Error())
	}
	if app != nil {
		sameVersion, err := a.models.AppVersionManager.GetByPackageNameVersion(types.ScopeProject, app.ID, storeAppVersion.PackageName, storeAppVersion.PackageVersion)
		if err != nil {
			return nil, nil, errors.New(code.DBError, err.Error())
		}
		if sameVersion != nil {
			return app, storeAppVersion, errors.New(code.ParamsError, "该应用已存在相同版本，请重新选择应用版本")
		}
	} else {
		app = &types.App{
			Scope:      importForm.Scope,
			ScopeId:    importForm.ScopeId,
			Name:       storeApp.Name,
			Status:     types.AppStatusUninstall,
			Namespace:  importForm.Namespace,
			Type:       storeApp.Type,
			CreateUser: importForm.User,
			UpdateUser: importForm.User,
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		}
	}
	storeAppVersion.ID = 0
	storeAppVersion.ScopeId = app.ID
	storeAppVersion.Scope = importForm.Scope
	if err := a.models.AppManager.ImportApp(app, storeAppVersion); err != nil {
		return app, storeAppVersion, errors.New(code.DBError, "导入应用失败: "+err.Error())
	}
	return app, storeAppVersion, nil
}

func (a *AppService) ImportProjectApp(
	originApp *types.App,
	version *types.AppVersion,
	destProjectId uint,
	destAppName string,
	user string) error {
	app, err := a.models.AppManager.GetByName(types.ScopeProject, destProjectId, destAppName)
	if err != nil {
		return errors.New(code.DBError, err)
	}
	if app != nil {
		sameVersion, err := a.models.AppVersionManager.GetByPackageNameVersion(types.ScopeProject, app.ID, version.PackageName, version.PackageVersion)
		if err != nil {
			return errors.New(code.DBError, err)
		}
		if sameVersion != nil {
			return errors.New(code.ParamsError, "该应用已存在相同版本，请重新选择应用版本")
		}
	} else {
		app = &types.App{
			Scope:      types.ScopeProject,
			ScopeId:    destProjectId,
			Name:       destAppName,
			Status:     types.AppStatusUninstall,
			Type:       originApp.Type,
			CreateUser: user,
			UpdateUser: user,
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		}
	}
	version.ID = 0
	version.ScopeId = app.ID
	version.Scope = types.ScopeProject
	if err := a.models.AppManager.ImportApp(app, version); err != nil {
		return errors.New(code.DBError, "克隆应用失败: "+err.Error())
	}
	return nil
}

// ImportToStore 工作空间应用发布到应用商店
func (a *AppService) ImportToStore(originApp *types.App, version *types.AppVersion, destAppName string, user string) error {
	app, err := a.models.AppStoreManager.GetByName(destAppName)
	if err != nil {
		return errors.New(code.DBError, err)
	}
	if app != nil {
		sameVersion, err := a.models.AppVersionManager.GetByPackageNameVersion(types.ScopeAppStore, app.ID, version.PackageName, version.PackageVersion)
		if err != nil {
			return errors.New(code.DBError, err)
		}
		if sameVersion != nil {
			return errors.New(code.ParamsError, "该应用已存在相同版本")
		}
		app.UpdateUser = user
	} else {
		app = &types.AppStore{
			Name:        destAppName,
			Description: originApp.Description,
			Type:        originApp.Type,
			Icon:        nil,
			CreateUser:  user,
			UpdateUser:  user,
			CreateTime:  time.Now(),
			UpdateTime:  time.Now(),
		}
	}
	version.ID = 0
	version.ScopeId = app.ID
	version.Scope = types.ScopeAppStore
	if err := a.models.AppStoreManager.ImportApp(app, version); err != nil {
		return errors.New(code.DBError, "发布应用失败: "+err.Error())
	}
	return nil
}

type DuplicateAppForm struct {
	Name      string `json:"name" form:"name"`
	AppId     uint   `json:"app_id" form:"app_id"`
	VersionId uint   `json:"version_id" form:"version_id"`
	Scope     string `json:"scope" form:"scope"`
	ScopeId   uint   `json:"scope_id" form:"scope_id"`
	User      string `json:"user" form:"user"`
}

// DuplicateApp 克隆应用到工作空间，或者发布到应用商店
func (a *AppService) DuplicateApp(ser *DuplicateAppForm) (*types.App, *types.AppVersion, error) {
	originApp, err := a.models.AppManager.GetAppWithVersion(ser.AppId)
	if err != nil {
		return nil, nil, errors.New(code.DataNotExists, "获取应用失败："+err.Error())
	}
	version, err := a.models.AppVersionManager.GetById(ser.VersionId)
	if err != nil {
		return nil, nil, errors.New(code.DBError, "获取应用版本失败："+err.Error())
	}
	if ser.Scope == types.ScopeProject {
		return originApp, version, a.ImportProjectApp(originApp, version, ser.ScopeId, ser.Name, ser.User)
	} else if ser.Scope == types.ScopeAppStore {
		return originApp, version, a.ImportToStore(originApp, version, ser.Name, ser.User)
	} else {
		return originApp, version, errors.New(code.ParamsError, "参数scope错误")
	}
}

type ImportCustomAppForm struct {
	Scope              string `json:"scope" form:"scope"`
	ScopeId            uint   `json:"scope_id" form:"scope_id"`
	Name               string `json:"name" form:"name"`
	PackageVersion     string `json:"package_version" form:"package_version"`
	AppVersion         string `json:"app_version" form:"app_version"`
	Description        string `json:"description" form:"description"`
	VersionDescription string `json:"version_description" form:"version_description"`
	Type               string `json:"type" form:"type"`
	ChartBytes         []byte `json:"chart_bytes" form:"chart_bytes"`
	User               string `json:"user" form:"user"`
}

// ImportCustomApp 工作空间导入自定义应用
func (a *AppService) ImportCustomApp(serializer *ImportCustomAppForm) (*types.App, *types.AppVersion, error) {
	app, err := a.models.AppManager.GetByName(serializer.Scope, serializer.ScopeId, serializer.Name)
	if err != nil {
		return nil, nil, errors.New(code.DBError, err)
	}
	if app != nil {
		sameVersion, err := a.models.AppVersionManager.GetByPackageNameVersion(types.ScopeProject, app.ID, serializer.Name, serializer.PackageVersion)
		if err != nil {
			return app, nil, errors.New(code.DBError, err)
		}
		if sameVersion != nil {
			return app, nil, errors.New(code.ParamsError, "该应用已存在相同版本，请重新输入版本号")
		}
		app.UpdateUser = serializer.User
		app.UpdateTime = time.Now()
		app.Description = serializer.Description
	} else {
		if serializer.Type != types.AppTypeOrdinaryApp && serializer.Type != types.AppTypeMiddleware {
			return nil, nil, errors.New(code.ParamsError, "应用类型参数错误")
		}
		app = &types.App{
			Scope:       serializer.Scope,
			ScopeId:     serializer.ScopeId,
			Name:        serializer.Name,
			Status:      types.AppStatusUninstall,
			Type:        serializer.Type,
			Description: serializer.Description,
			CreateUser:  serializer.User,
			UpdateUser:  serializer.User,
			CreateTime:  time.Now(),
			UpdateTime:  time.Now(),
		}
	}

	charts, err := loader.LoadArchive(bytes.NewBuffer(serializer.ChartBytes))
	if err != nil {
		return app, nil, errors.New(code.GetError, err.Error())
	}
	values := ""
	for _, rawFile := range charts.Raw {
		if rawFile.Name == "values.yaml" {
			values = string(rawFile.Data)
			break
		}
	}
	appVersion := &types.AppVersion{
		PackageName:    serializer.Name,
		PackageVersion: serializer.PackageVersion,
		AppVersion:     serializer.AppVersion,
		Values:         values,
		Description:    serializer.VersionDescription,
		From:           types.AppVersionFromImport,
		CreateUser:     serializer.User,
		CreateTime:     time.Now(),
		UpdateTime:     time.Now(),
	}
	_, err = a.models.AppManager.CreateAppWithBytes(serializer.ChartBytes, app, appVersion)
	if err != nil {
		return app, appVersion, errors.New(code.CreateError, err.Error())
	}
	return app, appVersion, nil
}

// GetAppChartFiles 获取应用helm chart所有文件
func (a *AppService) GetAppChartFiles(appVersionId uint) (*types.App, *types.AppVersion, map[string]interface{}, error) {
	appVersion, err := a.models.AppVersionManager.GetById(appVersionId)
	if err != nil {
		return nil, nil, nil, errors.New(code.DataNotExists, err)
	}
	app, err := a.models.AppManager.GetById(appVersion.ScopeId)
	if err != nil {
		return nil, nil, nil, errors.New(code.DataNotExists, err)
	}
	appChart, err := a.models.AppVersionManager.GetChart(appVersion.ChartPath)
	if err != nil {
		return nil, nil, nil, errors.New(code.DataNotExists, err)
	}
	files, err := utils.ExtractTgzBytes(appChart.Content)
	if err != nil {
		return nil, nil, nil, errors.New(code.DecodeError, err)
	}

	chartfiles := map[string]interface{}{}
	// 平铺目录处理为层级结构
	for path, content := range files {
		parts := strings.Split(path, "/")
		node := chartfiles
		// 从 1 开始，去除第一层目录
		for i := 1; i < len(parts)-1; i += 1 {
			if sub, ok := node[parts[i]]; !ok {
				newNode := map[string]interface{}{}
				node[parts[i]] = newNode
				node = newNode
			} else {
				node = sub.(map[string]interface{})
			}
		}
		node[parts[len(parts)-1]] = base64.StdEncoding.EncodeToString(content)
	}
	if app.AppVersionId == appVersionId {
		// 如果要编辑的版本是当前应用的版本，替换最新的values.yaml
		chartfiles["values.yaml"] = base64.StdEncoding.EncodeToString([]byte(appVersion.Values))
	}
	return app, appVersion, chartfiles, nil
}
