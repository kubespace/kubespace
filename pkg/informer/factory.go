package informer

import (
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/cluster"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/config"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/pipeline"
)

type Factory interface {
	ClusterAgentInformer(token string) Informer
	PipelineRunInformer(cond *pipeline.PipelineRunWatchCondition) Informer
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
