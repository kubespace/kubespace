package pipeline

import (
	"github.com/kubespace/kubespace/pkg/informer/listwatcher"
	listwatcherconfig "github.com/kubespace/kubespace/pkg/informer/listwatcher/config"
	pipelinelistwatcher "github.com/kubespace/kubespace/pkg/informer/listwatcher/pipeline"
	"gorm.io/gorm"
)

type PipelineTriggerEventManager struct {
	db                              *gorm.DB
	pipelineTriggerEventListWatcher listwatcher.Interface
}

func NewPipelineTriggerEventManager(db *gorm.DB, listwatcherConfig *listwatcherconfig.ListWatcherConfig) *PipelineTriggerEventManager {
	return &PipelineTriggerEventManager{
		db:                              db,
		pipelineTriggerEventListWatcher: pipelinelistwatcher.NewPipelineTriggerEventListWatcher(listwatcherConfig, nil),
	}
}
