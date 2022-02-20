package project

import (
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

type ManagerProject struct {
	*gorm.DB
}

func NewManagerProject(db *gorm.DB) *ManagerProject {
	return &ManagerProject{DB: db}
}

func (p *ManagerProject) Create(project *types.Project) (*types.Project, error) {
	result := p.DB.Create(project)
	if result.Error != nil {
		return nil, result.Error
	}
	return project, nil
}

func (p *ManagerProject) Get(projectId uint) (*types.Project, error) {
	var ws types.Project
	if err := p.DB.First(&ws, projectId).Error; err != nil {
		return nil, err
	}
	return &ws, nil
}

func (p *ManagerProject) List() ([]types.Project, error) {
	var ws []types.Project
	result := p.DB.Find(&ws)
	if result.Error != nil {
		return nil, result.Error
	}
	return ws, nil
}

func (p *ManagerProject) Delete(project *types.Project) error {
	result := p.DB.Delete(project)
	if result.Error != nil {
		return result.Error
	}
	return nil
}


