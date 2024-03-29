package pipeline

import (
	"github.com/kubespace/kubespace/pkg/informer/listwatcher"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/config"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/storage"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

const PipelineTriggerEventWatchKey = "kubespace:pipeline:trigger_event"

// PipelineTriggerEventWatchCondition PipelineTriggerEvent监听条件
type PipelineTriggerEventWatchCondition struct {
	Status string
}

type pipelineTriggerEventListWatcher struct {
	storage.Storage
	config    *config.ListWatcherConfig
	db        *gorm.DB
	condition *PipelineTriggerEventWatchCondition
}

func NewPipelineTriggerEventListWatcher(config *config.ListWatcherConfig, cond *PipelineTriggerEventWatchCondition) listwatcher.Interface {
	a := &pipelineTriggerEventListWatcher{
		config:    config,
		db:        config.DB,
		condition: cond,
	}
	a.Storage = config.NewStorage(PipelineTriggerEventWatchKey, a.List, a.Filter, nil, &types.PipelineTriggerEvent{})
	return a
}

func (p *pipelineTriggerEventListWatcher) Filter(obj interface{}) bool {
	pipelineTriggerEvent, ok := obj.(types.PipelineTriggerEvent)
	if !ok {
		return false
	}
	if p.condition.Status != "" && pipelineTriggerEvent.Status != p.condition.Status {
		return false
	}
	return true
}

func (p *pipelineTriggerEventListWatcher) List() ([]interface{}, error) {
	var pipelineTriggerEvents []types.PipelineTriggerEvent
	var tx = p.db
	if p.condition.Status != "" {
		tx = tx.Where("status = ?", p.condition.Status)
	}
	if err := tx.Find(&pipelineTriggerEvents).Error; err != nil {
		return nil, err
	}
	var objs []interface{}
	for i := range pipelineTriggerEvents {
		objs = append(objs, pipelineTriggerEvents[i])
	}
	return objs, nil
}
