package manager

import (
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

type LdapManager struct {
	DB *gorm.DB
}

func NewLdapManager(db *gorm.DB) *LdapManager {
	return &LdapManager{
		DB: db,
	}
}

func (l *LdapManager) Create(ldap *types.Ldap) (*types.Ldap, error) {
	result := l.DB.Create(ldap)
	if result.Error != nil {
		return nil, result.Error
	}
	return ldap, nil
}

func (l *LdapManager) Get(id uint) (*types.Ldap, error) {
	var ldap types.Ldap
	if err := l.DB.First(&ldap, id).Error; err != nil {
		return nil, err
	}
	return &ldap, nil
}

func (l *LdapManager) List() ([]types.Ldap, error) {
	var ldaps []types.Ldap

	result := l.DB.Find(&ldaps)
	if result.Error != nil {
		return nil, result.Error
	}
	return ldaps, nil
}

func (l *LdapManager) Update(ldap *types.Ldap) (*types.Ldap, error) {
	result := l.DB.Save(ldap)
	if result.Error != nil {
		return nil, result.Error
	}
	return ldap, nil
}

func (l *LdapManager) Delete(ldap *types.Ldap) error {
	result := l.DB.Delete(ldap)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
