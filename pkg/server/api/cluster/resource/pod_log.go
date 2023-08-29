package resource

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/service/cluster"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
	"net/http"
)

type podLogHandler struct {
	models     *model.Models
	kubeClient *cluster.KubeClient
}

func PodLogHandler(conf *config.ServerConfig) api.Handler {
	return &podLogHandler{
		models:     conf.Models,
		kubeClient: conf.ServiceFactory.Cluster.KubeClient,
	}
}

func (h *podLogHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	if c.Query("project_id") != "" {
		projectId, err := utils.ParseUint(c.Query("project_id"))
		if err != nil {
			return true, nil, errors.New(code.ParamsError, err)
		}
		return true, &api.AuthPerm{
			Scope:   types.ScopeProject,
			ScopeId: projectId,
			Role:    types.RoleViewer,
		}, nil
	}
	clusterId, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopeCluster,
		ScopeId: clusterId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *podLogHandler) Handle(c *api.Context) *utils.Response {
	upGrader := &websocket.Upgrader{}
	upGrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		klog.Errorf("upgrader agent conn error: %s", err)
		return nil
	}
	podlog, err := newPodLog(ws, h.kubeClient, c.Param("id"), &podLogParams{
		Namespace: c.Param("namespace"),
		Name:      c.Param("pod"),
		Container: c.Query("container"),
	})
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		ws.Close()
		return nil
	}
	go podlog.consume()
	return nil
}

type podLogParams struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Container string `json:"container"`
}

type podLog struct {
	ws        *websocket.Conn
	logOuter  cluster.Outer
	clusterId string
	params    *podLogParams
	stopCh    chan struct{}
	stopped   bool
}

func newPodLog(ws *websocket.Conn, client *cluster.KubeClient, clusterId string, params *podLogParams) (*podLog, error) {
	if clusterId == "" {
		return nil, fmt.Errorf("param clusterId is empty")
	}
	if params.Namespace == "" {
		params.Namespace = "default"
	}
	if params.Name == "" {
		return nil, fmt.Errorf("param name is empty")
	}
	if params.Container == "" {
		return nil, fmt.Errorf("params container is empty")
	}
	pods, err := client.Pods(clusterId)
	if err != nil {
		return nil, err
	}
	outer, err := pods.Log(params)
	if err != nil {
		return nil, err
	}
	return &podLog{
		ws:        ws,
		logOuter:  outer,
		clusterId: clusterId,
		params:    params,
		stopCh:    make(chan struct{}),
	}, nil
}

func (p *podLog) consume() {
	defer p.close()
	go p.read()
	for {
		select {
		case res := <-p.logOuter.OutCh():
			if str, ok := res.(string); ok {
				p.ws.WriteMessage(websocket.TextMessage, []byte(str))
			}
		case <-p.stopCh:
			// websocket连接断开
			return
		case <-p.logOuter.StopCh():
			// execStream退出
			return
		}
	}
}

func (p *podLog) read() {
	defer p.close()
	for {
		_, _, err := p.ws.ReadMessage()
		if err != nil {
			klog.Error("read err:", err)
			break
		}
	}
}

func (p *podLog) close() {
	if p.stopped {
		return
	}
	p.stopped = true
	close(p.stopCh)
	p.logOuter.Close()
	p.ws.Close()
}
