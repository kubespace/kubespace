package manager

import (
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

type SettingsSecretManager struct {
	*CommonManager
}

func NewSettingsSecretManager(db *gorm.DB) *SettingsSecretManager {
	return &SettingsSecretManager{
		CommonManager: NewCommonManager(nil, db, "", false),
	}
}

func (s *SettingsSecretManager) Create(secret *types.SettingsSecret) (*types.SettingsSecret, error) {
	result := s.DB.Create(secret)
	if result.Error != nil {
		return nil, result.Error
	}
	return secret, nil
}

func (s *SettingsSecretManager) Update(secret *types.SettingsSecret) (*types.SettingsSecret, error) {
	result := s.DB.Save(secret)
	if result.Error != nil {
		return nil, result.Error
	}
	return secret, nil
}

func (s *SettingsSecretManager) Delete(secret *types.SettingsSecret) error {
	result := s.DB.Delete(secret)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *SettingsSecretManager) Get(secretId uint) (*types.SettingsSecret, error) {
	var secret types.SettingsSecret
	if err := s.DB.First(&secret, secretId).Error; err != nil {
		return nil, err
	}
	return &secret, nil
}

func (s *SettingsSecretManager) List() ([]types.SettingsSecret, error) {
	var secrets []types.SettingsSecret
	result := s.DB.Find(&secrets)
	if result.Error != nil {
		return nil, result.Error
	}
	return secrets, nil
}
