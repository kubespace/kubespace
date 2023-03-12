package service

import (
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/service/cluster"
	"github.com/kubespace/kubespace/pkg/service/pipeline"
	"github.com/kubespace/kubespace/pkg/service/project"
)

type Config struct {
	models *model.Models
}

func NewConfig(models *model.Models) *Config {
	return &Config{models: models}
}

type Factory struct {
	// 集群相关Service
	Cluster *ClusterFactory
	// 项目相关Service，如工作空间、应用
	Project *ProjectFactory
	// 流水线相关Service
	Pipeline *PipelineFactory
}

func NewServiceFactory(config *Config) *Factory {
	kubeClient := cluster.NewKubeClient(config.models)
	appBase := project.NewAppBaseService(config.models)
	appService := project.NewAppService(kubeClient, appBase)
	projectService := project.NewProjectService(config.models, kubeClient, appService)
	return &Factory{
		Cluster: &ClusterFactory{
			KubeClient: kubeClient,
		},
		Project: &ProjectFactory{
			ProjectService:  projectService,
			AppService:      appService,
			AppStoreService: project.NewAppStoreService(appBase),
		},
		Pipeline: &PipelineFactory{
			WorkspaceService:   pipeline.NewWorkspaceService(config.models),
			PipelineService:    pipeline.NewPipelineService(config.models),
			PipelineRunService: pipeline.NewPipelineRunService(config.models),
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
