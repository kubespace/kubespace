package project

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
	"os"
	"time"
)

type AppVersionManager struct {
	*gorm.DB
}

func NewAppVersionManager(db *gorm.DB) *AppVersionManager {
	return &AppVersionManager{DB: db}
}

func (v *AppVersionManager) PackageFilenameFromNameVersion(name string, version string) string {
	filename := fmt.Sprintf("%s-%s", name, version)
	return filename
}

func (v *AppVersionManager) CreateAppVersion(chartFilePath string, scope string, scopeId uint, appVersion *types.AppVersion) (*types.AppVersion, error) {
	content, err := os.ReadFile(chartFilePath)
	if err != nil {
		return nil, err
	}
	err = v.DB.Transaction(func(tx *gorm.DB) error {
		appVersion.Scope = scope
		appVersion.ScopeId = scopeId
		if err = v.DB.Create(appVersion).Error; err != nil {
			return err
		}

		path := v.PackageFilenameFromNameVersion(appVersion.PackageName, appVersion.PackageVersion)
		chart := &types.AppVersionChart{
			Path:       path,
			Content:    content,
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		}
		if err = v.DB.Create(chart).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return appVersion, nil
}
