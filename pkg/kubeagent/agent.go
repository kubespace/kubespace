package kubeagent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kubespace/kubespace/pkg/kubeagent/config"
	"github.com/kubespace/kubespace/pkg/kubernetes"
	kubeconfig "github.com/kubespace/kubespace/pkg/kubernetes/config"
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
	kubeConfig     *kubeconfig.KubeConfig
	tunnel         Tunnel
	kubeFactory    kubernetes.KubeFactory
	sessionWriters sync.Map
	serverCli      *utils.HttpClient
}

func NewAgent(config *config.AgentConfig) *Agent {
	a := &Agent{
		config:      config,
		kubeConfig:  config.KubeConfig,
		kubeFactory: kubernetes.NewKubeFactory(config.KubeConfig),
		serverCli:   config.ServerClient,
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

// OnSuccess agent在每个集群中独立运行，当重新连接tunnel后，server有可能更新到最新版本，
// agent从server下载当前版本匹配的yaml，并更新；
func (a *Agent) OnSuccess() {
	if a.kubeConfig.Client.RestConfig().BearerToken == "" {
		// 没有bearerToken表示agent未运行在集群pod中，不更新agent
		return
	}
	bytesBuf := new(bytes.Buffer)
	_, err := a.serverCli.Get("/import/agent/"+a.config.Token, nil, bytesBuf)
	if err != nil {
		klog.Errorf("get server agent yaml error: %s", err.Error())
		return
	}
	resHandler, err := a.kubeFactory.GetResource(kubetypes.ClusterType)
	if err != nil {
		klog.Errorf("get kubernetes cluster resource error: " + err.Error())
		return
	}
	klog.Infof("start apply agent yaml: %s", bytesBuf.String())
	resp := resHandler.Handle(kubetypes.ApplyAction, resource.ApplyParams{
		YamlStr: bytesBuf.String(),
	})
	if resp.IsSuccess() {
		klog.Infof("apply agent yaml success")
	} else {
		klog.Errorf("apply agent yaml error: %+v", resp)
	}
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
