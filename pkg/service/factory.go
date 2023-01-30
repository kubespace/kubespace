package service

import (
	"github.com/kubespace/kubespace/pkg/service/cluster"
	"github.com/kubespace/kubespace/pkg/service/config"
	"github.com/kubespace/kubespace/pkg/service/pipeline"
	"github.com/kubespace/kubespace/pkg/service/project"
)

type Factory struct {
	config   *config.ServiceConfig
	Cluster  *ClusterFactory
	Project  *ProjectFactory
	Pipeline *PipelineFactory
}

func NewServiceFactory(config *config.ServiceConfig) *Factory {
	kubeClient := cluster.NewKubeClient(config.Models)
	appBase := project.NewAppBaseService(config.Models)
	appService := project.NewAppService(kubeClient, appBase)
	projectService := project.NewProjectService(config.Models, kubeClient, appService)
	return &Factory{
		config: config,
		Cluster: &ClusterFactory{
			KubeClient: kubeClient,
		},
		Project: &ProjectFactory{
			ProjectService:  projectService,
			AppService:      appService,
			AppStoreService: project.NewAppStoreService(appBase),
		},
		Pipeline: &PipelineFactory{
			WorkspaceService:   pipeline.NewWorkspaceService(config.Models),
			PipelineService:    pipeline.NewPipelineService(config.Models),
			PipelineRunService: pipeline.NewPipelineRunService(config.Models, kubeClient),
		},
	}
}

type ClusterFactory struct {
	KubeClient *cluster.KubeClient
}

type ProjectFactory struct {
	ProjectService  *project.ProjectService
	AppService      *project.AppService
	AppStoreService *project.AppStoreService
}

type PipelineFactory struct {
	WorkspaceService   *pipeline.WorkspaceService
	PipelineService    *pipeline.ServicePipeline
	PipelineRunService *pipeline.ServicePipelineRun
}
