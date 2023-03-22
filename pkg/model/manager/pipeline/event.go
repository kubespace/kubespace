package pipeline

import (
	"github.com/kubespace/kubespace/pkg/informer/listwatcher"
	listwatcherconfig "github.com/kubespace/kubespace/pkg/informer/listwatcher/config"
	pipelinelistwatcher "github.com/kubespace/kubespace/pkg/informer/listwatcher/pipeline"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
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

func (r *PipelineTriggerEventManager) Create(event *types.PipelineTriggerEvent) error {
	if err := r.db.Omit("id").Create(event).Error; err != nil {
		return err
	}
	if err := r.pipelineTriggerEventListWatcher.Notify(event); err != nil {
		// 发送通知失败不报错，在controller端有list定时机制
		klog.Warningf("notify pipeline trigger event id=%d error: %s", event.ID, err.Error())
	}
	return nil
}

func (r *PipelineTriggerEventManager) Update(id uint, event *types.PipelineTriggerEvent) error {
	return r.db.Model(types.PipelineTriggerEvent{}).Where("id=?", id).Updates(event).Error
}
