package informer

import (
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/cluster"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/config"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/pipeline"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/spacelet"
)

type Factory interface {
	ClusterAgentInformer(token string) Informer

	PipelineRunInformer(cond *pipeline.PipelineRunWatchCondition) Informer
	PipelineRunJobInformer(cond *pipeline.PipelineRunJobWatchCondition) Informer
	PipelineTriggerInformer(cond *pipeline.PipelineTriggerWatchCondition) Informer
	PipelineTriggerEventInformer(cond *pipeline.PipelineTriggerEventWatchCondition) Informer
	PipelineCodeCacheInformer(cond *pipeline.PipelineCodeCacheWatchCondition) Informer

	SpaceletInformer(cond *spacelet.SpaceletWatchCondition) Informer
}

type informerFactory struct {
	config *config.ListWatcherConfig
}

func NewInformerFactory(config *config.ListWatcherConfig) Factory {
	return &informerFactory{
		config: config,
	}
}

func (s *informerFactory) ClusterAgentInformer(token string) Informer {
	agentListWatcher := cluster.NewAgentListWatcher(token, s.config)
	return NewInformer(agentListWatcher)
}

func (s *informerFactory) PipelineRunInformer(cond *pipeline.PipelineRunWatchCondition) Informer {
	return NewInformer(pipeline.NewPipelineRunListWatcher(s.config, cond))
}

func (s *informerFactory) PipelineRunJobInformer(cond *pipeline.PipelineRunJobWatchCondition) Informer {
	return NewInformer(pipeline.NewPipelineRunJobListWatcher(s.config, cond))
}

func (s *informerFactory) PipelineTriggerInformer(cond *pipeline.PipelineTriggerWatchCondition) Informer {
	return NewInformer(pipeline.NewPipelineTriggerListWatcher(s.config, cond))
}

func (s *informerFactory) PipelineTriggerEventInformer(cond *pipeline.PipelineTriggerEventWatchCondition) Informer {
	return NewInformer(pipeline.NewPipelineTriggerEventListWatcher(s.config, cond))
}

func (s *informerFactory) PipelineCodeCacheInformer(cond *pipeline.PipelineCodeCacheWatchCondition) Informer {
	return NewInformer(pipeline.NewPipelineCodeCacheListWatcher(s.config, cond))
}

func (s *informerFactory) SpaceletInformer(cond *spacelet.SpaceletWatchCondition) Informer {
	return NewInformer(spacelet.NewSpaceletListWatcher(s.config, cond))
}
