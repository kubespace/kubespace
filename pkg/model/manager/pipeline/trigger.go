package pipeline

import (
	"errors"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher"
	listwatcherconfig "github.com/kubespace/kubespace/pkg/informer/listwatcher/config"
	pipelinelistwatcher "github.com/kubespace/kubespace/pkg/informer/listwatcher/pipeline"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
	"time"
)

type PipelineTriggerManager struct {
	db                         *gorm.DB
	pipelineTriggerListWatcher listwatcher.Interface
}

func NewPipelineTriggerManager(db *gorm.DB, listwatcherConfig *listwatcherconfig.ListWatcherConfig) *PipelineTriggerManager {
	return &PipelineTriggerManager{
		db:                         db,
		pipelineTriggerListWatcher: pipelinelistwatcher.NewPipelineTriggerListWatcher(listwatcherConfig, nil),
	}
}

func (r *PipelineTriggerManager) Update(id uint, updates *types.PipelineTrigger) error {
	return r.db.Model(types.PipelineTrigger{}).Where("id=?", id).Updates(updates).Error
}

func (r *PipelineTriggerManager) List(cond *PipelineTriggerCondition) ([]*types.PipelineTrigger, error) {
	tx := r.conditionQuery(cond)
	if tx == nil {
		return nil, nil
	}
	var triggers []*types.PipelineTrigger
	if err := tx.Find(&triggers).Error; err != nil {
		return nil, err
	}
	return triggers, nil
}

type PipelineTriggerCondition struct {
	Id          uint
	WorkspaceId uint
	PipelineId  uint
}

func (r *PipelineTriggerManager) conditionQuery(cond *PipelineTriggerCondition) *gorm.DB {
	if cond == nil {
		return nil
	}
	tx := r.db.Model(types.PipelineTrigger{})
	if cond.Id != 0 {
		tx = tx.Where("id = ?", cond.Id)
		return tx
	}
	if cond.PipelineId != 0 {
		tx = tx.Where("pipeline_id = ?", cond.PipelineId)
		return tx
	}
	if cond.WorkspaceId != 0 {
		var pipelines []*types.Pipeline
		if err := r.db.Find(&pipelines, "workspace_id = ?", cond.WorkspaceId).Error; err != nil {
			return nil
		}
		var pipelineIds []uint
		for _, p := range pipelines {
			pipelineIds = append(pipelineIds, p.ID)
		}
		tx = tx.Where("pipeline_id in ?", pipelineIds)
		return tx
	}
	return nil
}

func (r *PipelineTriggerManager) UpdateTriggerTime(triggerTime *time.Time, condition *PipelineTriggerCondition) error {
	tx := r.conditionQuery(condition)
	if tx == nil {
		return nil
	}
	if err := tx.Updates(&types.PipelineTrigger{
		TriggerTime: triggerTime, UpdateTime: time.Now()}).Error; err != nil {
		return err
	}
	if triggerTime.Before(time.Now()) {
		// 如果触发时间小于当前时间，则发送通知给controller
		var triggers []*types.PipelineTrigger
		if err := tx.Find(&triggers).Error; err != nil {
			klog.Warningf("list triggers error: %s", err.Error())
			return nil
		}
		for _, t := range triggers {
			if err := r.pipelineTriggerListWatcher.Notify(t); err != nil {
				klog.Warningf("notify pipeline trigger id=%d error: %s", t.ID, err.Error())
			}
		}
	}
	return nil
}

func (r *PipelineTriggerManager) Get(id uint) (*types.PipelineTrigger, error) {
	var obj types.PipelineTrigger
	if err := r.db.First(&obj, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &obj, nil
}
