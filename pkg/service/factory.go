package service

import (
	"github.com/kubespace/kubespace/pkg/service/config"
	"github.com/kubespace/kubespace/pkg/service/pipeline"
)

type Factory struct {
	config          *config.ServiceConfig
	PipelineService *pipeline.ServicePipeline
}

func NewServiceFactory(config *config.ServiceConfig) *Factory {
	pipelineService := pipeline.NewPipelineService(config.Models)
	return &Factory{
		config:          config,
		PipelineService: pipelineService,
	}
}
