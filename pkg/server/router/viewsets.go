package router

import (
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/cluster"
	"github.com/kubespace/kubespace/pkg/server/views/pipeline"
	"github.com/kubespace/kubespace/pkg/server/views/project"
	"github.com/kubespace/kubespace/pkg/server/views/settings"
	"github.com/kubespace/kubespace/pkg/server/views/spacelet"
	"github.com/kubespace/kubespace/pkg/server/views/user"
)

type ViewSets map[string][]*views.View

func NewViewSets(conf *config.ServerConfig) *ViewSets {
	models := conf.Models

	// 集群相关操作
	clusterViews := cluster.NewCluster(conf)
	kubeResource := cluster.NewKubeResource(conf)

	// 用户角色
	userViews := user.NewUser(models)
	userRole := user.NewUserRole(models)
	settingsRole := user.NewRole(models)

	// 流水线
	pipelineWorkspace := pipeline.NewPipelineWorkspace(conf)
	pipelineViews := pipeline.NewPipeline(conf)
	pipelineRun := pipeline.NewPipelineRun(conf)
	pipelineResource := pipeline.NewPipelineResource(conf)
	spaceletViews := spacelet.NewSpaceletViews(conf)

	// 配置
	settingsSecret := settings.NewSettingsSecret(models)
	imageRegistry := settings.NewImageRegistry(models)
	settingLdap := settings.NewSettingsLdap(models)

	// 工作空间以及应用商店
	projectWorkspace := project.NewProject(conf)
	projectApps := project.NewProjectApp(conf)
	appStore := project.NewAppStore(conf)

	viewsets := &ViewSets{
		"cluster":          clusterViews.Views,
		"user":             userViews.Views,
		"user_role":        userRole.Views,
		"settings_role":    settingsRole.Views,
		"cluster/:cluster": kubeResource.Views,

		"pipeline/workspace": pipelineWorkspace.Views,
		"pipeline/pipeline":  pipelineViews.Views,
		"pipeline/build":     pipelineRun.Views,
		"pipeline/resource":  pipelineResource.Views,

		"spacelet": spaceletViews.Views,

		"settings/secret":         settingsSecret.Views,
		"settings/image_registry": imageRegistry.Views,
		"settings/ldap":           settingLdap.Views,

		"project/workspace": projectWorkspace.Views,
		"project/apps":      projectApps.Views,
		"appstore":          appStore.Views,
	}

	return viewsets
}
