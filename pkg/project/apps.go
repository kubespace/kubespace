package project

import (
	"bytes"
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/klog"
	"os"
	"time"
)

type AppService struct {
	*kube_resource.KubeResources
	models *model.Models
}

func NewAppService(kr *kube_resource.KubeResources, models *model.Models) *AppService {
	return &AppService{
		models:        models,
		KubeResources: kr,
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
	resp := a.Helm.Create(project.ClusterId, installParams)
	if !resp.IsSuccess() {
		return resp
	}
	projectApp.AppVersionId = serializer.AppVersionId
	projectApp.UpdateUser = user.Name
	projectApp.Status = types.AppStatusUnReady
	if err = a.models.ProjectAppManager.UpdateProjectApp(projectApp, "status", "app_version_id", "update_user"); err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	versionApp.Values = serializer.Values
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
	if err = a.models.ProjectAppManager.UpdateProjectApp(projectApp, "status", "update_user"); err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success}
}

func (a *AppService) ListApp(serializer serializers.ProjectAppListSerializer) *utils.Response {
	projectApps, err := a.models.ProjectAppManager.ListProjectApps(serializer.ProjectId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
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
		data = append(data, res)
	}
	return &utils.Response{Code: code.Success, Data: data}
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
