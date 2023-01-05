package websockets

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/kubespace/kubespace/pkg/core/redis"
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"k8s.io/klog"
	"time"
)

type KubeWebsocket struct {
	redisOptions  *redis.Options
	middleMessage *kube_resource.MiddleMessage
	cluster       string
	wsConn        *websocket.Conn
	stopped       bool
	models        *model.Models
}

func NewKubeWebsocket(cluster string, ws *websocket.Conn, redisOp *redis.Options, models *model.Models) (*KubeWebsocket, error) {
	middleMsg := kube_resource.NewMiddleMessage(redisOp)
	if middleMsg.HasWatchReceive(cluster) {
		return nil, fmt.Errorf("cluster %s has another agent connected", cluster)
	}
	return &KubeWebsocket{
		cluster:       cluster,
		redisOptions:  redisOp,
		middleMessage: middleMsg,
		wsConn:        ws,
		stopped:       false,
		models:        models,
	}, nil
}

func (k *KubeWebsocket) Consume() {
	klog.V(1).Info("start consume cluster ", k.cluster)
	go k.WsReceiveMsg()
	go k.MiddleRequestHandle()
}

func (k *KubeWebsocket) MiddleRequestHandle() {
	for !k.stopped {
		klog.V(1).Info("start receive request from cluster ", k.cluster)
		err := k.middleMessage.ReceiveRequest(k.cluster, func(mr *kube_resource.MiddleRequest) {
			serReq, _ := mr.Serializer()
			err := k.wsConn.WriteMessage(websocket.TextMessage, serReq)
			if err != nil {
				klog.Errorf("agent cluster=%s websocket write message error: %s", k.cluster, err.Error())
			}
		})
		if err != nil {
			klog.Errorf("receive cluster %s middle message error: %s", k.cluster, err.Error())
		}
		if !k.stopped {
			time.Sleep(5 * time.Second)
		}
	}
	klog.V(1).Info("cluster %s middle request handle end", k.cluster)
}

func (k *KubeWebsocket) WsReceiveMsg() {
	defer k.Clean()
	for {
		_, data, err := k.wsConn.ReadMessage()
		if err != nil {
			klog.Error("read err:", err)
			break
		}
		klog.V(1).Infof("read data: %s", string(data))
		midResp, err := kube_resource.UnserialzerMiddleResponse(string(data))
		if err != nil {
			klog.Errorf("unserializer data error: %s", err.Error())
			continue
		}
		if midResp.IsRequest() {
			k.middleMessage.SendResponse(midResp)
		} else if midResp.IsTerm() {
			k.middleMessage.SendTerm(midResp)
		} else if midResp.IsWatch() {
			k.middleMessage.SendWatch(k.cluster, midResp)
		} else if midResp.IsLog() {
			k.middleMessage.SendLog(midResp)
		}
	}
}

func (k *KubeWebsocket) Clean() {
	klog.V(1).Infof("start clean cluster %s websocket", k.cluster)
	clusterObj, err := k.models.ClusterManager.GetByName(k.cluster)
	if err != nil {
		klog.Errorf("get cluster %s object error: %s", k.cluster, err.Error())
	} else {
		clusterObj.Status = types.ClusterPending
		k.models.ClusterManager.Update(clusterObj)
	}
	k.middleMessage.Close()
	k.stopped = true
	k.wsConn.Close()
	klog.V(1).Infof("end clean cluster %s websocket", k.cluster)
}
