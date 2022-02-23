package project

import (
	"errors"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

type AppManager struct {
	*gorm.DB
}

func NewAppManager(db *gorm.DB) *AppManager {
	return &AppManager{DB: db}
}

func (a *AppManager) Create(app *types.ProjectApp, appVersion *types.AppVersion) (*types.ProjectApp, error) {
	err := a.DB.Transaction(func(tx *gorm.DB) error {
		if app.ID == 0 {
			if err := tx.Create(app).Error; err != nil {
				return err
			}
		}
		appVersion.ProjectAppId = app.ID
		if err := tx.Create(appVersion).Error; err != nil {
			return err
		}
		if app.Status == types.AppStatusUninstall {
			app.AppVersionId = appVersion.ID
			if err := tx.Save(app).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (a *AppManager) GetByName(projectId uint, name string) (*types.ProjectApp, error) {
	var app types.ProjectApp
	err := a.DB.First(&app, "project_id = ? and name = ?", projectId, name).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &app, nil
}

func (a *AppManager) GetAppVersion(appId uint, packageName, packageVersion string) (*types.AppVersion, error) {
	var version types.AppVersion
	err := a.DB.First(&version, "project_app_id = ? and package_name = ? and package_version = ?", appId, packageName, packageVersion).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &version, nil
}
