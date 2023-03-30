package v1_1_3_b_add_ldap

import (
	"github.com/kubespace/kubespace/pkg/model/migrate/migration"
	v1_1_5_a "github.com/kubespace/kubespace/pkg/model/migrate/v1_1/v1_1_5_a_chg_pipeline_source_column"
	"gorm.io/gorm"
	"time"
)

var MigrateVersion = "v1.1.5_b"

func init() {
	migration.Register(&migration.Migration{
		Version:       MigrateVersion,
		ParentVersion: v1_1_5_a.MigrateVersion,
		MigrateFunc:   Migrate,
		Description:   "增加pipeline_trigger和pipeline_code_cache表",
	})
}

// PipelineTrigger 流水线触发配置
type PipelineTrigger struct {
	ID         uint        `gorm:"primaryKey" json:"id"`
	PipelineId uint        `gorm:"" json:"pipeline_id"`
	Type       string      `json:"type"`
	Config     interface{} `gorm:"type:json" json:"config"`
	// 下一次触发时间
	NextTriggerTime *time.Time `gorm:"" json:"next_trigger_time"`
	UpdateUser      string     `gorm:"size:50;not null" json:"update_user"`
	CreateTime      time.Time  `gorm:"not null;autoCreateTime" json:"create_time"`
	UpdateTime      time.Time  `gorm:"not null;autoUpdateTime" json:"update_time"`
}

// PipelineCodeCache 缓存代码分支的最新commit
type PipelineCodeCache struct {
	ID          uint `gorm:"primaryKey" json:"id"`
	WorkspaceId uint `gorm:"" json:"workspace_id"`
	// 缓存状态，open（开启）/close（关闭），当该代码空间下没有流水线代码触发配置时，状态为close，不需要缓存
	Status      string      `gorm:"size:50" json:"status"`
	CommitCache interface{} `gorm:"type:json" json:"branch_cache"`
	CreateTime  time.Time   `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime  time.Time   `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&PipelineTrigger{}, &PipelineCodeCache{})
}
