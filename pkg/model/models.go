package model

import (
	"github.com/kubespace/kubespace/pkg/core/mysql"
	"github.com/kubespace/kubespace/pkg/core/redis"
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/model/listwatcher/config"
	"github.com/kubespace/kubespace/pkg/model/manager"
	"github.com/kubespace/kubespace/pkg/model/manager/cluster"
	"github.com/kubespace/kubespace/pkg/model/manager/pipeline"
	"github.com/kubespace/kubespace/pkg/model/manager/project"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

type Options struct {
	RedisOptions *redis.Options
	MysqlOptions *mysql.Options
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

func NewModels(options *Options) (*Models, error) {
	db, err := mysql.NewMysqlDb(options.MysqlOptions)
	if err != nil {
		return nil, err
	}
	redisClient := redis.NewRedisClient(options.RedisOptions)
	listWatcherConfig := config.NewListWatcherConfig(db, redisClient)

	middleMessage := kube_resource.NewMiddleMessageWithClient(nil, redisClient)
	role := manager.NewRoleManager(redisClient)
	tk := manager.NewTokenManager(redisClient)
	app := manager.NewAppManager(redisClient)

	user := manager.NewUserManager(db)
	userRole := manager.NewUserRoleManager(db, user)
	pipelinePluginMgr := pipeline.NewPipelinePluginManager(db)
	pipelineMgr := pipeline.NewPipelineManager(db)
	pipelineWorkspaceMgr := pipeline.NewWorkspaceManager(db, pipelineMgr)
	pipelineRunMgr := pipeline.NewPipelineRunManager(db, pipelinePluginMgr, middleMessage)
	pipelineResourceMgr := pipeline.NewResourceManager(db)
	jobLogMgr := pipeline.NewJobLogManager(db)
	pipelineReleaseMgr := pipeline.NewReleaseManager(db)

	secrets := manager.NewSettingsSecretManager(db)
	imageRegistry := manager.NewSettingsImageRegistryManager(db)

	appVersionMgr := project.NewAppVersionManager(db)
	projectAppMgr := project.NewAppManager(appVersionMgr, db)
	appStoreMgr := project.NewAppStoreManager(appVersionMgr, db)
	projectMgr := project.NewManagerProject(db, projectAppMgr)

	cm := cluster.NewClusterManager(db, listWatcherConfig, projectAppMgr)

	return &Models{
		db:                       db,
		ListWatcherConfig:        listWatcherConfig,
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
