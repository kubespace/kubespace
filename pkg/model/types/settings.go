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
	Type        string    `json:"type"`
	User        string    `json:"user"`
	Password    string    `json:"password"`
	PrivateKey  string    `json:"private_key"`
	AccessToken string    `json:"access_token"`
	CreateUser  string    `gorm:"size:255;not null" json:"create_user"`
	UpdateUser  string    `gorm:"size:255;not null" json:"update_user"`
	CreateTime  time.Time `gorm:"not null;autoCreateTime" json:"create_time"`
	UpdateTime  time.Time `gorm:"not null;autoUpdateTime" json:"update_time"`
}

type SettingsImageRegistry struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Registry   string    `gorm:"size:255;not null;uniqueIndex:ScopeTypeKey" json:"registry"`
	User       string    `gorm:"size:255;not null;uniqueIndex:ScopeTypeKey" json:"user"`
	Password   string    `gorm:"size:255;not null;uniqueIndex:ScopeTypeKey" json:"password"`
	CreateUser string    `gorm:"size:255;not null" json:"create_user"`
	UpdateUser string    `gorm:"size:255;not null" json:"update_user"`
	CreateTime time.Time `gorm:"not null;autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"not null;autoUpdateTime" json:"update_time"`
}
