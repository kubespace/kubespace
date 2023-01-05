package informer

import (
	"github.com/kubespace/kubespace/pkg/informer/cluster"
	listwatcher_cluster "github.com/kubespace/kubespace/pkg/model/listwatcher/cluster"
	"github.com/kubespace/kubespace/pkg/model/listwatcher/config"
)

type InformerFactory interface {
	ClusterAgentInformer(token string) cluster.AgentInformer
}

type informerFactory struct {
	config                *config.ListWatcherConfig
	clusterAgentInformers map[string]cluster.AgentInformer
}

func NewInformerFactory(config *config.ListWatcherConfig) InformerFactory {
	return &informerFactory{
		config: config,
	}
}

func (s *informerFactory) ClusterAgentInformer(token string) cluster.AgentInformer {
	c, ok := s.clusterAgentInformers[token]
	if ok {
		return c
	}
	agentListWatcher := listwatcher_cluster.NewAgentListWatcher(token, s.config)
	c = cluster.NewAgentInformer(token, agentListWatcher)
	s.clusterAgentInformers[token] = c
	return c
}
