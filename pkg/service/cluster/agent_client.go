package cluster

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/cluster"
	kubetypes "github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
)

func NewTraceId() string {
	return fmt.Sprintf("kubespace:agent:response:%s", utils.CreateUUID())
}

type AgentClient struct {
	models *model.Models
}

func NewAgentClient(models *model.Models) *AgentClient {
	return &AgentClient{models: models}
}

func (a *AgentClient) request(clusterObj *types.Cluster, resType, action string, params interface{}) *utils.Response {
	if handler, err := newAgentHandler(clusterObj, a.models); err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	} else {
		return handler.Handle("", resType, action, params)
	}
}

func (a *AgentClient) watch(clusterObj *types.Cluster, resType string, params interface{}) (Outer, error) {
	handler, err := newAgentHandler(clusterObj, a.models)
	if err != nil {
		return nil, err
	}
	traceId := NewTraceId()
	resp := handler.Handle(traceId, resType, kubetypes.WatchAction, params)
	if !resp.IsSuccess() {
		return nil, fmt.Errorf(resp.Msg)
	}
	o := newAgentOuter(traceId, handler)
	return o, nil
}

func (a *AgentClient) pods(clusterObj *types.Cluster) (PodClient, error) {
	if handler, err := newAgentHandler(clusterObj, a.models); err != nil {
		return nil, err
	} else {
		return &agentPod{handler}, nil
	}
}

type agentHandler struct {
	models     *model.Models
	clusterObj *types.Cluster
	watcher    cluster.AgentListWatcher
}

func newAgentHandler(clusterObj *types.Cluster, models *model.Models) (*agentHandler, error) {
	agentListWatcher := cluster.NewAgentListWatcher(clusterObj.Token, models.ListWatcherConfig)
	if watched, err := agentListWatcher.Watched(); err != nil {
		return nil, err
	} else if !watched {
		return nil, fmt.Errorf("connect kubernetes agent error")
	}
	return &agentHandler{
		models:     models,
		clusterObj: clusterObj,
		watcher:    agentListWatcher,
	}, nil
}

func (a *agentHandler) Handle(traceId, resType, action string, params interface{}) *utils.Response {
	if traceId == "" {
		traceId = NewTraceId()
	}
	return a.watcher.NotifyResult(&kubetypes.Request{
		TraceId:  traceId,
		Resource: resType,
		Action:   action,
		Params:   params,
	}, 30)
}

func (a *agentHandler) CloseSession(traceId string) *utils.Response {
	return a.watcher.NotifyResult(&kubetypes.Request{
		TraceId: traceId,
		Action:  kubetypes.CloseSession,
	}, 30)
}

func (a *agentHandler) Watch(traceId string, stopCh <-chan struct{}) <-chan []byte {
	return a.watcher.NotifyWatch(traceId, stopCh)
}

type agentOuter struct {
	*outer
	traceId string
	handler *agentHandler
}

func newAgentOuter(traceId string, handler *agentHandler) *agentOuter {
	o := &agentOuter{
		outer:   newOuter(),
		traceId: traceId,
		handler: handler,
	}
	o.writeOut()
	return o
}

func (a *agentOuter) writeOut() {
	resultCh := a.handler.Watch(a.traceId, a.stopCh)
	go func() {
		defer a.Close()
		defer a.handler.CloseSession(a.traceId)
		for {
			select {
			case msg, ok := <-resultCh:
				if !ok {
					klog.Infof("notify watch closed")
					return
				}
				a.outCh <- string(msg)
			case <-a.stopCh:
				klog.Infof("stop write out")
				return
			}
		}
	}()
}

type agentPod struct {
	handler *agentHandler
}

func (a *agentPod) Exec(params interface{}) (PodExec, error) {
	traceId := NewTraceId()
	resp := a.handler.Handle(traceId, kubetypes.PodType, kubetypes.ExecAction, params)
	if !resp.IsSuccess() {
		return nil, fmt.Errorf(resp.Msg)
	}
	return newAgentPodExec(traceId, a.handler), nil
}

func (a *agentPod) Log(params interface{}) (Outer, error) {
	traceId := NewTraceId()
	resp := a.handler.Handle(traceId, kubetypes.PodType, kubetypes.LogAction, params)
	if !resp.IsSuccess() {
		return nil, fmt.Errorf(resp.Msg)
	}
	return newAgentOuter(traceId, a.handler), nil
}

type agentPodExec struct {
	*agentOuter
}

func newAgentPodExec(traceId string, handler *agentHandler) *agentPodExec {
	a := &agentPodExec{
		agentOuter: newAgentOuter(traceId, handler),
	}
	a.writeOut()
	return a
}

func (a *agentPodExec) Stdin(params interface{}) error {
	resp := a.handler.Handle("", kubetypes.PodType, kubetypes.StdinAction, params)
	if !resp.IsSuccess() {
		return fmt.Errorf(resp.Msg)
	}
	return nil
}
