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

const SettingsTypeImageRegistry = "image_registry"

type Settings struct {
	ID         uint        `gorm:"primaryKey" json:"id"`
	Type       string      `gorm:"size:255;not null;uniqueIndex:ScopeTypeKey" json:"type"`
	Scope      string      `gorm:"size:255;not null;uniqueIndex:ScopeTypeKey" json:"scope"`
	ScopeId    string      `gorm:"size:255;not null;uniqueIndex:ScopeTypeKey" json:"scope_id"`
	Key        string      `gorm:"size:2000;not null;uniqueIndex:ScopeTypeKey" json:"key"`
	Value      interface{} `gorm:"type:json;not null;" json:"value"`
	CreateUser string      `gorm:"size:255;not null" json:"create_user"`
	UpdateUser string      `gorm:"size:255;not null" json:"update_user"`
	CreateTime time.Time   `gorm:"not null;autoCreateTime" json:"create_time"`
	UpdateTime time.Time   `gorm:"not null;autoUpdateTime" json:"update_time"`
}

type SettingsImageRegistry struct {
	Registry string `json:"registry"`
	User     string `json:"user"`
	Password string `json:"password"`
}
