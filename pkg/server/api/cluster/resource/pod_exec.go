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

type podExecHandler struct {
	models     *model.Models
	kubeClient *cluster.KubeClient
}

func PodExecHandler(conf *config.ServerConfig) api.Handler {
	return &podExecHandler{
		models:     conf.Models,
		kubeClient: conf.ServiceFactory.Cluster.KubeClient,
	}
}

func (h *podExecHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	if c.Query("project_id") != "" {
		projectId, err := utils.ParseUint(c.Query("project_id"))
		if err != nil {
			return true, nil, errors.New(code.ParamsError, err)
		}
		return true, &api.AuthPerm{
			Scope:   types.ScopeProject,
			ScopeId: projectId,
			Role:    types.RoleEditor,
		}, nil
	}
	clusterId, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopeCluster,
		ScopeId: clusterId,
		Role:    types.RoleEditor,
	}, nil
}

func (h *podExecHandler) Handle(c *api.Context) *utils.Response {
	upGrader := &websocket.Upgrader{}
	upGrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		klog.Errorf("upgrader agent conn error: %s", err)
		return nil
	}
	podexec, err := newPodExec(ws, h.kubeClient, c.Param("id"), &podExecParams{
		Namespace: c.Param("namespace"),
		Name:      c.Param("pod"),
		Container: c.Query("container"),
		Rows:      c.Query("rows"),
		Cols:      c.Query("cols"),
	})
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		ws.Close()
		return nil
	}
	go podexec.consume()
	return nil
}

type podExecParams struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Container string `json:"container"`
	SessionId string `json:"session_id"`
	Rows      string `json:"rows"`
	Cols      string `json:"cols"`
}

type podExec struct {
	ws        *websocket.Conn
	exec      cluster.PodExec
	clusterId string
	params    *podExecParams
	stopCh    chan struct{}
	stopped   bool
}

func newPodExec(ws *websocket.Conn, client *cluster.KubeClient, clusterId string, params *podExecParams) (*podExec, error) {
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
	if params.Cols == "" {
		params.Cols = "64"
	}
	if params.Rows == "" {
		params.Rows = "64"
	}
	if params.SessionId == "" {
		params.SessionId = utils.CreateUUID()
	}
	pods, err := client.Pods(clusterId)
	if err != nil {
		return nil, err
	}
	exec, err := pods.Exec(params)
	if err != nil {
		return nil, err
	}
	return &podExec{
		ws:        ws,
		exec:      exec,
		clusterId: clusterId,
		params:    params,
		stopCh:    make(chan struct{}),
	}, nil
}

func (p *podExec) consume() {
	defer p.close()
	go p.read()
	for {
		select {
		case res := <-p.exec.OutCh():
			if str, ok := res.(string); ok {
				p.ws.WriteMessage(websocket.TextMessage, []byte(str))
			}
		case <-p.stopCh:
			// websocket连接断开
			return
		case <-p.exec.StopCh():
			// execStream退出
			return
		}
	}
}

func (p *podExec) read() {
	defer p.close()
	for {
		_, data, err := p.ws.ReadMessage()
		if err != nil {
			klog.Error("read err:", err)
			break
		}
		klog.V(1).Infof("read data: %s", string(data))
		params := map[string]interface{}{
			"session_id": p.params.SessionId,
			"input":      data,
		}
		if err = p.exec.Stdin(params); err != nil {
			klog.Error("term stdin error: ", err.Error())
		}
	}
}

func (p *podExec) close() {
	if p.stopped {
		return
	}
	p.stopped = true
	close(p.stopCh)
	p.exec.Close()
	p.ws.Close()
}
