package v1_2_5_a_add_audit_operation

import (
	"github.com/kubespace/kubespace/pkg/model/migrate/migration"
	v1_2_5_a "github.com/kubespace/kubespace/pkg/model/migrate/v1_2/v1_2_5_a_add_audit_operation"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

var MigrateVersion = "v1.2.5_b"

func init() {
	migration.Register(&migration.Migration{
		Version:       MigrateVersion,
		ParentVersion: v1_2_5_a.MigrateVersion,
		MigrateFunc:   Migrate,
		Description:   "app表名修改",
	})
}

func Migrate(db *gorm.DB) error {
	if !db.Migrator().HasTable("apps") {
		if err := db.Migrator().RenameTable("project_apps", "apps"); err != nil {
			return err
		}
	}
	if !db.Migrator().HasTable("app_revisions") {
		if err := db.Migrator().RenameTable("project_app_revisions", "app_revisions"); err != nil {
			return err
		}
	}
	if !db.Migrator().HasColumn(&types.AppRevision{}, "app_id") {
		return db.Migrator().RenameColumn(&types.AppRevision{}, "project_app_id", "app_id")
	}
	return nil
}
