package types

import (
	"database/sql/driver"
	"github.com/kubespace/kubespace/pkg/core/db"
	"time"
)

const (
	// SpaceletStatusOnline spacele状态在线
	SpaceletStatusOnline = "online"
	// SpaceletStatusOffline spacelet状态不在线
	SpaceletStatusOffline = "offline"
)

// Spacelet 流水线执行代理节点，spacelet启动时会进行注册
type Spacelet struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Hostname   string         `gorm:"size:255;not null;" json:"hostname"`
	HostIp     string         `gorm:"size:255;not null;uniqueIndex:HostPortUnique" json:"hostip"`
	Port       int            `gorm:"not null;uniqueIndex:HostPortUnique" json:"port"`
	Labels     SpaceletLabels `gorm:"type:json;comment:spacelet标签" json:"labels"`
	Token      string         `gorm:"size:255;not null;" json:"token,omitempty"`
	Status     string         `gorm:"size:50;not null" json:"status"`
	CreateTime time.Time      `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime time.Time      `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}

type SpaceletLabels map[string]string

func (m *SpaceletLabels) Scan(value interface{}) error {
	return db.Scan(value, m)
}

// Value return json value, implement driver.Valuer interface
func (m SpaceletLabels) Value() (driver.Value, error) {
	return db.Value(m)
}

func (s Spacelet) TableName() string {
	return "spacelet"
}
