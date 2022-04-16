package pipeline

import (
	"github.com/kubespace/kubespace/pkg/model/manager"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

type WorkspaceManager struct {
	*manager.CommonManager
	PipelineManager *ManagerPipeline
}

func NewWorkspaceManager(db *gorm.DB, pipelineManager *ManagerPipeline) *WorkspaceManager {
	return &WorkspaceManager{
		CommonManager:   manager.NewCommonManager(nil, db, "", false),
		PipelineManager: pipelineManager,
	}
}

func (w *WorkspaceManager) Create(workspace *types.PipelineWorkspace, defaultPipelines []*types.Pipeline) (*types.PipelineWorkspace, error) {
	err := w.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(workspace).Error; err != nil {
			return err
		}
		for _, pipeline := range defaultPipelines {
			pipeline.WorkspaceId = workspace.ID
			if err := tx.Create(pipeline).Error; err != nil {
				return err
			}
			prevStageId := uint(0)
			for _, stage := range pipeline.Stages {
				stage.PipelineId = pipeline.ID
				stage.PrevStageId = prevStageId
				if err := tx.Create(stage).Error; err != nil {
					return err
				}
				prevStageId = stage.ID
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return workspace, nil
}

func (w *WorkspaceManager) Get(workspaceId uint) (*types.PipelineWorkspace, error) {
	var ws types.PipelineWorkspace
	if err := w.DB.First(&ws, workspaceId).Error; err != nil {
		return nil, err
	}
	return &ws, nil
}

func (w *WorkspaceManager) List() ([]types.PipelineWorkspace, error) {
	var ws []types.PipelineWorkspace
	result := w.DB.Find(&ws)
	if result.Error != nil {
		return nil, result.Error
	}
	return ws, nil
}

func (w *WorkspaceManager) Delete(workspace *types.PipelineWorkspace) error {
	var pipelines []types.Pipeline
	if err := w.DB.Where("workspace_id = ?", workspace.ID).Find(&pipelines).Error; err != nil {
		return err
	}
	for _, pipeline := range pipelines {
		if err := w.PipelineManager.Delete(pipeline.ID); err != nil {
			return err
		}
	}
	if err := w.DB.Delete(&types.UserRole{}, "scope = ? and scope_id = ?", types.RoleScopePipeline, workspace.ID).Error; err != nil {
		return err
	}
	if err := w.DB.Delete(workspace).Error; err != nil {
		return err
	}
	return nil
}
