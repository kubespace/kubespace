package spacelet

import (
	"github.com/kubespace/kubespace/pkg/informer/listwatcher"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/config"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/storage"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

const SpaceletWatchKey = "kubespace:spacelet:spacelet"

// SpaceletWatchCondition PipelineTrigger监听条件
type SpaceletWatchCondition struct {
	Status string
}

type spaceletListWatcher struct {
	storage.Storage
	config    *config.ListWatcherConfig
	db        *gorm.DB
	condition *SpaceletWatchCondition
}

func NewSpaceletListWatcher(config *config.ListWatcherConfig, cond *SpaceletWatchCondition) listwatcher.Interface {
	a := &spaceletListWatcher{
		config:    config,
		db:        config.DB,
		condition: cond,
	}
	resync := 30
	a.Storage = config.NewStorage(SpaceletWatchKey, a.List, a.Filter, &resync, &types.PipelineTrigger{})
	return a
}

func (p *spaceletListWatcher) Filter(obj interface{}) bool {
	_, ok := obj.(types.Spacelet)
	if !ok {
		return false
	}
	return true
}

func (p *spaceletListWatcher) List() ([]interface{}, error) {
	var spacelets []types.Spacelet
	var tx = p.db
	if p.condition.Status != "" {
		tx = tx.Where("status = ?", p.condition.Status)
	}
	if err := tx.Find(&spacelets).Error; err != nil {
		return nil, err
	}
	var objs []interface{}
	for i := range spacelets {
		objs = append(objs, spacelets[i])
	}
	return objs, nil
}
