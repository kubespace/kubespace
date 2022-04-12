package types

import (
	"time"
)

const (
	SettingsSecretTypePassword = "password"
	SettingsSecretTypeKey      = "key"
	SettingsSecretTypeToken    = "token"
)

type SettingsSecret struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Description string    `gorm:"size:2000;" json:"description"`
	Type        string    `gorm:"size:50;not null" json:"type"`
	User        string    `gorm:"size:255;" json:"user"`
	Password    string    `gorm:"size:255;" json:"password"`
	PrivateKey  string    `gorm:"size:2000;" json:"private_key"`
	AccessToken string    `gorm:"size:2000;" json:"access_token"`
	CreateUser  string    `gorm:"size:255;not null" json:"create_user"`
	UpdateUser  string    `gorm:"size:255;not null" json:"update_user"`
	CreateTime  time.Time `gorm:"not null;autoCreateTime" json:"create_time"`
	UpdateTime  time.Time `gorm:"not null;autoUpdateTime" json:"update_time"`
}

type SettingsImageRegistry struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Registry   string    `gorm:"size:255;not null;uniqueIndex" json:"registry"`
	User       string    `gorm:"size:255;not null;" json:"user"`
	Password   string    `gorm:"size:255;not null;" json:"password"`
	CreateUser string    `gorm:"size:255;not null" json:"create_user"`
	UpdateUser string    `gorm:"size:255;not null" json:"update_user"`
	CreateTime time.Time `gorm:"not null;autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"not null;autoUpdateTime" json:"update_time"`
}
