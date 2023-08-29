package types

import (
	"time"
)

const (
	SettingsSecretTypePassword = "password"
	SettingsSecretTypeKey      = "key"
	SettingsSecretTypeToken    = "token"
)

// Secret 用于所有地方的密钥，比如git、镜像仓库等
type Secret struct {
	Type        string `json:"type"`
	User        string `json:"user"`
	Password    string `json:"password"`
	PrivateKey  string `json:"private_key"`
	AccessToken string `json:"access_token"`
}

type SettingsSecret struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Description string    `gorm:"size:2000;" json:"description"`
	Type        string    `gorm:"size:50;not null" json:"type"`
	User        string    `gorm:"size:255;" json:"user"`
	Password    string    `gorm:"size:255;" json:"-"`
	PrivateKey  string    `gorm:"size:5000;" json:"-"`
	AccessToken string    `gorm:"size:2000;" json:"-"`
	CreateUser  string    `gorm:"size:255;not null" json:"create_user"`
	UpdateUser  string    `gorm:"size:255;not null" json:"update_user"`
	CreateTime  time.Time `gorm:"not null;autoCreateTime" json:"create_time"`
	UpdateTime  time.Time `gorm:"not null;autoUpdateTime" json:"update_time"`
}

func (s *SettingsSecret) GetSecret() *Secret {
	return &Secret{
		Type:        s.Type,
		User:        s.User,
		Password:    s.Password,
		PrivateKey:  s.PrivateKey,
		AccessToken: s.AccessToken,
	}
}

type ImageRegistry struct {
	Registry string `json:"registry"`
	User     string `json:"user"`
	Password string `json:"password"`
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

func (s *SettingsImageRegistry) GetImageRegistry() *ImageRegistry {
	return &ImageRegistry{
		Registry: s.Registry,
		User:     s.User,
		Password: s.Password,
	}
}
