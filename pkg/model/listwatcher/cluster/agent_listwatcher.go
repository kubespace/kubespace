package cluster

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/model/listwatcher/config"
	"github.com/kubespace/kubespace/pkg/model/listwatcher/watch_storage"
)

type AgentListWatcher interface {
	Stop() error
	Watch() <-chan []byte
	Notify(interface{}) error
}

const ClusterAgentWatchKey = "kubespace:cluster:agent:%s"

type agentListWatcher struct {
	watch_storage.Storage
	token  string
	config *config.ListWatcherConfig
}

func NewAgentListWatcher(token string, config *config.ListWatcherConfig) AgentListWatcher {
	watchKey := fmt.Sprintf(ClusterAgentWatchKey, token)
	return &agentListWatcher{
		token:   token,
		config:  config,
		Storage: config.NewWatchStorage(watchKey),
	}
}
