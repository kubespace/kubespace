package project

import (
	"bytes"
	"encoding/base64"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/utils"
	"helm.sh/helm/v3/pkg/chart/loader"
	"io"
	"io/ioutil"
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

func (s *AppStoreService) CreateStoreApp(
	user *types.User,
	serializer serializers.AppStoreCreateSerializer,
	chartIn io.Reader,
	icon io.Reader) (*types.AppStore, *types.AppVersion, error) {
	app, err := s.models.AppStoreManager.GetByName(serializer.Name)
	if err != nil {
		return nil, nil, errors.New(code.DBError, err)
	}
	if app != nil {
		sameVersion, err := s.models.AppVersionManager.GetByPackageNameVersion(types.AppVersionScopeStoreApp, app.ID, serializer.Name, serializer.PackageVersion)
		if err != nil {
			return app, nil, errors.New(code.DBError, err)
		}
		if sameVersion != nil {
			return app, nil, errors.New(code.ParamsError, "该应用已存在相同版本")
		}
		app.UpdateUser = user.Name
	} else {
		var iconBytes []byte
		if icon != nil {
			iconBytes, _ = ioutil.ReadAll(icon)
		}
		app = &types.AppStore{
			Name:        serializer.Name,
			Description: serializer.Description,
			Type:        serializer.Type,
			Icon:        iconBytes,
			CreateUser:  user.Name,
			UpdateUser:  user.Name,
			CreateTime:  time.Now(),
			UpdateTime:  time.Now(),
		}
	}
	chartBytes, err := ioutil.ReadAll(chartIn)
	if err != nil {
		return app, nil, errors.New(code.ParamsError, "获取chart文件失败: "+err.Error())
	}
	charts, err := loader.LoadArchive(bytes.NewBuffer(chartBytes))
	if err != nil {
		return app, nil, errors.New(code.ParamsError, "加载chart文件失败: "+err.Error())
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
		Description:    serializer.VersionDescription,
		From:           types.AppVersionFromImport,
		Values:         values,
		CreateUser:     user.Name,
		CreateTime:     time.Now(),
		UpdateTime:     time.Now(),
	}
	app, err = s.models.AppStoreManager.CreateStoreApp(chartBytes, app, appVersion)
	if err != nil {
		return app, appVersion, errors.New(code.DBError, "创建应用失败:"+err.Error())
	}
	return app, appVersion, nil
}

func (s *AppStoreService) ListStoreApp(ser *serializers.GetStoreAppSerializer) *utils.Response {
	apps, err := s.models.AppStoreManager.ListStoreApps()
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: "获取应用商店列表失败:" + err.Error()}
	}
	var data []map[string]interface{}
	for _, app := range apps {
		latestVersion, err := s.models.AppStoreManager.GetLatestVersion(app.ID)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
		iconBase64 := ""
		if len(app.Icon) > 0 {
			iconBase64 = base64.StdEncoding.EncodeToString(app.Icon)
		}
		a := map[string]interface{}{
			"id":          app.ID,
			"type":        app.Type,
			"name":        app.Name,
			"description": app.Description,
			"icon":        iconBase64,
			"create_user": app.CreateUser,
			"update_user": app.UpdateUser,
			"create_time": app.CreateTime,
			"update_time": app.UpdateTime,
		}
		if latestVersion != nil {
			a["latest_package_version"] = latestVersion.PackageVersion
			a["latest_app_version"] = latestVersion.AppVersion
			a["from"] = latestVersion.From
		}
		if ser.WithVersions {
			var appVersions []map[string]interface{}
			versions, err := s.models.AppVersionManager.List(types.AppVersionScopeStoreApp, app.ID)
			if err != nil {
				return &utils.Response{Code: code.DBError, Msg: err.Error()}
			}
			for _, v := range *versions {
				appVersions = append(appVersions, map[string]interface{}{
					"id":              v.ID,
					"package_name":    v.PackageName,
					"package_version": v.PackageVersion,
					"app_version":     v.AppVersion,
				})
			}
			a["versions"] = appVersions
		}
		data = append(data, a)
	}
	return &utils.Response{Code: code.Success, Data: data}
}

func (s *AppStoreService) GetStoreApp(appId uint, ser serializers.GetStoreAppSerializer) *utils.Response {
	app, err := s.models.AppStoreManager.GetById(appId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: "获取应用失败:" + err.Error()}
	}
	iconBase64 := ""
	if len(app.Icon) > 0 {
		iconBase64 = base64.StdEncoding.EncodeToString(app.Icon)
	}
	data := map[string]interface{}{
		"id":          app.ID,
		"type":        app.Type,
		"name":        app.Name,
		"description": app.Description,
		"icon":        iconBase64,
		"create_user": app.CreateUser,
		"update_user": app.UpdateUser,
		"create_time": app.CreateTime,
		"update_time": app.UpdateTime,
	}
	if ser.WithVersions {
		versions, err := s.models.AppVersionManager.List(types.AppVersionScopeStoreApp, appId)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
		data["versions"] = versions
	}
	return &utils.Response{Code: code.Success, Data: data}
}

func (s *AppStoreService) DeleteVersion(appId, versionId uint, user *types.User) *utils.Response {
	if err := s.models.AppStoreManager.DeleteStoreAppVersion(appId, versionId, user); err != nil {
		return &utils.Response{Code: code.DBError, Msg: "删除应用版本失败：" + err.Error()}
	}
	return &utils.Response{Code: code.Success}
}
