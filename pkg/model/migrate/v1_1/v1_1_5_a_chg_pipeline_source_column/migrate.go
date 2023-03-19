package v1_1_3_b_add_ldap

import (
	"github.com/kubespace/kubespace/pkg/model/migrate/migration"
	v1_1_4_a "github.com/kubespace/kubespace/pkg/model/migrate/v1_1/v1_1_4_a_stagerun_add_finishtime"
	"gorm.io/gorm"
)

var MigrateVersion = "v1.1.5_a"

func init() {
	migration.Register(&migration.Migration{
		Version:       MigrateVersion,
		ParentVersion: v1_1_4_a.MigrateVersion,
		MigrateFunc:   Migrate,
		Description:   "修改pipeline原trigger字段为source",
	})
}

type Pipeline struct {
	Triggers interface{} `gorm:"type:json" json:"triggers"`
	Sources  interface{} `gorm:"type:json" json:"sources"`
}

func Migrate(db *gorm.DB) error {
	return db.Migrator().RenameColumn(&Pipeline{}, "triggers", "sources")
}
