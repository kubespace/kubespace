package cluster

import (
	"github.com/gorilla/websocket"
	"github.com/kubespace/kubespace/pkg/informer"
	"k8s.io/klog/v2"
	"sync"
)

// agentTunnel kube-agent服务起来后发送连接请求，kubespace与agent建立websocket隧道
// agent_informer监听到要发送的请求数据后，从隧道发出，随后kube-agent从隧道接收到请求并处理
type agentTunnel struct {
	informer informer.Informer
	wsConn   *websocket.Conn
	stopCh   chan struct{}
	mu       *sync.Mutex
}

func newAgentTunnel(wsConn *websocket.Conn, clusterAgentInformer informer.Informer) (*agentTunnel, error) {
	k := &agentTunnel{
		informer: clusterAgentInformer,
		wsConn:   wsConn,
		stopCh:   make(chan struct{}),
		mu:       &sync.Mutex{},
	}
	k.informer.AddHandler(k)
	return k, nil
}

func (k *agentTunnel) Consume() {
	go k.informer.Run(k.stopCh)
	go k.consume()
}

func (k *agentTunnel) Check(obj interface{}) bool {
	return true
}

func (k *agentTunnel) Handle(obj interface{}) error {
	// websocket同时写会报错
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.wsConn.WriteJSON(obj)
}

func (k *agentTunnel) consume() {
	for {
		_, data, err := k.wsConn.ReadMessage()
		if err != nil {
			klog.Error("read err:", err)
			break
		}
		klog.V(1).Infof("read data: %s", string(data))
	}
	close(k.stopCh)
}
