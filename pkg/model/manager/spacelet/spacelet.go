package spacelet

import (
	"errors"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

type SpaceletManager struct {
	DB *gorm.DB
}

func NewSpaceletManager(db *gorm.DB) *SpaceletManager {
	return &SpaceletManager{
		DB: db,
	}
}

func (s *SpaceletManager) Create(object *types.Spacelet) (*types.Spacelet, error) {
	if err := s.DB.Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (s *SpaceletManager) Update(id uint, object *types.Spacelet) error {
	return s.DB.Model(types.Spacelet{}).Where("id=?", id).Updates(object).Error
}

func (s *SpaceletManager) Delete(id uint) error {
	return s.DB.Delete(types.Spacelet{}, "id=?", id).Error
}

func (s *SpaceletManager) Get(id uint) (*types.Spacelet, error) {
	var object types.Spacelet
	if err := s.DB.First(&object, "id=?", id).Error; err != nil {
		return nil, err
	}
	return &object, nil
}

func (s *SpaceletManager) GetByIpPort(hostIp string, port int) (*types.Spacelet, error) {
	var object types.Spacelet
	if err := s.DB.First(&object, "host_ip=? and port=?", hostIp, port).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &object, nil
}

type SpaceletListCondition struct {
	Status string
}

func (s *SpaceletManager) List(cond *SpaceletListCondition) ([]*types.Spacelet, error) {
	var objects []*types.Spacelet
	tx := s.DB
	if cond.Status != "" {
		tx = tx.Where("status = ?", cond.Status)
	}
	if err := tx.Find(&objects).Error; err != nil {
		return nil, err
	}
	return objects, nil
}
