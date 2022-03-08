package project

import (
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

type AppStoreManager struct {
	*gorm.DB
	*AppVersionManager
}

func NewAppStoreManager(versionManager *AppVersionManager, db *gorm.DB) *AppStoreManager {
	return &AppStoreManager{DB: db, AppVersionManager: versionManager}
}

func (a *AppManager) CreateStoreApp(chartBytes []byte, storeApp *types.AppStore, appVersion *types.AppVersion) (*types.AppStore, error) {
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
