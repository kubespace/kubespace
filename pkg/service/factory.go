package service

import (
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/service/cluster"
	"github.com/kubespace/kubespace/pkg/service/pipeline"
	"github.com/kubespace/kubespace/pkg/service/pipeline/pipeline_run"
	"github.com/kubespace/kubespace/pkg/service/project"
	"github.com/kubespace/kubespace/pkg/service/spacelet"
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
			PipelineRunService: pipeline_run.NewPipelineRunService(config.models),
			SpaceletService:    spacelet.NewSpaceletService(config.models),
		},
	}
}

// ClusterFactory 集群相关service
type ClusterFactory struct {
	// 集群资源操作客户端
	KubeClient *cluster.KubeClient
}

// ProjectFactory 工作空间相关service
type ProjectFactory struct {
	// 工作空间
	ProjectService *project.ProjectService
	// 应用
	AppService *project.AppService
	// 应用商店
	AppStoreService *project.AppStoreService
}

// PipelineFactory 流水线相关service
type PipelineFactory struct {
	// 流水线空间
	WorkspaceService *pipeline.WorkspaceService
	// 流水线
	PipelineService *pipeline.PipelineService
	// 流水线构建
	PipelineRunService *pipeline_run.PipelineRunService
	// spacelet
	SpaceletService *spacelet.SpaceletService
}
