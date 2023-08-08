package v1_1_6_a_pipeline_trigger_event_add_user

import (
	"github.com/kubespace/kubespace/pkg/model/migrate/migration"
	v1_1_5_b "github.com/kubespace/kubespace/pkg/model/migrate/v1_1/v1_1_5_b_pipeline_triggers_and_code_cache"
	"gorm.io/gorm"
)

var MigrateVersion = "v1.1.6_a"

func init() {
	migration.Register(&migration.Migration{
		Version:       MigrateVersion,
		ParentVersion: v1_1_5_b.MigrateVersion,
		MigrateFunc:   Migrate,
		Description:   "PipelineTriggerEvent增加triggerUser触发人字段",
	})
}

// PipelineTriggerEvent 根据流水线触发配置，当触发条件达到时生成触发事件，根据事件生成新的流水线构建任务
type PipelineTriggerEvent struct {
	TriggerUser string `gorm:"size:255;" json:"trigger_user"`
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&PipelineTriggerEvent{})
}
