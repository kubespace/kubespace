package api

import (
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/api/apps"
	"github.com/kubespace/kubespace/pkg/server/api/apps/appstore"
	"github.com/kubespace/kubespace/pkg/server/api/audit"
	"github.com/kubespace/kubespace/pkg/server/api/cluster"
	"github.com/kubespace/kubespace/pkg/server/api/pipeline"
	"github.com/kubespace/kubespace/pkg/server/api/project"
	"github.com/kubespace/kubespace/pkg/server/api/settings"
	"github.com/kubespace/kubespace/pkg/server/api/spacelet"
	"github.com/kubespace/kubespace/pkg/server/api/user"
	"github.com/kubespace/kubespace/pkg/server/config"
)

func Apis(c *config.ServerConfig) map[string]api.ApiGroup {
	return map[string]api.ApiGroup{
		"user":     user.ApiGroup(c),
		"audit":    audit.ApiGroup(c),
		"cluster":  cluster.ApiGroup(c),
		"pipeline": pipeline.ApiGroup(c),
		"project":  project.ApiGroup(c),
		"apps":     apps.ApiGroup(c),
		"appstore": appstore.ApiGroup(c),
		"settings": settings.ApiGroup(c),
		"spacelet": spacelet.ApiGroup(c),
	}
}

//type ViewSets map[string][]*api.Api
//
//func NewViewSets(conf *config.ServerConfig) *ViewSets {
//	models := conf.Models
//
//	// 集群相关操作
//	clusterViews := cluster.NewCluster(conf)
//	kubeResource := cluster.NewKubeResource(conf)
//
//	// 用户角色
//	userViews := user.NewUser(models)
//	userRole := user.NewUserRole(models)
//	settingsRole := user.NewRole(models)
//
//	// 流水线
//	pipelineWorkspace := pipeline.NewPipelineWorkspace(conf)
//	pipelineViews := pipeline.NewPipeline(conf)
//	pipelineRun := pipeline.NewPipelineRun(conf)
//	pipelineResource := pipeline.NewPipelineResource(conf)
//	spaceletViews := spacelet.NewSpaceletViews(conf)
//
//	// 配置
//	settingsViews := settings.NewSettings(conf)
//	settingsSecret := settings.NewSettingsSecret(models)
//	imageRegistry := settings.NewImageRegistry(models)
//	settingLdap := settings.NewSettingsLdap(models)
//
//	// 工作空间以及应用商店
//	projectWorkspace := project.NewProject(conf)
//	projectApps := project.NewProjectApp(conf)
//	appStore := project.NewAppStore(conf)
//
//	auditViews := audit.NewAuditOperate(conf.Models)
//
//	viewsets := &ViewSets{
//		"cluster":          clusterViews.Views,
//		"user":             userViews.Views,
//		"user_role":        userRole.Views,
//		"settings_role":    settingsRole.Views,
//		"cluster/:cluster": kubeResource.Views,
//
//		"pipeline/workspace": pipelineWorkspace.Views,
//		"pipeline/pipeline":  pipelineViews.Views,
//		"pipeline/build":     pipelineRun.Views,
//		"pipeline/resource":  pipelineResource.Views,
//
//		"spacelet": spaceletViews.Views,
//
//		"settings":                settingsViews.Views,
//		"settings/secret":         settingsSecret.Views,
//		"settings/image_registry": imageRegistry.Views,
//		"settings/ldap":           settingLdap.Views,
//
//		"project/workspace": projectWorkspace.Views,
//		"project/apps":      projectApps.Views,
//		"appstore":          appStore.Views,
//
//		"audit": auditViews.Views,
//	}
//
//	return viewsets
//}
