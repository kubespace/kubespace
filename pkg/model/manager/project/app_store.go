package project

import (
	"encoding/json"
	"errors"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
	"helm.sh/helm/v3/pkg/chart/loader"
	"io/ioutil"
	"k8s.io/klog"
	"os"
	"path/filepath"
	"time"
)

type AppStoreManager struct {
	*gorm.DB
	*AppVersionManager
}

func NewAppStoreManager(versionManager *AppVersionManager, db *gorm.DB) *AppStoreManager {
	appStoreMgr := &AppStoreManager{DB: db, AppVersionManager: versionManager}
	appStoreMgr.Init()
	return appStoreMgr
}

func (a *AppStoreManager) GetStoreApp(appId uint) (*types.AppStore, error) {
	var app types.AppStore
	var err error
	if err = a.DB.First(&app, "id = ?", appId).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (a *AppStoreManager) GetStoreAppByName(name string) (*types.AppStore, error) {
	var app types.AppStore
	var err error
	if err = a.DB.First(&app, "name = ?", name).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &app, nil
}

func (a *AppStoreManager) CreateStoreApp(chartBytes []byte, storeApp *types.AppStore, appVersion *types.AppVersion) (*types.AppStore, error) {
	var err error
	err = a.DB.Transaction(func(tx *gorm.DB) error {
		if storeApp.ID == 0 {
			if err = tx.Create(storeApp).Error; err != nil {
				return err
			}
		}
		appVersion, err = a.AppVersionManager.CreateAppVersionWithChartByte(chartBytes, types.AppVersionScopeStoreApp, storeApp.ID, appVersion)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return storeApp, nil
}

func (a *AppStoreManager) UpdateStoreApp(appId uint, columns map[string]interface{}) error {
	return a.DB.Model(&types.AppStore{}).Where("id=?", appId).Updates(columns).Error
}

func (a *AppStoreManager) DeleteStoreAppVersion(appId, versionId uint, user *types.User) error {
	err := a.AppVersionManager.DeleteVersion(versionId)
	if err != nil {
		return err
	}
	versions, _ := a.AppVersionManager.ListAppVersions(types.AppVersionScopeStoreApp, appId)
	if versions != nil && len(*versions) == 0 {
		return a.DB.Delete(&types.AppStore{}, "id=?", appId).Error
	} else {
		return a.UpdateStoreApp(appId, map[string]interface{}{"update_user": user.Name})
	}
}

func (a *AppStoreManager) ListStoreApps() ([]*types.AppStore, error) {
	var apps []types.AppStore
	var err error
	if err = a.DB.Find(&apps).Error; err != nil {
		return nil, err
	}
	var rets []*types.AppStore
	for i, _ := range apps {
		rets = append(rets, &apps[i])
	}
	return rets, nil
}

func (a *AppStoreManager) ImportApp(storeApp *types.AppStore, appVersion *types.AppVersion) error {
	return a.DB.Transaction(func(tx *gorm.DB) error {
		if storeApp.ID == 0 {
			if err := tx.Create(storeApp).Error; err != nil {
				return err
			}
			appVersion.ScopeId = storeApp.ID
		}
		if err := tx.Create(appVersion).Error; err != nil {
			return err
		}
		return nil
	})
}

func (a *AppStoreManager) GetLatestVersion(appId uint) (*types.AppVersion, error) {
	var appVersion types.AppVersion
	if err := a.DB.Order("create_time desc").First(&appVersion, "scope = ? and scope_id = ?", types.AppVersionScopeStoreApp, appId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &appVersion, nil
}

func (a *AppStoreManager) Init() {
	apps, err := a.ListStoreApps()
	if err != nil {
		klog.Errorf("list store apps error: %s", err)
		return
	}
	if len(apps) == 0 {
		pwd, _ := os.Getwd()
		//获取文件或目录相关信息
		appDir := filepath.Join(pwd, "apps")
		fileInfoList, err := ioutil.ReadDir(appDir)
		if err != nil {
			klog.Errorf("read dir error: %v", err)
		}
		for i := range fileInfoList {
			klog.Infof("start load %s", fileInfoList[i].Name())
			a.loadApp(filepath.Join(appDir, fileInfoList[i].Name()))

		}
	}
}

type AppFileMetadata struct {
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
	Type     string   `json:"type"`
}

func (a *AppStoreManager) loadApp(appPath string) {
	metadataBytes, err := ioutil.ReadFile(filepath.Join(appPath, "metadata"))
	if err != nil {
		klog.Errorf("read app metadata %s error: %s", appPath, err)
		return
	}
	var metadata AppFileMetadata
	err = json.Unmarshal(metadataBytes, &metadata)
	if err != nil {
		klog.Errorf("unmarshal metadata %s error: %s", appPath, err)
		return
	}
	if metadata.Name == "" {
		klog.Errorf("metadata %s app name is empty", appPath)
		return
	}
	if len(metadata.Versions) == 0 {
		klog.Errorf("app %s versions is empty", appPath)
		return
	}
	if metadata.Type == "" {
		klog.Errorf("app %s type is empty", appPath)
		return
	}
	icon, err := ioutil.ReadFile(filepath.Join(appPath, "icon.png"))

	app := &types.AppStore{
		Name:       metadata.Name,
		Type:       metadata.Type,
		Icon:       icon,
		CreateUser: "admin",
		UpdateUser: "admin",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	for _, appVersion := range metadata.Versions {
		versionPath := filepath.Join(appPath, appVersion)
		charts, err := loader.LoadFile(versionPath)
		if err != nil {
			klog.Errorf("load %s file error: %s", versionPath, err)
			break
		}
		values := ""
		for _, rawFile := range charts.Raw {
			if rawFile.Name == "values.yaml" {
				values = string(rawFile.Data)
				break
			}
		}
		if app.Description == "" {
			app.Description = charts.Metadata.Description
		}
		newAppVersion := &types.AppVersion{
			PackageName:    charts.Name(),
			PackageVersion: charts.Metadata.Version,
			AppVersion:     charts.AppVersion(),
			From:           types.AppVersionFromImport,
			Values:         values,
			Description:    "平台内置应用",
			CreateUser:     "admin",
			CreateTime:     time.Now(),
			UpdateTime:     time.Now(),
		}
		chartByte, _ := ioutil.ReadFile(versionPath)
		_, err = a.CreateStoreApp(chartByte, app, newAppVersion)
		if err != nil {
			klog.Errorf("create app %s-%s error: %s", app.Name, newAppVersion.PackageVersion, err)
		}
	}
}
