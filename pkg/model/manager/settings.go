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

type SettingsManager struct {
	*CommonManager
}

func NewSettingsManager(db *gorm.DB) *SettingsManager {
	return &SettingsManager{
		CommonManager: NewCommonManager(nil, db, "", false),
	}
}

func (s *SettingsManager) Create(settings *types.Settings) (*types.Settings, error) {
	result := s.DB.Create(settings)
	if result.Error != nil {
		return nil, result.Error
	}
	return settings, nil
}

func (s *SettingsManager) Update(settings *types.Settings) (*types.Settings, error) {
	result := s.DB.Save(settings)
	if result.Error != nil {
		return nil, result.Error
	}
	return settings, nil
}

func (s *SettingsManager) Delete(settings *types.Settings) error {
	result := s.DB.Delete(settings)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *SettingsManager) Get(settingsId uint) (*types.Settings, error) {
	var settings types.Settings
	if err := s.DB.First(&settings, settingsId).Error; err != nil {
		return nil, err
	}
	return &settings, nil
}

func (s *SettingsManager) List() ([]types.Settings, error) {
	var settings []types.Settings
	result := s.DB.Find(&settings)
	if result.Error != nil {
		return nil, result.Error
	}
	return settings, nil
}
