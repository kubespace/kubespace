package model

import (
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/model/manager"
	"github.com/kubespace/kubespace/pkg/model/manager/pipeline"
	"github.com/kubespace/kubespace/pkg/model/manager/project"
	"github.com/kubespace/kubespace/pkg/model/mysql"
	"github.com/kubespace/kubespace/pkg/redis"
)

type Models struct {
	*manager.ClusterManager
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

func NewModels(redisOp *redis.Options, mysqlOptions *mysql.Options) (*Models, error) {
	client := redis.NewRedisClient(redisOp)

	db, err := mysql.NewMysqlDb(mysqlOptions)
	if err != nil {
		return nil, err
	}

	middleMessage := kube_resource.NewMiddleMessageWithClient(nil, client)
	role := manager.NewRoleManager(client)
	tk := manager.NewTokenManager(client)
	app := manager.NewAppManager(client)

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

	cm := manager.NewClusterManager(db, projectAppMgr)

	return &Models{
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
