package types

import (
	"gorm.io/datatypes"
	"time"
)

const (
	AuditOperationCreate  = "创建"
	AuditOperationDelete  = "删除"
	AuditOperationUpdate  = "更新"
	AuditOperationInstall = "安装"
	AuditOperationUpgrade = "升级"
	// AuditOperationDestroy 应用销毁
	AuditOperationDestroy = "销毁"
	AuditOperationPatch   = "Patch"
	AuditOperationApply   = "Apply"
	AuditOperationClone   = "克隆"
	AuditOperationRelease = "发布"
	AuditOperationImport  = "导入"
)
const (
	AuditResourceApp        = "应用"
	AuditResourceAppVersion = "应用版本"
	AuditResourceProject    = "工作空间"

	AuditResourceAppStore = "应用商店"

	AuditResourceCluster          = "集群"
	AuditResourceClusterComponent = "集群组件"

	AuditResourcePipeSpace        = "流水线空间"
	AuditResourcePipeline         = "流水线"
	AuditResourcePipelineResource = "流水线资源"

	AuditResourcePlatformSecret   = "平台密钥"
	AuditResourcePlatformRegistry = "镜像仓库"
	AuditResourcePlatformSpacelet = "Spacelet"
	AuditResourcePlatformUser     = "用户"

	AuditResourcePermission = "权限"
)

// AuditOperate 操作审计
type AuditOperate struct {
	ID                   uint           `gorm:"primaryKey" json:"id"`
	Operator             string         `gorm:"size:255;not null;comment:操作人" json:"operator"`
	Operation            string         `gorm:"size:255;not null;comment:操作类型(create/delete/update)" json:"operation"`
	OperateDetail        string         `gorm:"type:text;not null;comment:操作详情" json:"operate_detail"`
	Scope                string         `gorm:"size:50;not null;comment:操作对象所属范围(project/pipeline/cluster/platform)" json:"scope"`
	ScopeId              uint           `gorm:"not null;comment:所属范围id" json:"scope_id"`
	ScopeName            string         `gorm:"size:512;not null;comment:所属范围名称" json:"scope_name"`
	Namespace            string         `gorm:"size:512;not null;comment:集群命名空间" json:"namespace"`
	ResourceId           uint           `gorm:"not null;comment:操作资源id" json:"resource_id"`
	ResourceType         string         `gorm:"size:255;not null;comment:资源类型" json:"resource_type"`
	ResourceName         string         `gorm:"size:512;not null;comment:资源名称" json:"resource_name"`
	Code                 string         `gorm:"size:255;not null;comment:操作返回码" json:"code"`
	Message              string         `gorm:"type:text;not null;comment:操作返回信息" json:"message"`
	OperateData          datatypes.JSON `gorm:"type:json;comment:操作数据，操作参数以及返回数据" json:"operate_data"`
	OperateDataInterface interface{}    `gorm:"-" json:"-"`
	Ip                   string         `gorm:"size:512;comment:来源ip" json:"ip"`
	CreateTime           time.Time      `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
}

func (s AuditOperate) TableName() string {
	return "audit_operate"
}
