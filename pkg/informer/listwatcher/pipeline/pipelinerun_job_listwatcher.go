package pipeline

import (
	"github.com/kubespace/kubespace/pkg/informer/listwatcher"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/config"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/storage"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"gorm.io/gorm"
)

const PipelineRunJobWatchKey = "kubespace:pipeline:run_job"

type PipelineRunJobWatchCondition struct {
	WithList   bool     `json:"with_list"`
	PipelineId uint     `json:"pipeline_id"`
	Id         uint     `json:"id"`
	StatusIn   []string `json:"status_in"`
}

type pipelineRunJobListWatcher struct {
	storage.Storage
	config    *config.ListWatcherConfig
	db        *gorm.DB
	condition *PipelineRunJobWatchCondition
}

func NewPipelineRunJobListWatcher(config *config.ListWatcherConfig, cond *PipelineRunJobWatchCondition) listwatcher.Interface {
	a := &pipelineRunJobListWatcher{
		config:    config,
		db:        config.DB,
		condition: cond,
	}
	var listFunc storage.ListFunc
	if cond != nil && cond.WithList {
		listFunc = a.List
	}
	a.Storage = config.NewStorage(PipelineRunJobWatchKey, listFunc, a.Filter, nil, &types.PipelineRun{})
	return a
}

func (p *pipelineRunJobListWatcher) Filter(obj interface{}) bool {
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

func (p *pipelineRunJobListWatcher) List() ([]interface{}, error) {
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
