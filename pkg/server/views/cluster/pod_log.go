package cluster

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/kubespace/kubespace/pkg/service/cluster"
	"k8s.io/klog/v2"
)

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
