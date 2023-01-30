package websockets

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/kubespace/kubespace/pkg/core/redis"
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"k8s.io/klog/v2"
)

type ApiWebsocket struct {
	redisOptions *redis.Options
	wsConn       *websocket.Conn
	*kube_resource.KubeResources
	sendChan        chan []byte
	watchCluster    string
	closeChan       chan struct{}
	globalMiddleMsg *kube_resource.MiddleMessage
	stopped         bool
}

func NewApiWebsocket(ws *websocket.Conn, redisOp *redis.Options, kr *kube_resource.KubeResources) *ApiWebsocket {
	return &ApiWebsocket{
		redisOptions:    redisOp,
		wsConn:          ws,
		KubeResources:   kr,
		sendChan:        make(chan []byte),
		globalMiddleMsg: kube_resource.NewMiddleMessage(redisOp),
		stopped:         false,
	}
}

func (a *ApiWebsocket) Consume() {
	klog.V(1).Info("start consume api ")
	go a.WsReceiveMsg()
	go a.receiveGlobalMsg()
	go a.writeMsg()
}

func (a *ApiWebsocket) WsReceiveMsg() {
	defer a.wsConn.Close()
	middleMessage := kube_resource.NewMiddleMessage(a.redisOptions)
	stopWatch := false
	startWatch := false
	stopChan := make(chan struct{})
	defer func() {
		a.stopped = true
		a.globalMiddleMsg.Close()
		stopWatch = true
		middleMessage.Close()
		if startWatch {
			<-stopChan
		}
		klog.V(1).Info("end receive watch")
	}()
	klog.V(1).Info("start read message")
	for {
		_, data, err := a.wsConn.ReadMessage()
		if err != nil {
			klog.Error("read err:", err)
			break
		}
		klog.V(1).Infof("read data: %s", string(data))
		var apiMsg ApiMsg
		err = json.Unmarshal(data, &apiMsg)
		if err != nil {
			klog.Error("json error")
			continue
		}
		if apiMsg.Action == "watchCluster" {
			stopWatch = true
			clusterMsg := apiMsg.Params.(map[string]interface{})
			cluster := clusterMsg["cluster"].(string)
			if cluster == a.watchCluster {
				continue
			}
			if startWatch {
				middleMessage.Close()
				<-stopChan
				middleMessage = kube_resource.NewMiddleMessage(a.redisOptions)
			}
			if cluster != "" {
				startWatch = true
				go func() {
					stopWatch = false
					resp := a.Watch.OpenWatch(cluster, nil)
					if !resp.IsSuccess() {
						startWatch = false
						klog.Errorf("open watch error: %s", resp.Msg)
						return
					}
					a.watchCluster = cluster
					for !stopWatch {
						klog.V(1).Info("start receive watch data")
						middleMessage.ReceiveWatch(cluster, func(data string) {
							a.sendChan <- []byte(data)
						})
					}
					a.Watch.CloseWatch(cluster)
					klog.Info("end receive")
					stopChan <- struct{}{}
					klog.Info("end receive watch data")
				}()
			} else {
				startWatch = false
			}
		}
	}
	klog.V(1).Info("end receive api websocket")
}

func (a *ApiWebsocket) receiveGlobalMsg() {
	klog.V(1).Infof("start receive global cluster msg")
	for !a.stopped {
		a.globalMiddleMsg.ReceiveGlobalWatch(func(data string) {
			a.sendChan <- []byte(data)
		})
	}
	klog.V(1).Infof("end receive global cluster msg")
}

func (a *ApiWebsocket) writeMsg() {
	for {
		select {
		case msg, ok := <-a.sendChan:
			if ok {
				err := a.wsConn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					klog.Errorf("write api websocket error: %s", err.Error())
				}
			}
		case <-a.closeChan:
			klog.V(1).Info("write websocket msg close")
			return
		}
	}
}

type ApiMsg struct {
	Action string      `json:"action"`
	Params interface{} `json:"params"`
}
