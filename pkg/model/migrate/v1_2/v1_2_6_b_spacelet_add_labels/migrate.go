package v1_2_6_b_spacelet_add_labels

import (
	"github.com/kubespace/kubespace/pkg/model/migrate/migration"
	v1_2_6_a "github.com/kubespace/kubespace/pkg/model/migrate/v1_2/v1_2_6_a_update_app_scope"
	"gorm.io/gorm"
)

var MigrateVersion = "v1.2.6_b"

func init() {
	migration.Register(&migration.Migration{
		Version:       MigrateVersion,
		ParentVersion: v1_2_6_a.MigrateVersion,
		MigrateFunc:   Migrate,
		Description:   "Spacelet增加labels字段,pipelineRunJob增加schedulePolicy字段",
	})
}

// Spacelet 流水线执行代理节点，spacelet启动时会进行注册
type Spacelet struct {
	Labels interface{} `gorm:"type:json;comment:spacelet标签" json:"labels"`
}

func (s Spacelet) TableName() string {
	return "spacelet"
}

type PipelineRunJob struct {
	SchedulePolicy interface{} `gorm:"type:json;comment:任务执行时调度到spacelet策略" json:"schedule_policy"`
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Spacelet{}, &PipelineRunJob{})
}
