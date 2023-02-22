package v1_1_2_a_pipeline_workspace_chg_code

import (
	"github.com/kubespace/kubespace/pkg/model/migrate/migration"
	v1_1_2_a "github.com/kubespace/kubespace/pkg/model/migrate/v1_1/v1_1_2_a_pipeline_workspace_chg_code"
	"gorm.io/gorm"
	"time"
)

var MigrateVersion = "v1.1.3_a"

func init() {
	migration.Register(&migration.Migration{
		Version:       MigrateVersion,
		ParentVersion: v1_1_2_a.MigrateVersion,
		MigrateFunc:   Migrate,
		Description:   "增加spacelet表，pipeline_job增加spacelet_id",
	})
}

type Spacelet struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Hostname   string    `gorm:"size:255;not null;" json:"hostname"`
	HostIp     string    `gorm:"size:255;not null;uniqueIndex:HostPortUnique" json:"hostip"`
	Port       int       `gorm:"not null;uniqueIndex:HostPortUnique" json:"port"`
	Token      string    `gorm:"size:255;not null;" json:"token"`
	Status     string    `gorm:"size:50;not null" json:"status"`
	CreateTime time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}

func (Spacelet) TableName() string {
	return "spacelet"
}

type PipelineRunJob struct {
	SpaceletId uint `gorm:""`
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(Spacelet{}, PipelineRunJob{})
}
