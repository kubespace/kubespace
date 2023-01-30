package kubeagent

import (
	"encoding/json"
	"fmt"
	"github.com/kubespace/kubespace/pkg/kubeagent/config"
	"github.com/kubespace/kubespace/pkg/kubernetes"
	"github.com/kubespace/kubespace/pkg/kubernetes/resource"
	kubetypes "github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"k8s.io/klog/v2"
	"runtime"
	"sync"
)

type Agent struct {
	config         *config.AgentConfig
	tunnel         Tunnel
	kubeFactory    kubernetes.KubeFactory
	sessionWriters sync.Map
}

func NewAgent(config *config.AgentConfig) *Agent {
	a := &Agent{
		config:      config,
		kubeFactory: kubernetes.NewKubeFactory(config.KubeConfig),
	}
	a.tunnel = NewTunnel(config.Token, config.ServerHost, a)
	return a
}

func (a *Agent) Run(stopCh <-chan struct{}) {
	go a.tunnel.Run(stopCh)
	for {
		select {
		case receiveBytes := <-a.tunnel.Receive():
			go a.Handle(receiveBytes)
		case <-stopCh:
			break
		}
	}
}

func (a *Agent) Handle(receiveBytes []byte) {
	//klog.Infof(string(receiveBytes))
	var req kubetypes.Request
	var resp *utils.Response
	if err := json.Unmarshal(receiveBytes, &req); err != nil {
		klog.Errorf("agent receive unexpected request: %s", string(receiveBytes))
		resp = &utils.Response{Code: code.UnMarshalError, Msg: err.Error()}
	} else {
		resp = a.handle(&req)
	}
	a.tunnel.Send(&kubetypes.Response{TraceId: req.TraceId, Data: resp})
}

func (a *Agent) handle(req *kubetypes.Request) (resp *utils.Response) {
	defer func() {
		if err := recover(); err != nil {
			klog.Error("do request error: ", err)
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			klog.Errorf("==> %s\n", string(buf[:n]))
			msg := fmt.Sprintf("%s", err)
			resp = &utils.Response{Code: "UnknownError", Msg: msg}
		}
	}()
	switch {
	case req.Resource == kubetypes.PodType && (req.Action == kubetypes.ExecAction || req.Action == kubetypes.LogAction):
		if podHandler, err := a.kubeFactory.GetPod(); err != nil {
			resp = &utils.Response{Code: code.GetError, Msg: err.Error()}
		} else {
			writer := newSessionWriter(req.TraceId, a.tunnel)
			a.sessionWriters.Store(req.TraceId, writer)
			if req.Action == kubetypes.ExecAction {
				resp = podHandler.Exec(req.Params, writer)
			} else if req.Action == kubetypes.LogAction {
				resp = podHandler.Log(req.Params, writer)
			}
		}
	case req.Action == kubetypes.CloseSession:
		obj, ok := a.sessionWriters.LoadAndDelete(req.TraceId)
		if ok {
			writer, ok := obj.(resource.OutWriter)
			if ok {
				writer.Close()
			}
		}
		resp = &utils.Response{Code: code.Success}
	case req.Action == kubetypes.WatchAction:
		if resHandler, err := a.kubeFactory.GetResource(req.Resource); err != nil {
			resp = &utils.Response{Code: code.GetError, Msg: err.Error()}
		} else {
			writer := newSessionWriter(req.TraceId, a.tunnel)
			a.sessionWriters.Store(req.TraceId, writer)
			resp = resHandler.Watch(req.Params, writer)
		}
	default:
		if resHandler, err := a.kubeFactory.GetResource(req.Resource); err != nil {
			resp = &utils.Response{Code: code.GetError, Msg: err.Error()}
		} else {
			resp = resHandler.Handle(req.Action, req.Params)
		}
	}
	return
}

func (a *Agent) OnSuccess() {

}

type sessionWriter struct {
	traceId string
	tunnel  Tunnel
	stopCh  chan struct{}
	stopped bool
}

func newSessionWriter(traceId string, tunnel Tunnel) resource.OutWriter {
	return &sessionWriter{
		traceId: traceId,
		tunnel:  tunnel,
		stopCh:  make(chan struct{}),
	}
}

func (s *sessionWriter) Write(out interface{}) error {
	s.tunnel.Send(&kubetypes.Response{TraceId: s.traceId, Data: out})
	return nil
}

func (s *sessionWriter) StopCh() <-chan struct{} {
	return s.stopCh
}

func (s *sessionWriter) Close() {
	if s.stopped {
		return
	}
	s.stopped = true
	close(s.stopCh)
}
