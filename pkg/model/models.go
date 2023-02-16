package model

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/db"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/config"
	"github.com/kubespace/kubespace/pkg/model/manager"
	"github.com/kubespace/kubespace/pkg/model/manager/cluster"
	"github.com/kubespace/kubespace/pkg/model/manager/pipeline"
	"github.com/kubespace/kubespace/pkg/model/manager/project"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

type Config struct {
	DB                *db.DB
	ListWatcherConfig *config.ListWatcherConfig
}

type Models struct {
	db                *gorm.DB
	ListWatcherConfig *config.ListWatcherConfig

	ClusterManager *cluster.ClusterManager

	UserManager     *manager.UserManager
	UserRoleManager *manager.UserRoleManager
	TokenManager    *manager.TokenManager
	RoleManager     *manager.RoleManager

	PipelineManager          *pipeline.ManagerPipeline
	PipelineRunManager       *pipeline.ManagerPipelineRun
	PipelineWorkspaceManager *pipeline.WorkspaceManager
	PipelinePluginManager    *pipeline.ManagerPipelinePlugin
	PipelineResourceManager  *pipeline.ResourceManager
	PipelineJobLogManager    *pipeline.JobLog
	PipelineReleaseManager   *pipeline.Release

	ProjectAppManager        *project.AppManager
	ProjectAppVersionManager *project.AppVersionManager
	ProjectManager           *project.ManagerProject
	AppStoreManager          *project.AppStoreManager

	SettingsSecretManager *manager.SettingsSecretManager
	ImageRegistryManager  *manager.ImageRegistryManager
}

func NewModels(c *Config) (*Models, error) {
	if err := DbMigrate(c.DB.Instance); err != nil {
		return nil, fmt.Errorf("migrate db error: %s", err.Error())
	}

	role := manager.NewRoleManager(c.DB.RedisInstance)
	tk := manager.NewTokenManager(c.DB.RedisInstance)

	user := manager.NewUserManager(c.DB.Instance)
	userRole := manager.NewUserRoleManager(c.DB.Instance, user)
	pipelinePluginMgr := pipeline.NewPipelinePluginManager(c.DB.Instance)
	pipelineMgr := pipeline.NewPipelineManager(c.DB.Instance, userRole)
	pipelineWorkspaceMgr := pipeline.NewWorkspaceManager(c.DB.Instance, pipelineMgr)
	pipelineRunMgr := pipeline.NewPipelineRunManager(c.DB.Instance, pipelinePluginMgr, c.ListWatcherConfig)
	pipelineResourceMgr := pipeline.NewResourceManager(c.DB.Instance)
	jobLogMgr := pipeline.NewJobLogManager(c.DB.Instance)
	pipelineReleaseMgr := pipeline.NewReleaseManager(c.DB.Instance)

	secrets := manager.NewSettingsSecretManager(c.DB.Instance)
	imageRegistry := manager.NewSettingsImageRegistryManager(c.DB.Instance)

	appVersionMgr := project.NewAppVersionManager(c.DB.Instance)
	projectAppMgr := project.NewAppManager(appVersionMgr, c.DB.Instance)
	appStoreMgr := project.NewAppStoreManager(appVersionMgr, c.DB.Instance)
	projectMgr := project.NewManagerProject(c.DB.Instance, projectAppMgr)

	cm := cluster.NewClusterManager(c.DB.Instance, c.ListWatcherConfig, projectAppMgr)

	return &Models{
		db:                       c.DB.Instance,
		ListWatcherConfig:        c.ListWatcherConfig,
		ClusterManager:           cm,
		UserManager:              user,
		UserRoleManager:          userRole,
		TokenManager:             tk,
		RoleManager:              role,
		PipelineManager:          pipelineMgr,
		PipelineRunManager:       pipelineRunMgr,
		PipelineWorkspaceManager: pipelineWorkspaceMgr,
		PipelinePluginManager:    pipelinePluginMgr,
		PipelineResourceManager:  pipelineResourceMgr,
		PipelineJobLogManager:    jobLogMgr,
		PipelineReleaseManager:   pipelineReleaseMgr,
		SettingsSecretManager:    secrets,
		ProjectManager:           projectMgr,
		ProjectAppManager:        projectAppMgr,
		ProjectAppVersionManager: appVersionMgr,
		ImageRegistryManager:     imageRegistry,
		AppStoreManager:          appStoreMgr,
	}, nil
}

func DbMigrate(db *gorm.DB) error {
	var err error
	migrateTypes := []interface{}{
		&types.Cluster{},
		&types.User{},
		&types.UserRole{},
		&types.PipelineWorkspace{},
		&types.Pipeline{},
		&types.PipelineStage{},
		&types.PipelinePlugin{},
		&types.PipelineRun{},
		&types.PipelineRunStage{},
		&types.PipelineRunJob{},
		&types.PipelineRunJobLog{},
		&types.PipelineResource{},
		&types.PipelineWorkspaceRelease{},

		&types.SettingsSecret{},
		&types.SettingsImageRegistry{},

		&types.Project{},
		&types.ProjectApp{},
		&types.AppVersion{},
		&types.AppVersionChart{},
		&types.AppStore{},
		&types.ProjectAppRevision{},
	}
	for _, model := range migrateTypes {
		err = db.AutoMigrate(model)
		if err != nil {
			return err
		}
	}

	return nil
}
