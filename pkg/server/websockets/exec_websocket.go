package websockets

import (
	"encoding/base64"
	"github.com/gorilla/websocket"
	"github.com/kubespace/kubespace/pkg/core/redis"
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog"
	"time"
)

type ExecWebsocket struct {
	redisOptions  *redis.Options
	middleMessage *kube_resource.MiddleMessage
	cluster       string
	wsConn        *websocket.Conn
	stopped       bool
	*kube_resource.KubeResources
	namespace string
	pod       string
	container string
	rows      string
	cols      string
	sessionId string
}

func NewExecWebsocket(
	cluster string,
	ws *websocket.Conn,
	redisOp *redis.Options,
	kr *kube_resource.KubeResources,
	namespace, pod, container string,
	rows, cols string) *ExecWebsocket {
	middleMsg := kube_resource.NewMiddleMessage(redisOp)
	sessionId := utils.CreateUUID()
	return &ExecWebsocket{
		cluster:       cluster,
		redisOptions:  redisOp,
		middleMessage: middleMsg,
		wsConn:        ws,
		stopped:       false,
		KubeResources: kr,
		namespace:     namespace,
		pod:           pod,
		container:     container,
		rows:          rows,
		cols:          cols,
		sessionId:     sessionId,
	}
}

func (e *ExecWebsocket) Consume() {
	klog.V(1).Info("start consume exec ", e.cluster)
	execParams := map[string]interface{}{
		"namespace":  e.namespace,
		"name":       e.pod,
		"container":  e.container,
		"session_id": e.sessionId,
		"rows":       e.rows,
		"cols":       e.cols,
	}
	resp := e.Pod.Exec(e.cluster, execParams)
	if !resp.IsSuccess() {
		e.wsConn.WriteMessage(websocket.TextMessage, []byte(resp.Msg))
		e.wsConn.Close()
		return
	}
	go e.WsReceiveMsg()
	go e.MiddleTermHandle()
}

func (e *ExecWebsocket) MiddleTermHandle() {
	klog.V(1).Infof("start receive term session %s", e.sessionId)
	for !e.stopped {
		err := e.middleMessage.ReceiveTerm(e.sessionId, func(data string) {
			d, err := base64.StdEncoding.DecodeString(data)
			if err != nil {
				klog.Errorf("write cluster %s decode term data error: %s", e.cluster, err.Error())
			} else {
				err = e.wsConn.WriteMessage(websocket.TextMessage, d)
				if err != nil {
					klog.Errorf("write cluster %s terminal websocket error: %s", e.cluster, err.Error())
				}
			}
		})
		if err != nil {
			klog.Errorf("receive cluster %s terminal message error: %s", e.cluster, err.Error())
		}
		if !e.stopped {
			time.Sleep(5 * time.Second)
		}
	}
	klog.V(1).Infof("end receive term session %s data", e.sessionId)
}

func (e *ExecWebsocket) WsReceiveMsg() {
	defer e.Clean()
	for {
		_, data, err := e.wsConn.ReadMessage()
		if err != nil {
			klog.Error("read err:", err)
			break
		}
		klog.V(1).Infof("read data: %s", string(data))
		params := map[string]interface{}{
			"session_id": e.sessionId,
			"input":      data,
		}
		resp := e.Pod.Stdin(e.cluster, params)
		if !resp.IsSuccess() {
			klog.Error("term stdin error: ", resp.Msg)
		}
	}
}

func (e *ExecWebsocket) Clean() {
	klog.V(1).Infof("start clean cluster %s websocket", e.cluster)
	e.stopped = true
	e.middleMessage.Close()
	e.wsConn.Close()
	execParams := map[string]interface{}{
		"session_id": e.sessionId,
	}
	e.Pod.CloseExecConn(e.cluster, execParams)
	//if !resp.IsSuccess() {
	//	e.wsConn.WriteMessage(websocket.TextMessage, []byte(resp.Msg))
	//	e.wsConn.Close()
	//	return
	//}
	klog.V(1).Infof("end clean cluster %s websocket", e.cluster)
}
