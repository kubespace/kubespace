package types

import "time"

// Spacelet 流水线执行代理节点，spacelet启动时会进行注册
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

func (s Spacelet) TableName() string {
	return "spacelet"
}
