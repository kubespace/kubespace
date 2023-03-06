package v1_1_3_b_add_ldap

import (
	"github.com/kubespace/kubespace/pkg/model/migrate/migration"
	v1_1_3_a "github.com/kubespace/kubespace/pkg/model/migrate/v1_1/v1_1_3_a_add_spacelet"
	"gorm.io/gorm"
)

var MigrateVersion = "v1.1.3_b"

func init() {
	migration.Register(&migration.Migration{
		Version:       MigrateVersion,
		ParentVersion: v1_1_3_a.MigrateVersion,
		MigrateFunc:   Migrate,
		Description:   "增加ldap服务表",
	})
}

type Ldap struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"size:64;not null" json:"name"`
	Enable      string `gorm:"size:16;not null" json:"enable"`
	Url         string `gorm:"size:255;not null" json:"url"`
	MaxConn     int    `gorm:"type:uint" json:"max_conn"`
	BaseDN      string `gorm:"size:255;not null" json:"base_dn"`
	AdminDN     string `gorm:"size:64;not null" json:"admin_dn"`
	AdminDNPass string `gorm:"size:64;not null" json:"admin_dn_pass"`
}

func (Ldap) TableName() string {
	return "ldap"
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(Ldap{})
}
