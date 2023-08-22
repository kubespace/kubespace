package project

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	kubetypes "github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/service/cluster"
	"github.com/kubespace/kubespace/pkg/third/helm"
	"github.com/kubespace/kubespace/pkg/utils"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"io"
	"io/ioutil"
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

func (a *AppService) CreateProjectApp(user *types.User, serializer serializers.CreateAppSerializer) (*types.App, *utils.Response) {
	app, err := a.models.AppManager.GetByName(serializer.Scope, serializer.ScopeId, serializer.Name)
	if err != nil {
		return nil, &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if app != nil {
		sameVersion, err := a.models.AppVersionManager.GetByPackageNameVersion(types.AppVersionScopeProjectApp, app.ID, serializer.Name, serializer.Version)
		if err != nil {
			return nil, &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
		if sameVersion != nil {
			return nil, &utils.Response{Code: code.ParamsError, Msg: "该应用已存在相同版本，请重新输入版本号"}
		}
		app.UpdateUser = user.Name
		app.UpdateTime = time.Now()
		app.Description = serializer.Description
	} else {
		if serializer.Type != types.AppTypeOrdinaryApp && serializer.Type != types.AppTypeMiddleware {
			return nil, &utils.Response{Code: code.ParamsError, Msg: "应用类型参数错误"}
		}
		app = &types.App{
			Scope:       serializer.Scope,
			ScopeId:     serializer.ScopeId,
			Name:        serializer.Name,
			Status:      types.AppStatusUninstall,
			Type:        serializer.Type,
			Description: serializer.Description,
			CreateUser:  user.Name,
			UpdateUser:  user.Name,
			CreateTime:  time.Now(),
			UpdateTime:  time.Now(),
		}
	}
	tmpChartDir, err := os.MkdirTemp("/tmp", "")
	defer os.RemoveAll(tmpChartDir)
	if err != nil {
		return nil, &utils.Response{Code: code.CreateError, Msg: err.Error()}
	}
	chartGen := &helm.ChartGeneration{
		NeedModifyVersion: false,
		PackageVersion:    serializer.Version,
		AppVersion:        serializer.Version,
		Files:             serializer.ChartFiles,
		Base64Encoded:     false,
	}
	if serializer.From == types.AppVersionFromImport {
		chartGen.NeedModifyVersion = true
		chartGen.Base64Encoded = true
	}
	chartDir, chartPath, err := chartGen.GenerateChart()
	if chartDir != "" {
		defer os.RemoveAll(chartDir)
	}
	if err != nil {
		return nil, &utils.Response{Code: code.HelmError, Msg: err.Error()}
	}
	appVersion := &types.AppVersion{
		PackageName:    serializer.Name,
		PackageVersion: serializer.Version,
		AppVersion:     serializer.Version,
		Values:         serializer.Values,
		Description:    serializer.VersionDescription,
		From:           serializer.From,
		CreateUser:     user.Name,
		CreateTime:     time.Now(),
		UpdateTime:     time.Now(),
	}
	_, err = a.models.AppManager.CreateApp(chartPath, app, appVersion)
	if err != nil {
		return nil, &utils.Response{Code: code.CreateError, Msg: err.Error()}
	}
	return app, &utils.Response{Code: code.Success}
}

func (a *AppService) InstallApp(user *types.User, serializer serializers.InstallAppSerializer) *utils.Response {
	versionApp, err := a.models.AppVersionManager.GetById(serializer.AppVersionId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if versionApp.Scope != types.AppVersionScopeProjectApp && versionApp.ScopeId != serializer.AppId {
		return &utils.Response{Code: code.ParamsError, Msg: "当前应用不存在该版本，请重新选择"}
	}
	app, err := a.models.AppManager.Get(serializer.AppId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	var clusterId string
	var namespace string
	if app.Scope == types.AppVersionScopeProjectApp {
		project, err := a.models.ProjectManager.Get(app.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
		clusterId = project.ClusterId
		namespace = project.Namespace
	} else {
		clusterId = fmt.Sprintf("%d", app.ScopeId)
		namespace = app.Namespace
	}
	appChart, err := a.models.AppVersionManager.GetChart(versionApp.ChartPath)
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: "not found chart path=" + versionApp.ChartPath}
	}
	installParams := map[string]interface{}{
		"name":        app.Name,
		"namespace":   namespace,
		"chart_bytes": appChart.Content,
		"values":      serializer.Values,
	}
	var resp *utils.Response
	if serializer.Upgrade {
		resp = a.kubeClient.Update(clusterId, kubetypes.HelmType, installParams)
	} else {
		resp = a.kubeClient.Create(clusterId, kubetypes.HelmType, installParams)
	}
	if !resp.IsSuccess() {
		return resp
	}
	app.AppVersionId = serializer.AppVersionId
	app.UpdateUser = user.Name
	app.Status = types.AppStatusNotReady
	if err = a.models.AppManager.UpdateApp(app, "status", "app_version_id", "update_user", "update_time"); err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	versionApp.Values = serializer.Values
	if err = a.models.AppVersionManager.UpdateAppVersion(versionApp, "values"); err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if _, err = a.models.AppManager.CreateRevision(versionApp, app); err != nil {
		klog.Errorf("create project app id=%s, name=%s revision error: %s", app.ID, app.Name, err)
	}
	return &utils.Response{Code: code.Success}
}

func (a *AppService) DestroyApp(user *types.User, serializer serializers.InstallAppSerializer) *utils.Response {
	app, err := a.models.AppManager.GetAppWithVersion(serializer.AppId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	var clusterId string
	var namespace string
	if app.Scope == types.AppVersionScopeProjectApp {
		project, err := a.models.ProjectManager.Get(app.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
		clusterId = project.ClusterId
		namespace = project.Namespace
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
		return resp
	}
	app.UpdateUser = user.Name
	app.Status = types.AppStatusUninstall
	if err = a.models.AppManager.UpdateApp(app, "status", "update_user", "update_time"); err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success}
}

type AppRuntimeStatus struct {
	Name          string                       `json:"name"`
	RuntimeStatus string                       `json:"runtime_status"`
	Objects       []*unstructured.Unstructured `json:"objects"`
	PodsNum       int                          `json:"pods_num"`
	ReadyPodsNum  int                          `json:"ready_pods_num"`
}

func (a *AppService) updateAppStatus(scope string, scopeId uint, projectApps []*types.App) (map[string]*AppRuntimeStatus, error) {
	var clusterId string
	var namespaceApps = map[string][]string{}
	if scope == types.AppVersionScopeProjectApp {
		project, err := a.models.ProjectManager.Get(scopeId)
		if err != nil {
			return nil, err
		}
		clusterId = project.ClusterId
		for _, app := range projectApps {
			namespaceApps[project.Namespace] = append(namespaceApps[project.Namespace], app.Name)
		}
	} else {
		clusterId = fmt.Sprintf("%d", scopeId)
		for _, app := range projectApps {
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
	projectApps, err := a.models.AppManager.ListApps(scope, scopeId)
	if err != nil {
		return nil, err
	}
	nameStatusMap, err := a.updateAppStatus(scope, scopeId, projectApps)
	if err != nil {
		return nil, err
	}
	if nameStatusMap != nil {
		for idx, app := range projectApps {
			if _, ok := nameStatusMap[app.Name]; ok {
				projectApps[idx].Status = nameStatusMap[app.Name].RuntimeStatus
				projectApps[idx].PodsNum = nameStatusMap[app.Name].PodsNum
				projectApps[idx].ReadyPodsNum = nameStatusMap[app.Name].ReadyPodsNum
			} else {
				projectApps[idx].Status = types.AppStatusUninstall
			}
		}
	}
	return projectApps, nil
}

func (a *AppService) GetApp(appId uint) *utils.Response {
	projectApp, err := a.models.AppManager.GetAppWithVersion(appId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	var clusterId string
	var namespace string
	if projectApp.Scope == types.AppVersionScopeProjectApp {
		project, err := a.models.ProjectManager.Get(projectApp.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
		clusterId = project.ClusterId
		namespace = project.Namespace
	} else {
		clusterId = fmt.Sprintf("%d", projectApp.ScopeId)
		namespace = projectApp.Namespace
	}
	clusterObj, err := a.models.ClusterManager.GetByName(clusterId)
	if err != nil {
		klog.Errorf("get app %s cluster error: %s", appId, err.Error())
	}
	data := map[string]interface{}{
		"id":              projectApp.ID,
		"name":            projectApp.Name,
		"status":          projectApp.Status,
		"cluster_id":      clusterId,
		"cluster":         clusterObj,
		"namespace":       namespace,
		"app_version_id":  projectApp.AppVersionId,
		"app_version":     projectApp.AppVersion.AppVersion,
		"type":            projectApp.Type,
		"from":            projectApp.AppVersion.From,
		"update_user":     projectApp.UpdateUser,
		"create_time":     projectApp.CreateTime,
		"update_time":     projectApp.UpdateTime,
		"package_name":    projectApp.AppVersion.PackageName,
		"package_version": projectApp.AppVersion.PackageVersion,
	}
	appCharts, err := a.models.AppVersionManager.GetChart(projectApp.AppVersion.ChartPath)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	chart, err := loader.LoadArchive(bytes.NewReader(appCharts.Content))
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: err.Error()}
	}
	actionConfig := new(action.Configuration)
	clientInstall := action.NewInstall(actionConfig)
	clientInstall.ReleaseName = projectApp.Name
	clientInstall.Namespace = namespace
	clientInstall.ClientOnly = true
	clientInstall.DryRun = true
	values := map[string]interface{}{}
	yaml.Unmarshal([]byte(projectApp.AppVersion.Values), &values)
	releaseDetail, err := clientInstall.Run(chart, values)
	if err != nil {
		klog.Errorf("install release error: %s", err)
		return &utils.Response{Code: code.HelmError, Msg: err.Error()}
	}
	data["manifest"] = releaseDetail.Manifest
	if projectApp.Status == types.AppStatusUninstall {
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
			"name":          projectApp.Name,
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

func (a *AppService) ListAppStatus(serializer serializers.AppListSerializer) *utils.Response {
	var projectApps []*types.App
	var err error
	if serializer.Name != "" {
		app, err := a.models.AppManager.GetByName(serializer.Scope, serializer.ScopeId, serializer.Name)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
		projectApps = append(projectApps, app)
	} else {
		projectApps, err = a.models.AppManager.ListApps(serializer.Scope, serializer.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
	}
	nameStatusMap, err := a.updateAppStatus(serializer.Scope, serializer.ScopeId, projectApps)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if nameStatusMap != nil {
		for _, app := range projectApps {
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
	}
	return &utils.Response{Code: code.Success, Data: nameStatusMap}
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

// ImportStoreApp 从应用商店导入应用
func (a *AppService) ImportStoreApp(ser serializers.ImportStoreAppSerializers, user *types.User) (*types.App, *types.AppVersion, *utils.Response) {
	storeApp, err := a.models.AppStoreManager.GetById(ser.StoreAppId)
	if err != nil {
		return nil, nil, &utils.Response{Code: code.DBError, Msg: "获取商店应用失败: " + err.Error()}
	}
	storeAppVersion, err := a.models.AppVersionManager.GetById(ser.AppVersionId)
	if err != nil {
		return nil, nil, &utils.Response{Code: code.DBError, Msg: "获取商店应用版本失败: " + err.Error()}
	}
	app, err := a.models.AppManager.GetByName(ser.Scope, ser.ScopeId, storeApp.Name)
	if err != nil {
		return nil, nil, &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if app != nil {
		sameVersion, err := a.models.AppVersionManager.GetByPackageNameVersion(types.AppVersionScopeProjectApp, app.ID, storeAppVersion.PackageName, storeAppVersion.PackageVersion)
		if err != nil {
			return nil, nil, &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
		if sameVersion != nil {
			return app, storeAppVersion, &utils.Response{Code: code.ParamsError, Msg: "该应用已存在相同版本，请重新选择应用版本"}
		}
		app.UpdateUser = user.Name
	} else {
		app = &types.App{
			Scope:      ser.Scope,
			ScopeId:    ser.ScopeId,
			Name:       storeApp.Name,
			Status:     types.AppStatusUninstall,
			Namespace:  ser.Namespace,
			Type:       storeApp.Type,
			CreateUser: user.Name,
			UpdateUser: user.Name,
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		}
	}
	storeAppVersion.ID = 0
	storeAppVersion.ScopeId = app.ID
	storeAppVersion.Scope = ser.Scope
	if err := a.models.AppManager.ImportApp(app, storeAppVersion); err != nil {
		return app, storeAppVersion, &utils.Response{Code: code.DBError, Msg: "导入应用失败: " + err.Error()}
	}
	return app, storeAppVersion, &utils.Response{Code: code.Success}
}

func (a *AppService) ImportProjectApp(originApp *types.App, version *types.AppVersion, destProjectId uint, destAppName string, user *types.User) *utils.Response {
	app, err := a.models.AppManager.GetByName(types.AppVersionScopeProjectApp, destProjectId, destAppName)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if app != nil {
		sameVersion, err := a.models.AppVersionManager.GetByPackageNameVersion(types.AppVersionScopeProjectApp, app.ID, version.PackageName, version.PackageVersion)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
		if sameVersion != nil {
			return &utils.Response{Code: code.ParamsError, Msg: "该应用已存在相同版本，请重新选择应用版本"}
		}
		app.UpdateUser = user.Name
	} else {
		app = &types.App{
			Scope:      types.AppVersionScopeProjectApp,
			ScopeId:    destProjectId,
			Name:       destAppName,
			Status:     types.AppStatusUninstall,
			Type:       originApp.Type,
			CreateUser: user.Name,
			UpdateUser: user.Name,
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		}
	}
	version.ID = 0
	version.ScopeId = app.ID
	version.Scope = types.AppVersionScopeProjectApp
	if err := a.models.AppManager.ImportApp(app, version); err != nil {
		return &utils.Response{Code: code.DBError, Msg: "克隆应用失败: " + err.Error()}
	}
	return &utils.Response{Code: code.Success}
}

// ImportToStore 工作空间应用发布到应用商店
func (a *AppService) ImportToStore(originApp *types.App, version *types.AppVersion, destAppName string, user *types.User) *utils.Response {
	app, err := a.models.AppStoreManager.GetByName(destAppName)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if app != nil {
		sameVersion, err := a.models.AppVersionManager.GetByPackageNameVersion(types.AppVersionScopeStoreApp, app.ID, version.PackageName, version.PackageVersion)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
		if sameVersion != nil {
			return &utils.Response{Code: code.ParamsError, Msg: "该应用已存在相同版本"}
		}
		app.UpdateUser = user.Name
	} else {
		app = &types.AppStore{
			Name:        destAppName,
			Description: originApp.Description,
			Type:        originApp.Type,
			Icon:        nil,
			CreateUser:  user.Name,
			UpdateUser:  user.Name,
			CreateTime:  time.Now(),
			UpdateTime:  time.Now(),
		}
	}
	version.ID = 0
	version.ScopeId = app.ID
	version.Scope = types.AppVersionScopeStoreApp
	if err := a.models.AppStoreManager.ImportApp(app, version); err != nil {
		return &utils.Response{Code: code.DBError, Msg: "发布应用失败: " + err.Error()}
	}
	return &utils.Response{Code: code.Success}
}

// DuplicateApp 克隆应用到工作空间，或者发布到应用商店
func (a *AppService) DuplicateApp(ser *serializers.DuplicateAppSerializer, user *types.User) (*types.App, *types.AppVersion, *utils.Response) {
	originApp, err := a.models.AppManager.GetAppWithVersion(ser.AppId)
	if err != nil {
		return nil, nil, &utils.Response{Code: code.DBError, Msg: "获取应用失败：" + err.Error()}
	}
	version, err := a.models.AppVersionManager.GetById(ser.VersionId)
	if err != nil {
		return nil, nil, &utils.Response{Code: code.DBError, Msg: "获取应用版本失败：" + err.Error()}
	}
	if ser.Scope == types.AppVersionScopeProjectApp {
		return originApp, version, a.ImportProjectApp(originApp, version, ser.ScopeId, ser.Name, user)
	} else if ser.Scope == types.AppVersionScopeStoreApp {
		return originApp, version, a.ImportToStore(originApp, version, ser.Name, user)
	} else {
		return originApp, version, &utils.Response{Code: code.ParamsError, Msg: "参数scope错误"}
	}
}

// ImportCustomApp 工作空间导入自定义应用
func (a *AppService) ImportCustomApp(
	user *types.User,
	serializer serializers.ImportCustomAppSerializer,
	chartIn io.Reader) (*types.App, *types.AppVersion, *utils.Response) {
	app, err := a.models.AppManager.GetByName(serializer.Scope, serializer.ScopeId, serializer.Name)
	if err != nil {
		return nil, nil, &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if app != nil {
		sameVersion, err := a.models.AppVersionManager.GetByPackageNameVersion(types.AppVersionScopeProjectApp, app.ID, serializer.Name, serializer.PackageVersion)
		if err != nil {
			return app, nil, &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
		if sameVersion != nil {
			return app, nil, &utils.Response{Code: code.ParamsError, Msg: "该应用已存在相同版本，请重新输入版本号"}
		}
		app.UpdateUser = user.Name
		app.UpdateTime = time.Now()
		app.Description = serializer.Description
	} else {
		if serializer.Type != types.AppTypeOrdinaryApp && serializer.Type != types.AppTypeMiddleware {
			return nil, nil, &utils.Response{Code: code.ParamsError, Msg: "应用类型参数错误"}
		}
		app = &types.App{
			Scope:       serializer.Scope,
			ScopeId:     serializer.ScopeId,
			Name:        serializer.Name,
			Status:      types.AppStatusUninstall,
			Type:        serializer.Type,
			Description: serializer.Description,
			CreateUser:  user.Name,
			UpdateUser:  user.Name,
			CreateTime:  time.Now(),
			UpdateTime:  time.Now(),
		}
	}

	chartBytes, err := ioutil.ReadAll(chartIn)
	if err != nil {
		return app, nil, &utils.Response{Code: code.GetError, Msg: "获取chart文件失败: " + err.Error()}
	}
	charts, err := loader.LoadArchive(bytes.NewBuffer(chartBytes))
	if err != nil {
		return app, nil, &utils.Response{Code: code.GetError, Msg: err.Error()}
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
		CreateUser:     user.Name,
		CreateTime:     time.Now(),
		UpdateTime:     time.Now(),
	}
	_, err = a.models.AppManager.CreateAppWithBytes(chartBytes, app, appVersion)
	if err != nil {
		return app, appVersion, &utils.Response{Code: code.CreateError, Msg: err.Error()}
	}
	return app, appVersion, &utils.Response{Code: code.Success}
}

// GetAppChartFiles 获取应用helm chart所有文件
func (a *AppService) GetAppChartFiles(appVersionId uint) (*types.App, *types.AppVersion, map[string]interface{}, error) {
	appVersion, err := a.models.AppVersionManager.GetById(appVersionId)
	if err != nil {
		return nil, nil, nil, err
	}
	app, err := a.models.AppManager.GetAppWithVersion(appVersion.ScopeId)
	if err != nil {
		return nil, nil, nil, err
	}
	appChart, err := a.models.AppVersionManager.GetChart(appVersion.ChartPath)
	if err != nil {
		return nil, nil, nil, err
	}
	files, err := utils.ExtractTgzBytes(appChart.Content)
	if err != nil {
		return nil, nil, nil, err
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
