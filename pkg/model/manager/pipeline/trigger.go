package pipeline

import (
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

type PipelineTriggerManager struct {
	db *gorm.DB
}

func NewPipelineTriggerManager(db *gorm.DB) *PipelineTriggerManager {
	return &PipelineTriggerManager{db: db}
}

func (r *PipelineTriggerManager) Update(id uint, updates *types.PipelineTrigger) error {
	return r.db.Model(types.PipelineTrigger{}).Where("id=?", id).Updates(updates).Error
}
