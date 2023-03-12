package v1_1_3_b_add_ldap

import (
	"github.com/kubespace/kubespace/pkg/model/migrate/migration"
	v1_1_3_b "github.com/kubespace/kubespace/pkg/model/migrate/v1_1/v1_1_3_b_add_ldap"
	"gorm.io/gorm"
	"time"
)

var MigrateVersion = "v1.1.4_a"

func init() {
	migration.Register(&migration.Migration{
		Version:       MigrateVersion,
		ParentVersion: v1_1_3_b.MigrateVersion,
		MigrateFunc:   Migrate,
		Description:   "pipeline_stage_run表增加finish_time",
	})
}

type PipelineRunStage struct {
	FinishTime *time.Time `gorm:"autoCreateTime;" json:"finish_time"`
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(PipelineRunStage{})
}
