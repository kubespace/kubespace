package v1_1_2_a_pipeline_workspace_chg_code

import (
	"github.com/kubespace/kubespace/pkg/model/migrate/migration"
	v1_1_2_a "github.com/kubespace/kubespace/pkg/model/migrate/v1_1/v1_1_2_a_pipeline_workspace_chg_code"
	"gorm.io/gorm"
	"time"
)

var MigrateVersion = "v1.1.2_b"

func init() {
	migration.Register(&migration.Migration{
		Version:       MigrateVersion,
		ParentVersion: v1_1_2_a.MigrateVersion,
		MigrateFunc:   Migrate,
		Description:   "用户角色增加ParentScope/ParentScopeId/",
	})
}

type UserRole struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserId        uint      `gorm:"not null;uniqueIndex:idx_user_scope_id" json:"user_id"`
	Scope         string    `gorm:"size:50;not null;uniqueIndex:idx_user_scope_id" json:"scope"`
	ScopeId       uint      `gorm:"not null;uniqueIndex:idx_user_scope_id" json:"scope_id"`
	ScopeRegex    string    `gorm:"size:500;uniqueIndex:idx_user_scope_id" json:"scope_regex"`
	ParentScope   string    `gorm:"size:50;uniqueIndex:idx_user_scope_id" json:"parent_scope"`
	ParentScopeId uint      `gorm:"not null;uniqueIndex:idx_user_scope_id" json:"parent_scope_id"`
	Role          string    `gorm:"size:50;not null;" json:"role"`
	CreateTime    time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime    time.Time `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(UserRole{})
}
