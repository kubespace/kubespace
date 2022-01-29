package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	SettingsSecretKindCode = "code"
	SettingsSecretKindRegistry = "registry"
	SettingsSecretKindSsh = "ssh"

	SettingsSecretTypePassword = "password"
	SettingsSecretTypeKey = "key"
	SettingsSecretTypeToken = "token"
)

type SettingsSecret struct {
	ID uint `gorm:"primaryKey" json:"id"`
	Name string `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Kind string `gorm:"size:255;not null" json:"kind"`
	Value SettingsSecretValue `gorm:"type:json;" json:"value"`
	CreateUser string `gorm:"size:255;not null" json:"create_user"`
	UpdateUser string `gorm:"size:255;not null" json:"update_user"`
	CreateTime time.Time               `gorm:"not null;autoCreateTime" json:"create_time"`
	UpdateTime time.Time               `gorm:"not null;autoUpdateTime" json:"update_time"`
}

type SettingsSecretValue struct {
	Type string `json:"type"`
	User string `json:"user"`
	Password string `json:"password"`
	PrivateKey string `json:"private_key"`
	AccessToken string `json:"access_token"`
}

func (s *SettingsSecretValue) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to convert to bytes:", value))
	}
	err := json.Unmarshal(bytes, s)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bytes: %s", string(bytes))
	}
	return nil
}

// Value return json value, implement driver.Valuer interface
func (s SettingsSecretValue) Value() (driver.Value, error) {
	bytes, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}
