package v1_2_6_a_update_app_scope

import (
	"github.com/kubespace/kubespace/pkg/model/migrate/migration"
	v1_2_5_b "github.com/kubespace/kubespace/pkg/model/migrate/v1_2/v1_2_5_b_alter_app_name"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

var MigrateVersion = "v1.2.6_a"

func init() {
	migration.Register(&migration.Migration{
		Version:       MigrateVersion,
		ParentVersion: v1_2_5_b.MigrateVersion,
		MigrateFunc:   Migrate,
		Description:   "App修改scope字段内容",
	})
}

func Migrate(db *gorm.DB) error {
	if err := db.Model(&types.App{}).Where("scope = ?", "project_app").Update("scope", types.ScopeProject).Error; err != nil {
		return err
	}
	if err := db.Model(&types.App{}).Where("scope = ?", "component").Update("scope", types.ScopeCluster).Error; err != nil {
		return err
	}
	if err := db.Model(&types.App{}).Where("scope = ?", "store_app").Update("scope", types.ScopeAppStore).Error; err != nil {
		return err
	}
	if err := db.Model(&types.AppVersion{}).Where("scope = ?", "project_app").Update("scope", types.ScopeProject).Error; err != nil {
		return err
	}
	if err := db.Model(&types.AppVersion{}).Where("scope = ?", "component").Update("scope", types.ScopeCluster).Error; err != nil {
		return err
	}
	if err := db.Model(&types.AppVersion{}).Where("scope = ?", "store_app").Update("scope", types.ScopeAppStore).Error; err != nil {
		return err
	}
	var apps []*types.App
	if err := db.Where("scope=?", types.ScopeCluster).Find(&apps).Error; err != nil {
		return err
	}
	for _, app := range apps {
		if err := db.Model(&types.AppVersion{}).Where(
			"scope=? and scope_id=?", types.ScopeProject, app.ID).Update("scope", types.ScopeCluster).Error; err != nil {
			return err
		}
	}
	return nil
}
