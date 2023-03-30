package pipeline

import (
	"github.com/kubespace/kubespace/pkg/informer/listwatcher"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/config"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/storage"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
	"time"
)

const PipelineTriggerWatchKey = "kubespace:pipeline:trigger"

// PipelineTriggerWatchCondition PipelineTrigger监听条件
type PipelineTriggerWatchCondition struct {
	Triggered bool
}

type pipelineTriggerListWatcher struct {
	storage.Storage
	config    *config.ListWatcherConfig
	db        *gorm.DB
	condition *PipelineTriggerWatchCondition
}

func NewPipelineTriggerListWatcher(config *config.ListWatcherConfig, cond *PipelineTriggerWatchCondition) listwatcher.Interface {
	a := &pipelineTriggerListWatcher{
		config:    config,
		db:        config.DB,
		condition: cond,
	}
	a.Storage = config.NewStorage(PipelineTriggerWatchKey, a.List, a.Filter, nil, &types.PipelineTrigger{})
	return a
}

func (p *pipelineTriggerListWatcher) Filter(obj interface{}) bool {
	_, ok := obj.(types.PipelineTrigger)
	if !ok {
		return false
	}
	return true
}

func (p *pipelineTriggerListWatcher) List() ([]interface{}, error) {
	var pipelineTriggers []types.PipelineTrigger
	var tx = p.db
	if p.condition.Triggered {
		tx = tx.Where("trigger_time >= ", time.Now())
	}
	if err := tx.Find(&pipelineTriggers).Error; err != nil {
		return nil, err
	}
	var objs []interface{}
	for i := range pipelineTriggers {
		objs = append(objs, pipelineTriggers[i])
	}
	return objs, nil
}
