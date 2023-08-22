package types

import "time"

type Project struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Description string    `gorm:"size:2000;not null;" json:"description"`
	ClusterId   string    `gorm:"size:255;not null;" json:"cluster_id"`
	Namespace   string    `gorm:"size:255;not null;" json:"namespace"`
	Owner       string    `gorm:"size:255;not null" json:"owner"`
	CreateUser  string    `gorm:"size:255;not null" json:"create_user"`
	UpdateUser  string    `gorm:"size:255;not null" json:"update_user"`
	CreateTime  time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime  time.Time `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}
