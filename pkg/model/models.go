package model

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/db"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/config"
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/model/manager"
	"github.com/kubespace/kubespace/pkg/model/manager/cluster"
	"github.com/kubespace/kubespace/pkg/model/manager/pipeline"
	"github.com/kubespace/kubespace/pkg/model/manager/project"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

type Config struct {
	Db                *db.DB
	ListWatcherConfig *config.ListWatcherConfig
}

type Models struct {
	db                *gorm.DB
	ListWatcherConfig *config.ListWatcherConfig

	*cluster.ClusterManager
	*manager.UserManager
	*manager.UserRoleManager
	*manager.TokenManager
	*manager.RoleManager
	*manager.AppManager
	*pipeline.ManagerPipeline
	*pipeline.ManagerPipelineRun
	PipelineWorkspaceManager *pipeline.WorkspaceManager
	PipelinePluginManager    *pipeline.ManagerPipelinePlugin
	PipelineResourceManager  *pipeline.ResourceManager
	PipelineJobLogManager    *pipeline.JobLog
	PipelineReleaseManager   *pipeline.Release

	*manager.SettingsSecretManager
	*manager.ImageRegistryManager
	ProjectAppManager        *project.AppManager
	ProjectAppVersionManager *project.AppVersionManager
	ProjectManager           *project.ManagerProject
	AppStoreManager          *project.AppStoreManager
}

func NewModels(c *Config) (*Models, error) {
	if err := DbMigrate(c.Db.Instance); err != nil {
		return nil, fmt.Errorf("migrate db error: %s", err.Error())
	}

	middleMessage := kube_resource.NewMiddleMessageWithClient(nil, c.Db.RedisInstance)
	role := manager.NewRoleManager(c.Db.RedisInstance)
	tk := manager.NewTokenManager(c.Db.RedisInstance)
	app := manager.NewAppManager(c.Db.RedisInstance)

	user := manager.NewUserManager(c.Db.Instance)
	userRole := manager.NewUserRoleManager(c.Db.Instance, user)
	pipelinePluginMgr := pipeline.NewPipelinePluginManager(c.Db.Instance)
	pipelineMgr := pipeline.NewPipelineManager(c.Db.Instance)
	pipelineWorkspaceMgr := pipeline.NewWorkspaceManager(c.Db.Instance, pipelineMgr)
	pipelineRunMgr := pipeline.NewPipelineRunManager(c.Db.Instance, pipelinePluginMgr, middleMessage)
	pipelineResourceMgr := pipeline.NewResourceManager(c.Db.Instance)
	jobLogMgr := pipeline.NewJobLogManager(c.Db.Instance)
	pipelineReleaseMgr := pipeline.NewReleaseManager(c.Db.Instance)

	secrets := manager.NewSettingsSecretManager(c.Db.Instance)
	imageRegistry := manager.NewSettingsImageRegistryManager(c.Db.Instance)

	appVersionMgr := project.NewAppVersionManager(c.Db.Instance)
	projectAppMgr := project.NewAppManager(appVersionMgr, c.Db.Instance)
	appStoreMgr := project.NewAppStoreManager(appVersionMgr, c.Db.Instance)
	projectMgr := project.NewManagerProject(c.Db.Instance, projectAppMgr)

	cm := cluster.NewClusterManager(c.Db.Instance, c.ListWatcherConfig, projectAppMgr)

	return &Models{
		db:                       c.Db.Instance,
		ListWatcherConfig:        c.ListWatcherConfig,
		ClusterManager:           cm,
		UserManager:              user,
		UserRoleManager:          userRole,
		TokenManager:             tk,
		RoleManager:              role,
		AppManager:               app,
		ManagerPipeline:          pipelineMgr,
		ManagerPipelineRun:       pipelineRunMgr,
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
