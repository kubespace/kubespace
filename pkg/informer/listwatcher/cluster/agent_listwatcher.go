package cluster

import (
	"encoding/json"
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/config"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/storage"
	kubetypes "github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/utils"
)

const ClusterAgentWatchKey = "kubespace:cluster:agent:%s"

type AgentListWatcher interface {
	listwatcher.Interface
	Watched() (bool, error)
	NotifyResult(req *kubetypes.Request, timeout int) *utils.Response
	NotifyWatch(traceId string, stopCh <-chan struct{}) <-chan []byte
	NotifyResponse(resp *kubetypes.Response) error
}

type agentListWatcher struct {
	storage.Storage
	token  string
	config *config.ListWatcherConfig
}

func NewAgentListWatcher(token string, config *config.ListWatcherConfig) AgentListWatcher {
	watchKey := fmt.Sprintf(ClusterAgentWatchKey, token)
	a := &agentListWatcher{
		token:  token,
		config: config,
	}
	resyncSec := 0
	a.Storage = config.NewStorage(watchKey, nil, nil, &resyncSec, &kubetypes.Request{})
	return a
}

func (a *agentListWatcher) NotifyResult(req *kubetypes.Request, timeout int) *utils.Response {
	resBytes, err := a.Storage.NotifyResult(req.TraceId, timeout, req)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	var resp utils.Response
	if err = json.Unmarshal(resBytes, &resp); err != nil {
		return &utils.Response{Code: code.UnMarshalError, Msg: fmt.Sprintf("unmarshal response error: %s", err.Error())}
	}
	return &resp
}

func (a *agentListWatcher) NotifyResponse(resp *kubetypes.Response) error {
	if str, ok := resp.Data.(string); ok {
		return a.Storage.NotifyResponse(resp.TraceId, []byte(str))
	}
	if bytes, ok := resp.Data.([]byte); ok {
		return a.Storage.NotifyResponse(resp.TraceId, bytes)
	}
	respBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return err
	}
	return a.Storage.NotifyResponse(resp.TraceId, respBytes)
}
