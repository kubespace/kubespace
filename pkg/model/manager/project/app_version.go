package project

import (
	"errors"
	"fmt"
	"github.com/kubespace/kubespace/pkg/model/manager"
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

// CreateWithChartPath 通过chart本地存储创建
func (v *AppVersionManager) CreateWithChartPath(chartFilePath string, scope string, scopeId uint, appVersion *types.AppVersion) (*types.AppVersion, error) {
	content, err := os.ReadFile(chartFilePath)
	if err != nil {
		return nil, err
	}
	return v.CreateWithChartByte(content, scope, scopeId, appVersion)
}

// CreateWithChartByte 通过chart二进制内容创建
func (v *AppVersionManager) CreateWithChartByte(chartBytes []byte, scope string, scopeId uint, appVersion *types.AppVersion) (*types.AppVersion, error) {
	var err error
	err = v.DB.Transaction(func(tx *gorm.DB) error {
		path := v.NewPackageFilenameFromNameVersion(appVersion.PackageName, appVersion.PackageVersion)
		appVersion.Scope = scope
		appVersion.ScopeId = scopeId
		appVersion.ChartPath = path
		if err = tx.Create(appVersion).Error; err != nil {
			return err
		}
		var cnt int64
		if err = tx.Model(&types.AppVersion{}).Where("scope = ? and scope_id = ? ", scope, scopeId).Count(&cnt).Error; err != nil {
			return err
		}
		if cnt > 50 {
			if err = tx.Delete(&types.AppVersion{}, "scope = ? and scope_id = ? order by id limit ?", scope, scopeId, cnt-50).Error; err != nil {
				return err
			}
		}
		chart := &types.AppVersionChart{
			Path:       path,
			Content:    chartBytes,
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		}
		if err = tx.Create(chart).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return appVersion, nil
}

func (v *AppVersionManager) GetById(id uint, opfs ...manager.OptionFunc) (*types.AppVersion, error) {
	var appVersion types.AppVersion
	ops := manager.GetOptions(opfs)
	if err := v.DB.First(&appVersion, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) && ops.NotFoundReturnNil {
			return nil, nil
		}
		return nil, err
	}
	return &appVersion, nil
}

func (v *AppVersionManager) GetChart(chartPath string) (*types.AppVersionChart, error) {
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

func (v *AppVersionManager) List(scope string, scopeId uint) ([]*types.AppVersion, error) {
	var appVersions []*types.AppVersion
	var err error
	if err = v.DB.Where("scope = ? and scope_id = ?", scope, scopeId).Order("id desc").Find(&appVersions).Error; err != nil {
		return nil, err
	}
	return appVersions, nil
}

// Delete 删除应用版本以及该版本构建历史，如果该版本没有其他应用使用，则同时删除chart
func (v *AppVersionManager) Delete(id uint) error {
	var appVersion types.AppVersion
	if err := v.DB.First(&appVersion, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	var chartCount int64
	if err := v.DB.Model(&types.AppVersion{}).Where("chart_path=?", appVersion.ChartPath).Count(&chartCount).Error; err != nil {
		return err
	}
	if err := v.DB.Delete(&appVersion, "id = ?", id).Error; err != nil {
		return err
	}
	if chartCount == 1 {
		if err := v.DB.Delete(&types.AppVersionChart{}, "path=?", appVersion.ChartPath).Error; err != nil {
			return err
		}
	}
	if err := v.DB.Delete(&types.AppRevision{}, "app_version_id=?", appVersion.ID).Error; err != nil {
		return err
	}
	return nil
}

func (v *AppVersionManager) GetByPackageNameVersion(scope string, scopeId uint, packageName, packageVersion string) (*types.AppVersion, error) {
	var version types.AppVersion
	err := v.DB.First(&version, "scope = ? and scope_id = ? and package_name = ? and package_version = ?", scope, scopeId, packageName, packageVersion).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &version, nil
}
