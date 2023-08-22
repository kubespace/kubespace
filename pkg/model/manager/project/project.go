package project

import (
	"fmt"
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
	var cnt int64
	if err := p.DB.Model(&types.Project{}).Where("cluster_id = ? and namespace = ?", project.ClusterId, project.Namespace).Count(&cnt).Error; err != nil {
		return nil, err
	}
	if cnt > 0 {
		return nil, fmt.Errorf("已有工作空间绑定该集群命名空间")
	}
	result := p.DB.Create(project)
	if result.Error != nil {
		return nil, result.Error
	}
	return project, nil
}

func (p *ManagerProject) Update(project *types.Project) (*types.Project, error) {
	result := p.DB.Save(project)
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
	var apps []types.App
	var err error
	if err = p.DB.Where("scope = ? and scope_id = ?", types.AppVersionScopeProjectApp, project.ID).Find(&apps).Error; err != nil {
		return err
	}
	for _, app := range apps {
		if err = p.ProjectAppManager.DeleteApp(app.ID); err != nil {
			return err
		}
	}
	if err = p.DB.Delete(&types.UserRole{}, "scope = ? and scope_id = ?", types.ScopeProject, project.ID).Error; err != nil {
		return err
	}
	result := p.DB.Delete(project)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (p *ManagerProject) Clone(originProjectId uint, newProject *types.Project) (*types.Project, error) {
	err := p.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(newProject).Error; err != nil {
			return err
		}
		var originApps []types.App
		var err error
		if err = tx.Where("scope = ? and scope_id = ?", types.AppVersionScopeProjectApp, originProjectId).Find(&originApps).Error; err != nil {
			return err
		}
		for _, app := range originApps {
			var appVersion types.AppVersion
			if err = tx.First(&appVersion, "id = ?", app.AppVersionId).Error; err != nil {
				return err
			}
			app.ID = 0
			app.ScopeId = newProject.ID
			app.Status = types.AppStatusUninstall
			app.AppVersionId = 0
			if err = tx.Create(&app).Error; err != nil {
				return err
			}
			appVersion.ScopeId = app.ID
			appVersion.ID = 0
			if err = tx.Create(&appVersion).Error; err != nil {
				return err
			}
			app.AppVersionId = appVersion.ID
			if err = tx.Save(&app).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return newProject, nil
}
