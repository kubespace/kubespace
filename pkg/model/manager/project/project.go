package project

import (
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

type ManagerProject struct {
	*gorm.DB
	ProjectAppManager *AppManager
}

func NewManagerProject(db *gorm.DB, appManager *AppManager) *ManagerProject {
	return &ManagerProject{DB: db, ProjectAppManager: appManager}
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
	var apps []types.ProjectApp
	var err error
	if err = p.DB.Where("project_id = ?", project.ID).Find(&apps).Error; err != nil {
		return err
	}
	for _, app := range apps {
		if err = p.ProjectAppManager.DeleteProjectApp(app.ID); err != nil {
			return err
		}
	}
	result := p.DB.Delete(project)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
