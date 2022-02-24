package project

import (
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/klog"
	"os"
	"time"
)

type AppService struct {
	models *model.Models
}

func NewAppService(models *model.Models) *AppService {
	return &AppService{
		models: models,
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
			Values:     serializer.Values,
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
