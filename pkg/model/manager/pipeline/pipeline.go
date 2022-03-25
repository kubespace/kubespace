package pipeline

import (
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

type ManagerPipeline struct {
	DB *gorm.DB
}

func NewPipelineManager(db *gorm.DB) *ManagerPipeline {
	return &ManagerPipeline{DB: db}
}

func (p *ManagerPipeline) CreatePipeline(pipeline *types.Pipeline, stages []*types.PipelineStage) (*types.Pipeline, error) {
	err := p.DB.Transaction(func(tx *gorm.DB) error {
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
		return nil
	})
	if err != nil {
		return nil, err
	}
	return pipeline, nil
}

func (p *ManagerPipeline) UpdatePipeline(pipeline *types.Pipeline, stages []*types.PipelineStage) (*types.Pipeline, error) {
	err := p.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(pipeline).Error; err != nil {
			return err
		}
		var prevStageId uint = 0
		oriStages, err := p.Stages(pipeline.ID)
		if err != nil {
			return err
		}
		for _, stage := range stages {
			stage.PipelineId = pipeline.ID
			stage.PrevStageId = prevStageId
			if stage.ID == 0 {
				if err = tx.Create(stage).Error; err != nil {
					return err
				}
			} else {
				if err = tx.Save(stage).Error; err != nil {
					return err
				}
			}
			prevStageId = stage.ID
		}
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
		return nil
	})
	if err != nil {
		return nil, err
	}
	return pipeline, nil
}

func (p *ManagerPipeline) Get(pipelineId uint) (*types.Pipeline, error) {
	var pipeline types.Pipeline
	if err := p.DB.First(&pipeline, pipelineId).Error; err != nil {
		return nil, err
	}
	return &pipeline, nil
}

func (p *ManagerPipeline) List(workspaceId uint) ([]types.Pipeline, error) {
	var ps []types.Pipeline
	result := p.DB.Where("workspace_id = ?", workspaceId).Find(&ps)
	if result.Error != nil {
		return nil, result.Error
	}
	return ps, nil
}

func (p *ManagerPipeline) Stages(pipelineId uint) ([]types.PipelineStage, error) {
	var stages []types.PipelineStage
	if err := p.DB.Where("pipeline_id = ?", pipelineId).Find(&stages).Error; err != nil {
		return nil, err
	}

	return stages, nil
}

func (p *ManagerPipeline) Delete(pipelineId uint) error {
	return p.DB.Transaction(func(tx *gorm.DB) error {
		var pipelineRuns []types.PipelineRun
		if err := tx.Order("id desc").Where("pipeline_id = ?", pipelineId).Find(&pipelineRuns).Error; err != nil {
			return err
		}
		for _, pipelineRun := range pipelineRuns {
			if err := tx.Delete(&types.PipelineRunJobLog{}, "pipeline_run_id=?", pipelineRun.ID).Error; err != nil {
				return err
			}
			if err := tx.Delete(&types.PipelineRunJob{}, "pipeline_run_id=?", pipelineRun.ID).Error; err != nil {
				return err
			}
			if err := tx.Delete(&types.PipelineStage{}, "pipeline_run_id=?", pipelineRun.ID).Error; err != nil {
				return err
			}
		}
		if err := tx.Delete(&types.PipelineRun{}, "pipeline_id=?", pipelineId).Error; err != nil {
			return err
		}
		if err := tx.Delete(&types.Pipeline{}, "id=?", pipelineId).Error; err != nil {
			return err
		}
		return nil
	})
}
