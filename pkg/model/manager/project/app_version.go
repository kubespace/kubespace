package project

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
	"os"
	"strconv"
	"time"
)

type AppVersionManager struct {
	*gorm.DB
}

func NewAppVersionManager(db *gorm.DB) *AppVersionManager {
	return &AppVersionManager{DB: db}
}

func (v *AppVersionManager) NewPackageFilenameFromNameVersion(name string, version string) string {
	prefix := strconv.FormatInt(time.Now().Unix(), 10)
	filename := fmt.Sprintf("%s/%s-%s.tgz", prefix, name, version)
	return filename
}

func (v *AppVersionManager) CreateAppVersion(chartFilePath string, scope string, scopeId uint, appVersion *types.AppVersion) (*types.AppVersion, error) {
	content, err := os.ReadFile(chartFilePath)
	if err != nil {
		return nil, err
	}
	err = v.DB.Transaction(func(tx *gorm.DB) error {
		path := v.NewPackageFilenameFromNameVersion(appVersion.PackageName, appVersion.PackageVersion)
		appVersion.Scope = scope
		appVersion.ScopeId = scopeId
		appVersion.ChartPath = path
		if err = v.DB.Create(appVersion).Error; err != nil {
			return err
		}
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

func (v *AppVersionManager) GetAppVersion(appVersionId uint) (*types.AppVersion, error) {
	var appVersion types.AppVersion
	if err := v.DB.First(&appVersion, "id = ?", appVersionId).Error; err != nil {
		return nil, err
	}
	return &appVersion, nil
}

func (v *AppVersionManager) GetAppVersionChart(chartPath string) (*types.AppVersionChart, error) {
	var appVersionChart types.AppVersionChart
	if err := v.DB.First(&appVersionChart, "path = ?", chartPath).Error; err != nil {
		return nil, err
	}
	return &appVersionChart, nil
}

func (v *AppVersionManager) UpdateAppVersion(appVersion *types.AppVersion, columns ...string) error {
	appVersion.UpdateTime = time.Now()
	if len(columns) == 0 {
		columns = []string{"*"}
	} else {
		columns = append(columns, "update_time")
	}
	if err := v.DB.Model(appVersion).Select(columns).Updates(*appVersion).Error; err != nil {
		return err
	}
	return nil
}

func (v *AppVersionManager) ListAppVersions(scope string, scopeId uint) (*[]types.AppVersion, error) {
	var appVersions []types.AppVersion
	var err error
	if err = v.DB.Where("scope = ? and scope_id = ?", scope, scopeId).Order("create_time desc").Find(&appVersions).Error; err != nil {
		return nil, err
	}
	return &appVersions, nil
}
