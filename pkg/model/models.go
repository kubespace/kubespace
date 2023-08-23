package model

import (
	"github.com/kubespace/kubespace/pkg/core/db"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/config"
	"github.com/kubespace/kubespace/pkg/model/manager/audit"
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

	PipelineManager             *pipeline.ManagerPipeline
	PipelineRunManager          *pipeline.PipelineRunManager
	PipelineWorkspaceManager    *pipeline.WorkspaceManager
	PipelinePluginManager       *pipeline.PipelinePluginManager
	PipelineResourceManager     *pipeline.ResourceManager
	PipelineJobLogManager       *pipeline.JobLog
	PipelineReleaseManager      *pipeline.Release
	PipelineTriggerManager      *pipeline.PipelineTriggerManager
	PipelineTriggerEventManager *pipeline.PipelineTriggerEventManager
	PipelineCodeCacheManager    *pipeline.PipelineCodeCacheManager

	AppManager        *project.AppManager
	AppVersionManager *project.AppVersionManager
	ProjectManager    *project.ManagerProject
	AppStoreManager   *project.AppStoreManager

	SettingsSecretManager *settings.SettingsSecretManager
	ImageRegistryManager  *settings.ImageRegistryManager

	LdapManager     *settings.LdapManager
	SpaceletManager *spacelet.SpaceletManager

	AuditOperateManager *audit.AuditOperateManager
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
	pipelineTriggerMgr := pipeline.NewPipelineTriggerManager(c.DB.Instance, c.ListWatcherConfig)
	pipelineTriggerEventMgr := pipeline.NewPipelineTriggerEventManager(c.DB.Instance, c.ListWatcherConfig)
	pipelineCodeCacheMgr := pipeline.NewPipelineCodeCacheManager(c.DB.Instance)

	secrets := settings.NewSettingsSecretManager(c.DB.Instance)
	imageRegistry := settings.NewSettingsImageRegistryManager(c.DB.Instance)

	ldap := settings.NewLdapManager(c.DB.Instance)

	appVersionMgr := project.NewAppVersionManager(c.DB.Instance)
	AppMgr := project.NewAppManager(appVersionMgr, c.DB.Instance)
	appStoreMgr := project.NewAppStoreManager(appVersionMgr, c.DB.Instance)
	projectMgr := project.NewManagerProject(c.DB.Instance, AppMgr)

	cm := cluster.NewClusterManager(c.DB.Instance, c.ListWatcherConfig, AppMgr)

	sl := spacelet.NewSpaceletManager(c.DB.Instance)

	auditOperateMgr := audit.NewAuditOperateManager(c.DB.Instance)

	return &Models{
		db:                          c.DB.Instance,
		ListWatcherConfig:           c.ListWatcherConfig,
		ClusterManager:              cm,
		UserManager:                 userMgr,
		UserRoleManager:             userRole,
		TokenManager:                tk,
		RoleManager:                 role,
		PipelineManager:             pipelineMgr,
		PipelineRunManager:          pipelineRunMgr,
		PipelineWorkspaceManager:    pipelineWorkspaceMgr,
		PipelinePluginManager:       pipelinePluginMgr,
		PipelineResourceManager:     pipelineResourceMgr,
		PipelineJobLogManager:       jobLogMgr,
		PipelineReleaseManager:      pipelineReleaseMgr,
		PipelineTriggerManager:      pipelineTriggerMgr,
		PipelineTriggerEventManager: pipelineTriggerEventMgr,
		PipelineCodeCacheManager:    pipelineCodeCacheMgr,
		SettingsSecretManager:       secrets,
		LdapManager:                 ldap,
		ProjectManager:              projectMgr,
		AppManager:                  AppMgr,
		AppVersionManager:           appVersionMgr,
		ImageRegistryManager:        imageRegistry,
		AppStoreManager:             appStoreMgr,
		SpaceletManager:             sl,
		AuditOperateManager:         auditOperateMgr,
	}, nil
}
