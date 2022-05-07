package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog"
	"os"
	"sigs.k8s.io/yaml"
	"strings"
	"time"
)

type AppService struct {
	*kube_resource.KubeResources
	*AppBaseService
}

func NewAppService(kr *kube_resource.KubeResources, appBaseService *AppBaseService) *AppService {
	return &AppService{
		AppBaseService: appBaseService,
		KubeResources:  kr,
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

func (a *AppService) CreateProjectApp(user *types.User, serializer serializers.ProjectCreateAppSerializer) *utils.Response {
	app, err := a.models.ProjectAppManager.GetByName(serializer.Scope, serializer.ScopeId, serializer.Name)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if app != nil {
		sameVersion, err := a.models.ProjectAppManager.GetAppVersion(types.AppVersionScopeProjectApp, app.ID, serializer.Name, serializer.Version)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
		if sameVersion != nil {
			return &utils.Response{Code: code.ParamsError, Msg: "该应用已存在相同版本，请重新输入版本号"}
		}
		app.UpdateUser = user.Name
		app.UpdateTime = time.Now()
		app.Description = serializer.Description
	} else {
		if serializer.Type != types.AppTypeOrdinaryApp && serializer.Type != types.AppTypeMiddleware {
			return &utils.Response{Code: code.ParamsError, Msg: "应用类型参数错误"}
		}
		app = &types.ProjectApp{
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
		return &utils.Response{Code: code.CreateError, Msg: err.Error()}
	}
	if err = a.WriteFile(tmpChartDir+"/Chart.yaml", serializer.Chart); err != nil {
		return &utils.Response{Code: code.CreateError, Msg: err.Error()}
	}
	if err = a.WriteFile(tmpChartDir+"/values.yaml", serializer.Values); err != nil {
		return &utils.Response{Code: code.CreateError, Msg: err.Error()}
	}
	tmpTempDir := tmpChartDir + "/templates"
	err = os.MkdirAll(tmpTempDir, 0755)
	if err != nil {
		return &utils.Response{Code: code.CreateError, Msg: err.Error()}
	}
	for name, value := range serializer.Templates {
		if err = a.WriteFile(tmpTempDir+"/"+name, value); err != nil {
			return &utils.Response{Code: code.CreateError, Msg: err.Error()}
		}
	}
	pack := action.NewPackage()
	pack.Destination = tmpChartDir
	tgzPath, err := pack.Run(tmpChartDir, nil)
	if err != nil {
		return &utils.Response{Code: code.HelmError, Msg: err.Error()}
	}
	appVersion := &types.AppVersion{
		PackageName:    serializer.Name,
		PackageVersion: serializer.Version,
		AppVersion:     serializer.Version,
		Values:         serializer.Values,
		Description:    serializer.VersionDescription,
		From:           types.AppVersionFromSpace,
		CreateUser:     user.Name,
		CreateTime:     time.Now(),
		UpdateTime:     time.Now(),
	}
	_, err = a.models.ProjectAppManager.CreateProjectApp(tgzPath, app, appVersion)
	if err != nil {
		return &utils.Response{Code: code.CreateError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success}
}

func (a *AppService) InstallApp(user *types.User, serializer serializers.ProjectInstallAppSerializer) *utils.Response {
	versionApp, err := a.models.ProjectAppVersionManager.GetAppVersion(serializer.AppVersionId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if versionApp.Scope != types.AppVersionScopeProjectApp && versionApp.ScopeId != serializer.ProjectAppId {
		return &utils.Response{Code: code.ParamsError, Msg: "当前应用不存在该版本，请重新选择"}
	}
	projectApp, err := a.models.ProjectAppManager.GetProjectApp(serializer.ProjectAppId)
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
	installParams := map[string]interface{}{
		"name":       projectApp.Name,
		"namespace":  namespace,
		"chart_path": versionApp.ChartPath,
		"values":     serializer.Values,
	}
	var resp *utils.Response
	if serializer.Upgrade {
		resp = a.Helm.UpdateObj(clusterId, installParams)
	} else {
		resp = a.Helm.Create(clusterId, installParams)
	}
	if !resp.IsSuccess() {
		return resp
	}
	projectApp.AppVersionId = serializer.AppVersionId
	projectApp.UpdateUser = user.Name
	projectApp.Status = types.AppStatusNotReady
	if err = a.models.ProjectAppManager.UpdateProjectApp(projectApp, "status", "app_version_id", "update_user", "update_time"); err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	versionApp.Values = serializer.Values
	if err = a.models.ProjectAppVersionManager.UpdateAppVersion(versionApp, "values"); err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if _, err = a.models.ProjectAppManager.CreateRevision(versionApp, projectApp); err != nil {
		klog.Errorf("create project app id=%s, name=%s revision error: %s", projectApp.ID, projectApp.Name, err)
	}
	return &utils.Response{Code: code.Success}
}

func (a *AppService) DestroyApp(user *types.User, serializer serializers.ProjectInstallAppSerializer) *utils.Response {
	projectApp, err := a.models.ProjectAppManager.GetProjectApp(serializer.ProjectAppId)
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
	destroyParams := map[string]interface{}{
		"namespace": namespace,
		"name":      projectApp.Name,
	}
	resp := a.Helm.Delete(clusterId, destroyParams)
	if !resp.IsSuccess() {
		return resp
	}
	projectApp.UpdateUser = user.Name
	projectApp.Status = types.AppStatusUninstall
	if err = a.models.ProjectAppManager.UpdateProjectApp(projectApp, "status", "update_user", "update_time"); err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success}
}

type AppRuntimeStatus struct {
	Name      string                       `json:"name"`
	Status    string                       `json:"status"`
	Workloads []*unstructured.Unstructured `json:"workloads"`
}

func (a *AppService) updateAppStatus(scope string, scopeId uint, projectApps []*types.ProjectApp) (map[string]*AppRuntimeStatus, error) {
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
			"namespace": namespace,
			"names":     appNames,
		}
		res := a.KubeResources.Helm.Status(clusterId, statusParams)
		if res.IsSuccess() {
			var appStatuses []*AppRuntimeStatus
			dataBytes, _ := json.Marshal(res.Data)
			err := json.Unmarshal(dataBytes, &appStatuses)
			if err != nil {
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

func (a *AppService) ListApp(scope string, scopeId uint) ([]*types.ProjectApp, error) {
	projectApps, err := a.models.ProjectAppManager.ListProjectApps(scope, scopeId)
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
				projectApps[idx].Status = nameStatusMap[app.Name].Status
			} else {
				projectApps[idx].Status = types.AppStatusUninstall
			}
		}
	}
	return projectApps, nil
}

func (a *AppService) GetApp(appId uint) *utils.Response {
	projectApp, err := a.models.ProjectAppManager.GetProjectApp(appId)
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
	cluster, err := a.models.ClusterManager.GetByName(clusterId)
	if err != nil {
		klog.Errorf("get app %s cluster error: %s", appId, err.Error())
	}
	data := map[string]interface{}{
		"id":              projectApp.ID,
		"name":            projectApp.Name,
		"status":          projectApp.Status,
		"cluster_id":      clusterId,
		"cluster":         cluster,
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
	appCharts, err := a.models.ProjectAppVersionManager.GetAppVersionChart(projectApp.AppVersion.ChartPath)
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
			"namespace":      namespace,
			"name":           projectApp.Name,
			"with_workloads": true,
		}
		nameStatusMap := map[string]*AppRuntimeStatus{}
		res := a.KubeResources.Helm.Get(clusterId, statusParams)
		if res.IsSuccess() {
			var appStatuses []*AppRuntimeStatus
			dataBytes, _ := json.Marshal(res.Data)
			err = json.Unmarshal(dataBytes, &appStatuses)
			if err != nil {
				klog.Error("unmarshal app status error: ", err.Error())
			}
			for _, status := range appStatuses {
				nameStatusMap[status.Name] = status
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

func (a *AppService) ListAppStatus(serializer serializers.ProjectAppListSerializer) *utils.Response {
	var projectApps []*types.ProjectApp
	var err error
	if serializer.Name != "" {
		app, err := a.models.ProjectAppManager.GetProjectAppByName(serializer.Scope, serializer.ScopeId, serializer.Name)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
		projectApps = append(projectApps, app)
	} else {
		projectApps, err = a.models.ProjectAppManager.ListProjectApps(serializer.Scope, serializer.ScopeId)
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
				status = nameStatusMap[app.Name].Status
			} else {
				status = types.AppStatusUninstall
			}
			if status != "" && app.Status != status {
				app.Status = status
				err = a.models.ProjectAppManager.UpdateProjectApp(app, "status")
				if err != nil {
					klog.Error("update project app status error: ", err.Error())
				}
			}
		}
	}
	return &utils.Response{Code: code.Success, Data: nameStatusMap}
}

func (a *AppService) GetAppVersion(appVersionId uint) *utils.Response {
	appVersion, err := a.models.ProjectAppVersionManager.GetAppVersion(appVersionId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	app, err := a.models.ProjectAppManager.GetProjectApp(appVersion.ScopeId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	appCharts, err := a.models.ProjectAppVersionManager.GetAppVersionChart(appVersion.ChartPath)
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

func (a *AppService) ImportStoreApp(ser serializers.ImportStoreAppSerializers, user *types.User) *utils.Response {
	storeApp, err := a.models.AppStoreManager.GetStoreApp(ser.StoreAppId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: "获取商店应用失败: " + err.Error()}
	}
	storeAppVersion, err := a.models.ProjectAppVersionManager.GetAppVersion(ser.AppVersionId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: "获取商店应用版本失败: " + err.Error()}
	}
	app, err := a.models.ProjectAppManager.GetByName(ser.Scope, ser.ScopeId, storeApp.Name)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if app != nil {
		sameVersion, err := a.models.ProjectAppManager.GetAppVersion(types.AppVersionScopeProjectApp, app.ID, storeAppVersion.PackageName, storeAppVersion.PackageVersion)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
		if sameVersion != nil {
			return &utils.Response{Code: code.ParamsError, Msg: "该应用已存在相同版本，请重新选择应用版本"}
		}
		app.UpdateUser = user.Name
	} else {
		app = &types.ProjectApp{
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
	storeAppVersion.Scope = types.AppVersionScopeProjectApp
	if err := a.models.ProjectAppManager.ImportApp(app, storeAppVersion); err != nil {
		return &utils.Response{Code: code.DBError, Msg: "导入应用失败: " + err.Error()}
	}
	return &utils.Response{Code: code.Success}
}

func (a *AppService) ImportProjectApp(originApp *types.ProjectApp, version *types.AppVersion, destProjectId uint, destAppName string, user *types.User) *utils.Response {
	app, err := a.models.ProjectAppManager.GetByName(types.AppVersionScopeProjectApp, destProjectId, destAppName)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if app != nil {
		sameVersion, err := a.models.ProjectAppManager.GetAppVersion(types.AppVersionScopeProjectApp, app.ID, version.PackageName, version.PackageVersion)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
		if sameVersion != nil {
			return &utils.Response{Code: code.ParamsError, Msg: "该应用已存在相同版本，请重新选择应用版本"}
		}
		app.UpdateUser = user.Name
	} else {
		app = &types.ProjectApp{
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
	if err := a.models.ProjectAppManager.ImportApp(app, version); err != nil {
		return &utils.Response{Code: code.DBError, Msg: "克隆应用失败: " + err.Error()}
	}
	return &utils.Response{Code: code.Success}
}

func (a *AppService) ImportToStore(originApp *types.ProjectApp, version *types.AppVersion, destAppName string, user *types.User) *utils.Response {
	app, err := a.models.AppStoreManager.GetStoreAppByName(destAppName)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if app != nil {
		sameVersion, err := a.models.ProjectAppManager.GetAppVersion(types.AppVersionScopeStoreApp, app.ID, version.PackageName, version.PackageVersion)
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

func (a *AppService) DuplicateApp(ser *serializers.DuplicateAppSerializer, user *types.User) *utils.Response {
	originApp, err := a.models.ProjectAppManager.GetProjectApp(ser.AppId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: "获取应用失败：" + err.Error()}
	}
	version, err := a.models.ProjectAppVersionManager.GetAppVersion(ser.VersionId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: "获取应用版本失败：" + err.Error()}
	}
	if ser.Scope == types.AppVersionScopeProjectApp {
		return a.ImportProjectApp(originApp, version, ser.ScopeId, ser.Name, user)
	} else if ser.Scope == types.AppVersionScopeStoreApp {
		return a.ImportToStore(originApp, version, ser.Name, user)
	} else {
		return &utils.Response{Code: code.ParamsError, Msg: "参数scope错误"}
	}
}
