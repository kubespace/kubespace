package project

import (
	"bytes"
	"encoding/json"
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
	app, err := a.models.ProjectAppManager.GetByName(serializer.ProjectId, serializer.Name)
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
	} else {
		app = &types.ProjectApp{
			ProjectId:  serializer.ProjectId,
			Name:       serializer.Name,
			Status:     types.AppStatusUninstall,
			CreateUser: user.Name,
			UpdateUser: user.Name,
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
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
	klog.Info(tgzPath)
	appVersion := &types.AppVersion{
		PackageName:    serializer.Name,
		PackageVersion: serializer.Version,
		AppVersion:     serializer.Version,
		Values:         serializer.Values,
		Type:           types.AppVersionTypeOrdinaryApp,
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
	project, err := a.models.ProjectManager.Get(projectApp.ProjectId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	installParams := map[string]interface{}{
		"name":       projectApp.Name,
		"namespace":  project.Namespace,
		"chart_path": versionApp.ChartPath,
		"values":     serializer.Values,
	}
	var resp *utils.Response
	if serializer.Upgrade {
		resp = a.Helm.UpdateObj(project.ClusterId, installParams)
	} else {
		resp = a.Helm.Create(project.ClusterId, installParams)
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
	values, _ := json.Marshal(serializer.Values)
	versionApp.Values = string(values)
	if err = a.models.ProjectAppVersionManager.UpdateAppVersion(versionApp, "values"); err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success}
}

func (a *AppService) DestroyApp(user *types.User, serializer serializers.ProjectInstallAppSerializer) *utils.Response {
	projectApp, err := a.models.ProjectAppManager.GetProjectApp(serializer.ProjectAppId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	project, err := a.models.ProjectManager.Get(projectApp.ProjectId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	destroyParams := map[string]interface{}{
		"namespace": project.Namespace,
		"name":      projectApp.Name,
	}
	resp := a.Helm.Delete(project.ClusterId, destroyParams)
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

func (a *AppService) ListApp(serializer serializers.ProjectAppListSerializer) *utils.Response {
	projectApps, err := a.models.ProjectAppManager.ListProjectApps(serializer.ProjectId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	project, err := a.models.ProjectManager.Get(serializer.ProjectId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	var appNames []string
	for _, app := range projectApps {
		appNames = append(appNames, app.Name)
	}
	statusParams := map[string]interface{}{
		"namespace": project.Namespace,
		"names":     appNames,
	}
	nameStatusMap := map[string]*AppRuntimeStatus{}
	res := a.KubeResources.Helm.Status(project.ClusterId, statusParams)
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
		klog.Error("get app status error: ", err.Error())
	}
	var data []map[string]interface{}
	for _, app := range projectApps {
		res := map[string]interface{}{
			"id":              app.ID,
			"name":            app.Name,
			"status":          app.Status,
			"app_version_id":  app.AppVersionId,
			"type":            app.AppVersion.Type,
			"update_user":     app.UpdateUser,
			"update_time":     app.UpdateTime,
			"package_name":    app.AppVersion.PackageName,
			"package_version": app.AppVersion.PackageVersion,
		}
		if _, ok := nameStatusMap[app.Name]; ok {
			res["status"] = nameStatusMap[app.Name].Status
		}
		data = append(data, res)
		appNames = append(appNames, app.Name)
	}

	return &utils.Response{Code: code.Success, Data: data}
}

func (a *AppService) GetApp(appId uint) *utils.Response {
	projectApp, err := a.models.ProjectAppManager.GetProjectApp(appId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	project, err := a.models.ProjectManager.Get(projectApp.ProjectId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	data := map[string]interface{}{
		"id":              projectApp.ID,
		"name":            projectApp.Name,
		"status":          projectApp.Status,
		"cluster":         project.ClusterId,
		"namespace":       project.Namespace,
		"app_version_id":  projectApp.AppVersionId,
		"type":            projectApp.AppVersion.Type,
		"update_user":     projectApp.UpdateUser,
		"create_time":     projectApp.CreateTime,
		"update_time":     projectApp.UpdateTime,
		"package_name":    projectApp.AppVersion.PackageName,
		"package_version": projectApp.AppVersion.PackageVersion,
	}
	if projectApp.Status == types.AppStatusUninstall {
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
		clientInstall.Namespace = project.Namespace
		clientInstall.ClientOnly = true
		clientInstall.DryRun = true
		values := map[string]interface{}{}
		yaml.Unmarshal([]byte(projectApp.AppVersion.Values), &values)
		releaseDetail, err := clientInstall.Run(chart, values)
		if err != nil {
			klog.Errorf("install release error: %s", err)
			return &utils.Response{Code: code.HelmError, Msg: err.Error()}
		}
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
			"namespace":      project.Namespace,
			"name":           projectApp.Name,
			"with_workloads": true,
		}
		nameStatusMap := map[string]*AppRuntimeStatus{}
		res := a.KubeResources.Helm.Get(project.ClusterId, statusParams)
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
			klog.Error("get app status error: ", err.Error())
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
		app, err := a.models.ProjectAppManager.GetProjectAppByName(serializer.ProjectId, serializer.Name)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
		projectApps = append(projectApps, app)
	} else {
		projectApps, err = a.models.ProjectAppManager.ListProjectApps(serializer.ProjectId)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
	}
	project, err := a.models.ProjectManager.Get(serializer.ProjectId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	var appNames []string
	for _, app := range projectApps {
		appNames = append(appNames, app.Name)
	}
	statusParams := map[string]interface{}{
		"namespace": project.Namespace,
		"names":     appNames,
	}
	nameStatusMap := map[string]*AppRuntimeStatus{}
	res := a.KubeResources.Helm.Status(project.ClusterId, statusParams)
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
		for _, app := range projectApps {
			if _, ok := nameStatusMap[app.Name]; ok {
				if app.Status != nameStatusMap[app.Name].Status {
					app.Status = nameStatusMap[app.Name].Status
					err = a.models.ProjectAppManager.UpdateProjectApp(app, "status")
					if err != nil {
						klog.Error("update project app status error: ", err.Error())
					}
				}
			}
		}
	}
	return &utils.Response{Code: code.Success, Msg: res.Msg, Data: res.Data}
}

func (a *AppService) ListAppVersions(serializer serializers.ProjectAppVersionListSerializer) *utils.Response {
	appVersions, err := a.models.ProjectAppVersionManager.ListAppVersions(serializer.Scope, serializer.ScopeId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success, Data: appVersions}
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
		"package_name":    appVersion.PackageName,
		"package_version": appVersion.PackageVersion,
		"type":            appVersion.Type,
		"values":          appVersion.Values,
		"templates":       charts.Templates,
	}
	return &utils.Response{Code: code.Success, Data: res}
}
