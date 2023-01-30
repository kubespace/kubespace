package router

import (
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/cluster"
	"github.com/kubespace/kubespace/pkg/server/views/user"
)

type ViewSets map[string][]*views.View

func NewViewSets(conf *config.ServerConfig) *ViewSets {
	models := conf.Models
	clusterViews := cluster.NewCluster(conf)
	userViews := user.NewUser(models)
	userRole := user.NewUserRole(models)
	settingsRole := user.NewRole(models)

	kubeResource := cluster.NewKubeResource(conf)

	//pipelineRunService := pipeline.NewPipelineRunService(models, kr)
	//
	//pipelineWorkspace := pipeline_views.NewPipelineWorkspace(models)
	//pipelineViews := pipeline_views.NewPipeline(models, pipelineRunService)
	//pipelineRun := pipeline_views.NewPipelineRun(models, pipelineRunService)
	//pipelineResource := pipeline_views.NewPipelineResource(models)
	//
	//settingsSecret := settings_views.NewSettingsSecret(models)
	//imageRegistry := settings_views.NewImageRegistry(models)
	//
	//appBaseService := project2.NewAppBaseService(models)
	//projectAppService := project2.NewAppService(kr, appBaseService)
	//appStoreService := project2.NewAppStoreService(appBaseService)
	//projectService := project2.NewProjectService(models, kr, projectAppService)
	//
	//projectWorkspace := project_views.NewProject(models, projectService)
	//projectApps := project_views.NewProjectApp(models, projectAppService)
	//appStore := project_views.NewAppStore(models, appStoreService)

	viewsets := &ViewSets{
		"cluster":          clusterViews.Views,
		"user":             userViews.Views,
		"user_role":        userRole.Views,
		"settings_role":    settingsRole.Views,
		"cluster/:cluster": kubeResource.Views,

		//"pipeline/workspace": pipelineWorkspace.Views,
		//"pipeline/pipeline":  pipelineViews.Views,
		//"pipeline/build":     pipelineRun.Views,
		//"pipeline/resource":  pipelineResource.Views,
		//
		//"settings/secret":         settingsSecret.Views,
		//"settings/image_registry": imageRegistry.Views,
		//
		//"project/workspace": projectWorkspace.Views,
		//"project/apps":      projectApps.Views,
		//"appstore":          appStore.Views,
	}

	return viewsets
}
