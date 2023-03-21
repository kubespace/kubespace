package pipeline

import "gorm.io/gorm"

type PipelineTriggerManager struct {
	db *gorm.DB
}

func NewPipelineTriggerManager(db *gorm.DB) *PipelineTriggerManager {
	return &PipelineTriggerManager{db: db}
}
