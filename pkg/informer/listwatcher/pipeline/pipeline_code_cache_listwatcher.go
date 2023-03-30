package pipeline

import (
	"github.com/kubespace/kubespace/pkg/informer/listwatcher"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/config"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/storage"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

const PipelineCodeCacheWatchKey = "kubespace:pipeline:trigger_event"

// PipelineCodeCacheWatchCondition PipelineTriggerEvent监听条件
type PipelineCodeCacheWatchCondition struct {
	Status string
}

type pipelineCodeCacheListWatcher struct {
	storage.Storage
	config    *config.ListWatcherConfig
	db        *gorm.DB
	condition *PipelineCodeCacheWatchCondition
}

func NewPipelineCodeCacheListWatcher(config *config.ListWatcherConfig, cond *PipelineCodeCacheWatchCondition) listwatcher.Interface {
	a := &pipelineCodeCacheListWatcher{
		config:    config,
		db:        config.DB,
		condition: cond,
	}
	a.Storage = config.NewStorage(PipelineCodeCacheWatchKey, a.List, a.Filter, nil, &types.PipelineCodeCache{})
	return a
}

func (p *pipelineCodeCacheListWatcher) Filter(obj interface{}) bool {
	cache, ok := obj.(types.PipelineCodeCache)
	if !ok {
		return false
	}
	if p.condition.Status != "" && cache.Status != p.condition.Status {
		return false
	}
	return true
}

func (p *pipelineCodeCacheListWatcher) List() ([]interface{}, error) {
	var caches []types.PipelineCodeCache
	var tx = p.db
	if p.condition.Status != "" {
		tx = tx.Where("status = ?", p.condition.Status)
	}
	if err := tx.Find(&caches).Error; err != nil {
		return nil, err
	}
	var objs []interface{}
	for i := range caches {
		objs = append(objs, caches[i])
	}
	return objs, nil
}
