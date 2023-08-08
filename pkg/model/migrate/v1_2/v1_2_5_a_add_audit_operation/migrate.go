package v1_2_5_a_add_audit_operation

import (
	"github.com/kubespace/kubespace/pkg/model/migrate/migration"
	v1_1_6_a "github.com/kubespace/kubespace/pkg/model/migrate/v1_1/v1_1_6_a_pipeline_trigger_event_add_user"
	"gorm.io/gorm"
	"time"
)

var MigrateVersion = "v1.2.5_a"

func init() {
	migration.Register(&migration.Migration{
		Version:       MigrateVersion,
		ParentVersion: v1_1_6_a.MigrateVersion,
		MigrateFunc:   Migrate,
		Description:   "增加操作审计表",
	})
}

// AuditOperate 操作审计
type AuditOperate struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	Operator      string      `gorm:"size:255;not null;comment:操作人" json:"operator"`
	Operation     string      `gorm:"size:255;not null;comment:操作类型(create/delete/update)" json:"operation"`
	OperateDetail string      `gorm:"type:text;not null;comment:操作详情" json:"operate_detail"`
	Scope         string      `gorm:"size:50;not null;comment:操作对象所属范围(project/pipeline/cluster/platform)" json:"scope"`
	ScopeId       uint        `gorm:"not null;comment:所属范围id" json:"scope_id"`
	ScopeName     string      `gorm:"size:512;not null;comment:所属范围名称" json:"scope_name"`
	Namespace     string      `gorm:"size:512;not null;comment:集群命名空间" json:"namespace"`
	ResourceId    uint        `gorm:"not null;comment:操作资源id" json:"resource_id"`
	ResourceType  string      `gorm:"size:255;not null;comment:资源类型" json:"resource_type"`
	ResourceName  string      `gorm:"size:512;not null;comment:资源名称" json:"resource_name"`
	Code          string      `gorm:"size:255;not null;comment:操作返回码" json:"code"`
	Message       string      `gorm:"type:text;not null;comment:操作返回信息" json:"message"`
	OperateData   interface{} `gorm:"type:json;comment:操作数据，操作参数以及返回数据" json:"operate_data"`
	CreateTime    time.Time   `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
}

func (s AuditOperate) TableName() string {
	return "audit_operate"
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&AuditOperate{})
}
