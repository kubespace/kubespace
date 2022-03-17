package project

import (
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
	"helm.sh/helm/v3/pkg/chart/loader"
	"io/ioutil"
	"k8s.io/klog"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
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

func (a *AppStoreManager) GetStoreAppByName(name string) (*types.AppStore, error) {
	var app types.AppStore
	var err error
	if err = a.DB.First(&app, "name = ?", name).Error; err != nil {
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

func (a *AppStoreManager) Init() {
	apps, err := a.ListStoreApps()
	if err != nil {
		klog.Errorf("list store apps error: %s", err)
		return
	}
	if len(apps) == 0 {
		pwd, _ := os.Getwd()
		//获取文件或目录相关信息
		middlewareDir := filepath.Join(pwd, "kubespace_apps", "middleware")
		fileInfoList, err := ioutil.ReadDir(middlewareDir)
		if err != nil {
			klog.Errorf("read dir error: %v", err)
		}
		for i := range fileInfoList {
			klog.Infof("start load %s", fileInfoList[i].Name())
			if strings.HasSuffix(fileInfoList[i].Name(), ".tgz") {
				a.loadApp(filepath.Join(middlewareDir, fileInfoList[i].Name()), types.AppTypeMiddleware)
			}
		}

		componentDir := filepath.Join(pwd, "kubespace_apps", "component")
		fileInfoList, err = ioutil.ReadDir(componentDir)
		if err != nil {
			klog.Errorf("read dir error: %v", err)
		}
		for i := range fileInfoList {
			klog.Infof("start load %s", fileInfoList[i].Name())
			if strings.HasSuffix(fileInfoList[i].Name(), ".tgz") {
				a.loadApp(filepath.Join(middlewareDir, fileInfoList[i].Name()), types.AppTypeClusterComponent)
			}
		}
	}
}

func (a *AppStoreManager) loadApp(filePath string, appType string) {
	charts, err := loader.LoadFile(filePath)
	if err != nil {
		klog.Errorf("load %s file error: %s", filePath, err)
		return
	}

	app := &types.AppStore{
		Name:        charts.Name(),
		Description: charts.Metadata.Description,
		Type:        appType,
		CreateUser:  "admin",
		UpdateUser:  "admin",
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
	valuesByte, _ := yaml.Marshal(charts.Values)
	appVersion := &types.AppVersion{
		PackageName:    charts.Name(),
		PackageVersion: charts.Metadata.Version,
		AppVersion:     charts.AppVersion(),
		From:           types.AppVersionFromImport,
		Values:         string(valuesByte),
		CreateUser:     "admin",
		CreateTime:     time.Now(),
		UpdateTime:     time.Now(),
	}
	chartByte, _ := ioutil.ReadFile(filePath)
	_, err = a.CreateStoreApp(chartByte, app, appVersion)
	if err != nil {
		klog.Errorf("create app %s-%s error: %s", app.Name, appVersion.PackageVersion, err)
	}
}
