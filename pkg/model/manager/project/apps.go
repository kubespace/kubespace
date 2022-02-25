package project

import (
	"errors"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

type AppManager struct {
	*gorm.DB
	*AppVersionManager
}

func NewAppManager(chartManager *AppVersionManager, db *gorm.DB) *AppManager {
	return &AppManager{DB: db, AppVersionManager: chartManager}
}

func (a *AppManager) CreateProjectApp(chartFilePath string, app *types.ProjectApp, appVersion *types.AppVersion) (*types.ProjectApp, error) {
	var err error
	err = a.DB.Transaction(func(tx *gorm.DB) error {
		if app.ID == 0 {
			if err = tx.Create(app).Error; err != nil {
				return err
			}
		}
		appVersion, err = a.AppVersionManager.CreateAppVersion(chartFilePath, types.AppVersionScopeProjectApp, app.ID, appVersion)
		if err != nil {
			return err
		}
		if app.Status == types.AppStatusUninstall {
			app.AppVersionId = appVersion.ID
			if err = tx.Save(app).Error; err != nil {
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

func (a *AppManager) GetAppVersion(scope string, scopeId uint, packageName, packageVersion string) (*types.AppVersion, error) {
	var version types.AppVersion
	err := a.DB.First(&version, "scope = ? and scope_id = ? and package_name = ? and package_version = ?", scope, scopeId, packageName, packageVersion).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &version, nil
}

func (a *AppManager) ListProjectApps(projectId uint) ([]*types.ProjectApp, error) {
	var apps []types.ProjectApp
	var err error
	if err = a.DB.Where("project_id = ?", projectId).Find(&apps).Error; err != nil {
		return nil, err
	}
	var rets []*types.ProjectApp
	for i, app := range apps {
		apps[i].AppVersion, err = a.AppVersionManager.GetAppVersion(app.AppVersionId)
		if err != nil {
			return nil, err
		}
		rets = append(rets, &apps[i])
	}
	return rets, nil
}

func (a *AppManager) GetProjectApp(appId uint) (*types.ProjectApp, error) {
	var app types.ProjectApp
	var err error
	if err = a.DB.First(&app, "id = ?", appId).Error; err != nil {
		return nil, err
	}
	if app.AppVersion, err = a.AppVersionManager.GetAppVersion(app.AppVersionId); err != nil {
		return nil, err
	}
	return &app, nil
}
