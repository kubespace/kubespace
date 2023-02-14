package v1_1_2_a_pipeline_workspace_chg_code

import (
	"database/sql/driver"
	"github.com/kubespace/kubespace/pkg/core/db"
	"github.com/kubespace/kubespace/pkg/model/migrate/migration"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
	"time"
)

var MigrateVersion = "v1.1.2_a"

func init() {
	migration.Register(&migration.Migration{
		Version:       MigrateVersion,
		ParentVersion: migration.FirstVersion,
		MigrateFunc:   Migrate,
		Description:   "流水线工作空间对代码信息结构调整，增加code字段，将原code字段数据进行迁移",
	})
}

type OriginPipelineWorkspace struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Description  string    `gorm:"type:text;" json:"description"`
	Type         string    `gorm:"size:20;not null" json:"type"`
	CodeType     string    `gorm:"size:255" json:"code_type"`
	CodeUrl      string    `gorm:"size:255" json:"code_url"`
	CodeSecretId uint      `json:"code_secret_id"`
	CreateUser   string    `gorm:"size:50;not null" json:"create_user"`
	UpdateUser   string    `gorm:"size:50;not null" json:"update_user"`
	CreateTime   time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime   time.Time `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}

func (OriginPipelineWorkspace) TableName() string {
	return "pipeline_workspaces"
}

type PipelineWorkspace struct {
	ID          uint                   `gorm:"primaryKey" json:"id"`
	Name        string                 `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Description string                 `gorm:"type:text;" json:"description"`
	Type        string                 `gorm:"size:20;not null" json:"type"`
	Code        *PipelineWorkspaceCode `gorm:"type:json" json:"code"`
	CreateUser  string                 `gorm:"size:50;not null" json:"create_user"`
	UpdateUser  string                 `gorm:"size:50;not null" json:"update_user"`
	CreateTime  time.Time              `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime  time.Time              `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}

type PipelineWorkspaceCode struct {
	Type     string `json:"type"`
	ApiUrl   string `json:"api_url"`
	CloneUrl string `json:"clone_url"`
	SecretId uint   `json:"secret_id"`
}

func (c *PipelineWorkspaceCode) Scan(value interface{}) error {
	return db.Scan(value, c)
}

func (c PipelineWorkspaceCode) Value() (driver.Value, error) {
	return db.Value(c)
}

func Migrate(db *gorm.DB) error {
	var oriPipelineWorkspaces []*OriginPipelineWorkspace
	if err := db.Find(&oriPipelineWorkspaces).Error; err != nil {
		return err
	}
	if err := db.AutoMigrate(&PipelineWorkspace{}); err != nil {
		return err
	}
	for _, pw := range oriPipelineWorkspaces {
		if pw.Type != types.WorkspaceTypeCode {
			continue
		}
		if err := db.Model(&PipelineWorkspace{}).Where("id=?", pw.ID).Updates(&PipelineWorkspace{
			Code: &PipelineWorkspaceCode{
				Type:     pw.CodeType,
				ApiUrl:   "",
				CloneUrl: pw.CodeUrl,
				SecretId: pw.CodeSecretId,
			},
		}).Error; err != nil {
			return err
		}
	}
	return nil
}
