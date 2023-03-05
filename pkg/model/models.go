package model

import (
	"github.com/kubespace/kubespace/pkg/core/db"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/config"
	"github.com/kubespace/kubespace/pkg/model/manager/cluster"
	"github.com/kubespace/kubespace/pkg/model/manager/pipeline"
	"github.com/kubespace/kubespace/pkg/model/manager/project"
	"github.com/kubespace/kubespace/pkg/model/manager/settings"
	"github.com/kubespace/kubespace/pkg/model/manager/spacelet"
	"github.com/kubespace/kubespace/pkg/model/manager/user"
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

	UserManager     *user.UserManager
	UserRoleManager *user.UserRoleManager
	TokenManager    *user.TokenManager
	RoleManager     *user.RoleManager

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
	LdapManager           *manager.LdapManager
	SpaceletManager *spacelet.SpaceletManager

}

func NewModels(c *Config) (*Models, error) {
	role := user.NewRoleManager(c.DB.RedisInstance)
	tk := user.NewTokenManager(c.DB.RedisInstance)

	userMgr := user.NewUserManager(c.DB.Instance)
	userRole := user.NewUserRoleManager(c.DB.Instance, userMgr)
	pipelinePluginMgr := pipeline.NewPipelinePluginManager(c.DB.Instance)
	pipelineMgr := pipeline.NewPipelineManager(c.DB.Instance)
	pipelineWorkspaceMgr := pipeline.NewWorkspaceManager(c.DB.Instance, pipelineMgr)
	pipelineRunMgr := pipeline.NewPipelineRunManager(c.DB.Instance, pipelinePluginMgr, c.ListWatcherConfig)
	pipelineResourceMgr := pipeline.NewResourceManager(c.DB.Instance)
	jobLogMgr := pipeline.NewJobLogManager(c.DB.Instance)
	pipelineReleaseMgr := pipeline.NewReleaseManager(c.DB.Instance)


	secrets := manager.NewSettingsSecretManager(c.DB.Instance)
	imageRegistry := manager.NewSettingsImageRegistryManager(c.DB.Instance)
	ldap := manager.NewLdapManager(c.DB.Instance)


	appVersionMgr := project.NewAppVersionManager(c.DB.Instance)
	projectAppMgr := project.NewAppManager(appVersionMgr, c.DB.Instance)
	appStoreMgr := project.NewAppStoreManager(appVersionMgr, c.DB.Instance)
	projectMgr := project.NewManagerProject(c.DB.Instance, projectAppMgr)

	cm := cluster.NewClusterManager(c.DB.Instance, c.ListWatcherConfig, projectAppMgr)

	sl := spacelet.NewSpaceletManager(c.DB.Instance)

	return &Models{
		db:                       c.DB.Instance,
		ListWatcherConfig:        c.ListWatcherConfig,
		ClusterManager:           cm,
		UserManager:              userMgr,
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
		LdapManager:              ldap,
		ProjectManager:           projectMgr,
		ProjectAppManager:        projectAppMgr,
		ProjectAppVersionManager: appVersionMgr,
		ImageRegistryManager:     imageRegistry,
		AppStoreManager:          appStoreMgr,
		SpaceletManager:          sl,
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
		&types.Ldap{},

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

