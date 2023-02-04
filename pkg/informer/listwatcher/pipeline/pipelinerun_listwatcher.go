package pipeline

import (
	"github.com/kubespace/kubespace/pkg/informer/listwatcher"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/config"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/storage"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"gorm.io/gorm"
)

const PipelineRunWatchKey = "kubespace:pipeline:run"

type PipelineRunWatchCondition struct {
	WithList   bool     `json:"with_list"`
	PipelineId uint     `json:"pipeline_id"`
	Id         uint     `json:"id"`
	StatusIn   []string `json:"status_in"`
}

type pipelineRunListWatcher struct {
	storage.Storage
	config    *config.ListWatcherConfig
	db        *gorm.DB
	condition *PipelineRunWatchCondition
}

func NewPipelineRunListWatcher(config *config.ListWatcherConfig, cond *PipelineRunWatchCondition) listwatcher.Interface {
	a := &pipelineRunListWatcher{
		config:    config,
		db:        config.DB,
		condition: cond,
	}
	var listFunc storage.ListFunc
	if cond != nil && cond.WithList {
		listFunc = a.List
	}
	a.Storage = config.NewStorage(PipelineRunWatchKey, listFunc, a.Filter, nil, &types.PipelineRun{})
	return a
}

func (p *pipelineRunListWatcher) Filter(obj interface{}) bool {
	pipelineRun, ok := obj.(types.PipelineRun)
	if !ok {
		return false
	}
	if p.condition.Id > 0 && pipelineRun.ID != p.condition.Id {
		return false
	}
	if p.condition.PipelineId > 0 && pipelineRun.PipelineId != p.condition.PipelineId {
		return false
	}
	if len(p.condition.StatusIn) > 0 && !utils.Contains(p.condition.StatusIn, pipelineRun.Status) {
		return false
	}
	return true
}

func (p *pipelineRunListWatcher) List() ([]interface{}, error) {
	var pipelineRuns []types.PipelineRun
	var tx = p.db
	if p.condition.Id > 0 {
		tx = tx.Where("id=?", p.condition.Id)
	}
	if p.condition.PipelineId > 0 {
		tx = tx.Where("pipeline_id=?", p.condition.PipelineId)
	}
	if len(p.condition.StatusIn) > 0 {
		tx = tx.Where("status in ?", p.condition.StatusIn)
	}
	if err := tx.Find(&pipelineRuns).Error; err != nil {
		return nil, err
	}
	var objs []interface{}
	for i := range pipelineRuns {
		objs = append(objs, pipelineRuns[i])
	}
	return objs, nil
}
