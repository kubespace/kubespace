package pipeline

import (
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

type ManagerPipeline struct {
	db *gorm.DB
}

func NewPipelineManager(db *gorm.DB) *ManagerPipeline {
	return &ManagerPipeline{db: db}
}

func (p *ManagerPipeline) CreatePipeline(pipeline *types.Pipeline, stages []*types.PipelineStage, triggers []*types.PipelineTrigger) (*types.Pipeline, error) {
	err := p.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(pipeline).Error; err != nil {
			return err
		}
		var prevStageId uint = 0
		for _, stage := range stages {
			stage.PipelineId = pipeline.ID
			stage.PrevStageId = prevStageId
			if err := tx.Create(stage).Error; err != nil {
				return err
			}
			prevStageId = stage.ID
		}
		for _, trigger := range triggers {
			trigger.PipelineId = pipeline.ID
			if err := tx.Create(trigger).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return pipeline, nil
}

func (p *ManagerPipeline) UpdatePipeline(pipeline *types.Pipeline, stages []*types.PipelineStage, triggers []*types.PipelineTrigger) (*types.Pipeline, error) {
	err := p.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(pipeline).Error; err != nil {
			return err
		}
		var prevStageId uint = 0
		oriStages, err := p.Stages(pipeline.ID)
		if err != nil {
			return err
		}
		// 删掉原阶段不存在的
		for _, stage := range oriStages {
			hasNew := false
			for _, newStage := range stages {
				if stage.ID == newStage.ID {
					hasNew = true
					break
				}
			}
			if !hasNew {
				if err = tx.Delete(stage).Error; err != nil {
					return err
				}
			}
		}
		for _, stage := range stages {
			stage.PipelineId = pipeline.ID
			stage.PrevStageId = prevStageId
			if stage.ID == 0 {
				if err = tx.Create(stage).Error; err != nil {
					return err
				}
			} else {
				if err = tx.Updates(stage).Error; err != nil {
					return err
				}
			}
			prevStageId = stage.ID
		}
		var updateIds []uint
		for _, trigger := range triggers {
			if trigger.ID != 0 {
				updateIds = append(updateIds, trigger.ID)
			}
		}
		delTx := tx.Where("pipeline_id=?", pipeline.ID)
		if updateIds != nil {
			delTx = delTx.Where("id not in ?", updateIds)
		}
		// 删除不在更新中的trigger
		if err = delTx.Debug().Delete(&types.PipelineTrigger{}).Error; err != nil {
			return err
		}
		for _, trigger := range triggers {
			if trigger.ID == 0 {
				if err = tx.Create(trigger).Error; err != nil {
					return err
				}
			} else {
				if err = tx.Updates(trigger).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return pipeline, nil
}

func (p *ManagerPipeline) Get(pipelineId uint) (*types.Pipeline, error) {
	var pipeline types.Pipeline
	if err := p.db.First(&pipeline, pipelineId).Error; err != nil {
		return nil, err
	}
	return &pipeline, nil
}

func (p *ManagerPipeline) List(workspaceId uint) ([]types.Pipeline, error) {
	var ps []types.Pipeline
	result := p.db.Where("workspace_id = ?", workspaceId).Find(&ps)
	if result.Error != nil {
		return nil, result.Error
	}
	return ps, nil
}

func (p *ManagerPipeline) Stages(pipelineId uint) ([]*types.PipelineStage, error) {
	var stages []types.PipelineStage
	if err := p.db.Where("pipeline_id = ?", pipelineId).Find(&stages).Error; err != nil {
		return nil, err
	}

	var seqStages []*types.PipelineStage
	prevStageId := uint(0)
	for {
		hasNext := false
		for i, s := range stages {
			if s.PrevStageId == prevStageId {
				seqStages = append(seqStages, &stages[i])
				prevStageId = s.ID
				hasNext = true
				break
			}
		}
		if !hasNext {
			break
		}
	}
	return seqStages, nil
}

func (p *ManagerPipeline) Delete(pipelineId uint) error {
	return p.db.Transaction(func(tx *gorm.DB) error {
		var pipelineRuns []types.PipelineRun
		if err := tx.Order("id desc").Where("pipeline_id = ?", pipelineId).Find(&pipelineRuns).Error; err != nil {
			return err
		}
		for _, pipelineRun := range pipelineRuns {
			var pipelineRunJobs []types.PipelineRunJob
			if err := tx.Where("pipeline_run_id = ?", pipelineRun.ID).Find(&pipelineRunJobs).Error; err != nil {
				return err
			}
			for _, runJob := range pipelineRunJobs {
				if err := tx.Delete(&types.PipelineRunJobLog{}, "job_run_id=?", runJob.ID).Error; err != nil {
					return err
				}
			}
			if err := tx.Delete(&types.PipelineRunJob{}, "pipeline_run_id=?", pipelineRun.ID).Error; err != nil {
				return err
			}
			if err := tx.Delete(&types.PipelineRunStage{}, "pipeline_run_id=?", pipelineRun.ID).Error; err != nil {
				return err
			}
		}
		if err := tx.Delete(&types.PipelineRun{}, "pipeline_id=?", pipelineId).Error; err != nil {
			return err
		}
		if err := tx.Delete(&types.PipelineStage{}, "pipeline_id=?", pipelineId).Error; err != nil {
			return err
		}
		if err := tx.Delete(&types.PipelineTrigger{}, "pipeline_id=?", pipelineId).Error; err != nil {
			return err
		}
		if err := tx.Delete(&types.Pipeline{}, "id=?", pipelineId).Error; err != nil {
			return err
		}
		return nil
	})
}
