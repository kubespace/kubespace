package model

import (
	"github.com/kubespace/kubespace/pkg/model/manager"
	"github.com/kubespace/kubespace/pkg/model/manager/pipeline"
	"github.com/kubespace/kubespace/pkg/model/mysql"
	"github.com/kubespace/kubespace/pkg/redis"
)

type Models struct {
	*manager.ClusterManager
	*manager.UserManager
	*manager.TokenManager
	*manager.RoleManager
	*manager.AppManager
	*pipeline.ManagerPipeline
	*pipeline.ManagerPipelineRun
	PipelineWorkspaceManager *pipeline.WorkspaceManager
	PipelinePluginManager    *pipeline.ManagerPipelinePlugin
	*manager.SettingsSecretManager
}

func NewModels(redisOp *redis.Options, mysqlOptions *mysql.Options) (*Models, error) {
	client := redis.NewRedisClient(redisOp)
	cm := manager.NewClusterManager(client)
	role := manager.NewRoleManager(client)
	user := manager.NewUserManager(client, role)
	tk := manager.NewTokenManager(client)
	app := manager.NewAppManager(client)

	db, err := mysql.NewMysqlDb(mysqlOptions)
	if err != nil {
		return nil, err
	}

	pipelinePluginMgr := pipeline.NewPipelinePluginManager(db)
	pipelineMgr := pipeline.NewPipelineManager(db)
	pipelineWorkspaceMgr := pipeline.NewWorkspaceManager(db)
	pipelineRunMgr := pipeline.NewPipelineRunManager(db, pipelinePluginMgr)

	secrets := manager.NewSettingsSecretManager(db)

	return &Models{
		ClusterManager:           cm,
		UserManager:              user,
		TokenManager:             tk,
		RoleManager:              role,
		AppManager:               app,
		ManagerPipeline:          pipelineMgr,
		ManagerPipelineRun:       pipelineRunMgr,
		PipelineWorkspaceManager: pipelineWorkspaceMgr,
		PipelinePluginManager:    pipelinePluginMgr,
		SettingsSecretManager: secrets,
	}, nil
}