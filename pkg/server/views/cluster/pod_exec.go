package cluster

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/kubespace/kubespace/pkg/service/cluster"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
)

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
