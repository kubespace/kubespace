package pipeline

import (
	"github.com/kubespace/kubespace/pkg/model/manager"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

type WorkspaceManager struct {
	*manager.CommonManager
}

func NewWorkspaceManager(db *gorm.DB) *WorkspaceManager {
	return &WorkspaceManager{
		CommonManager: manager.NewCommonManager(nil, db, "", false),
	}
}

func (w *WorkspaceManager) Create(workspace *types.PipelineWorkspace) (*types.PipelineWorkspace, error) {
	result := w.DB.Create(workspace)
	if result.Error != nil {
		return nil, result.Error
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
	result := w.DB.Delete(workspace)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
