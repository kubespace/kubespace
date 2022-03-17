package project

import (
	"bytes"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"helm.sh/helm/v3/pkg/chart/loader"
	"io"
	"sigs.k8s.io/yaml"
	"time"
)

type AppStoreService struct {
	*AppBaseService
}

func NewAppStoreService(appBaseService *AppBaseService) *AppStoreService {
	return &AppStoreService{
		AppBaseService: appBaseService,
	}
}

func (s *AppStoreService) CreateStoreApp(user *types.User, serializer serializers.AppStoreCreateSerializer, chartIn io.Reader) *utils.Response {
	app, err := s.models.AppStoreManager.GetStoreAppByName(serializer.Name)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if app != nil {
		sameVersion, err := s.models.ProjectAppManager.GetAppVersion(types.AppVersionScopeStoreApp, app.ID, serializer.Name, serializer.Version)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
		if sameVersion != nil {
			return &utils.Response{Code: code.ParamsError, Msg: "该应用已存在相同版本"}
		}
		app.UpdateUser = user.Name
	} else {
		app = &types.AppStore{
			Name:        serializer.Name,
			Description: serializer.Description,
			Type:        serializer.Type,
			CreateUser:  user.Name,
			UpdateUser:  user.Name,
			CreateTime:  time.Now(),
			UpdateTime:  time.Now(),
		}
	}
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, chartIn); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	chartBytes := buf.Bytes()
	charts, err := loader.LoadArchive(chartIn)
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: err.Error()}
	}
	valuesByte, _ := yaml.Marshal(charts.Values)
	appVersion := &types.AppVersion{
		PackageName:    serializer.Name,
		PackageVersion: serializer.Version,
		AppVersion:     serializer.Version,
		From:           types.AppVersionFromImport,
		Values:         string(valuesByte),
		CreateUser:     user.Name,
		CreateTime:     time.Now(),
		UpdateTime:     time.Now(),
	}
	app, err = s.models.AppStoreManager.CreateStoreApp(chartBytes, app, appVersion)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: "创建应用失败:" + err.Error()}
	}
	return &utils.Response{Code: code.Success, Data: app}
}

//func (a *AppStoreService) ListStoreApp()
