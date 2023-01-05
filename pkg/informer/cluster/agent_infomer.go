package cluster

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/model/listwatcher/cluster"
)

type AgentInformer interface {
	Run()
}

func NewAgentInformer(token string, clusterAgentListWatcher cluster.AgentListWatcher) AgentInformer {
	return &agentInformer{
		token:            token,
		AgentListWatcher: clusterAgentListWatcher,
	}
}

type agentInformer struct {
	token string
	cluster.AgentListWatcher
}

func (a *agentInformer) WatchKey() string {
	return fmt.Sprintf("cluster:agent:%s", a.token)
}

func (a *agentInformer) Run() {

}
