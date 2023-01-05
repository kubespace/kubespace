package kube_resource

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	oredis "github.com/kubespace/kubespace/pkg/core/redis"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"k8s.io/klog"
	"time"
)

type MiddleMessage struct {
	options *oredis.Options
	client  *redis.Client
	context.Context
}

func NewMiddleMessage(op *oredis.Options) *MiddleMessage {
	client := oredis.NewRedisClient(op)
	var ctx = context.Background()

	return &MiddleMessage{
		options: op,
		client:  client,
		Context: ctx,
	}
}

func NewMiddleMessageWithClient(op *oredis.Options, client *redis.Client) *MiddleMessage {
	var ctx = context.Background()

	return &MiddleMessage{
		options: op,
		client:  client,
		Context: ctx,
	}
}

func (m *MiddleMessage) Close() {
	m.client.Close()
}

func (m *MiddleMessage) ClusterRequestQueueKey(cluster string) string {
	return "osp:cluster_request:" + cluster
}

func (m *MiddleMessage) ReceiveRequest(cluster string, reqHandle func(*MiddleRequest)) error {
	reqSubKey := m.ClusterRequestQueueKey(cluster)
	pubsub := m.client.Subscribe(m.Context, reqSubKey)
	defer pubsub.Close()
	klog.V(1).Infof("start receive pubsub %s message", cluster)
	for {
		data, err := pubsub.ReceiveMessage(m.Context)
		if err != nil {
			klog.Errorf("receive message error: %s", err.Error())
			return err
		}
		klog.V(1).Info(data.Payload)
		mr, err := UnserializerMiddleRequest(data.Payload)
		if err != nil {
			klog.Errorf("unserialzier request error: %s", err.Error())
			continue
		}
		reqHandle(mr)
	}
}

func (m *MiddleMessage) ClusterConnected(cluster string) bool {
	reqPubKey := m.ClusterRequestQueueKey(cluster)
	subNums, err := m.client.PubSubNumSub(m.Context, reqPubKey).Result()
	if err != nil {
		klog.Errorf("get cluster %s pubsub error: %s", cluster, err.Error())
		return false
	}
	if num, ok := subNums[reqPubKey]; !ok || num <= 0 {
		klog.Infof("cluster %s is not in subscribe", cluster)
		return false
	}
	return true
}

func (m *MiddleMessage) SendRequest(request *MiddleRequest) *utils.Response {
	clusterConnect := m.ClusterConnected(request.Cluster)
	if !clusterConnect {
		return &utils.Response{Code: code.RequestError, Msg: "connect kubernetes agent error"}
	}
	reqPubKey := m.ClusterRequestQueueKey(request.Cluster)
	reqData, err := request.Serializer()
	if err != nil {
		klog.Errorf("middle request error: %s", err.Error())
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	_, err = m.client.Publish(m.Context, reqPubKey, reqData).Result()
	if err != nil {
		klog.Error("publish request error: %s", err.Error())
		return &utils.Response{Code: code.RedisError, Msg: err.Error()}
	}

	resData, err := m.client.BLPop(m.Context, time.Duration(request.Timeout)*time.Second, request.RequestId).Result()
	klog.V(1).Info(resData)
	if len(resData) < 2 {
		klog.Errorf("get cluster %s response empty", request.Cluster)
		return &utils.Response{Code: code.RequestError, Msg: "request kubernetes agent timeout"}
	}
	if err != nil {
		klog.Errorf("get response error: %s", err.Error())
		return &utils.Response{Code: code.RedisError, Msg: err.Error()}
	}
	var resp utils.Response
	json.Unmarshal([]byte(resData[1]), &resp)

	return &resp
}

func (m *MiddleMessage) SendResponse(midRes *MiddleResponse) {
	reqId := midRes.RequestId
	serData, _ := midRes.Serializer()
	pipeLine := m.client.Pipeline()
	pipeLine.LPush(m.Context, reqId, serData)
	pipeLine.Expire(m.Context, reqId, time.Second*3)
	pipeLine.Exec(m.Context)
}

func (m *MiddleMessage) GlobalWatchQueueKey() string {
	return "osp:global_watch"
}

func (m *MiddleMessage) HasGlobalWatchReceive() bool {
	watchKey := m.GlobalWatchQueueKey()
	subNums, err := m.client.PubSubNumSub(m.Context, watchKey).Result()
	if err != nil {
		klog.Errorf("get pubsub %s error: %s", watchKey, err.Error())
		return false
	}
	if num, ok := subNums[watchKey]; !ok || num <= 0 {
		klog.Errorf("watch global %s is not in subscribe")
		return false
	}
	return true
}

func (m *MiddleMessage) ReceiveGlobalWatch(reqHandle func(string)) error {
	reqSubKey := m.GlobalWatchQueueKey()
	pubsub := m.client.Subscribe(m.Context, reqSubKey)
	defer pubsub.Close()
	klog.V(1).Infof("start receive global pubsub message")
	for {
		data, err := pubsub.ReceiveMessage(m.Context)
		if err != nil {
			klog.Errorf("receive message error: %s", err.Error())
			return err
		}
		//klog.V(1).Info(data.Payload)
		reqHandle(data.Payload)
	}
}

func (m *MiddleMessage) SendGlobalWatch(data interface{}) {
	watchPubKey := m.GlobalWatchQueueKey()
	subNums, err := m.client.PubSubNumSub(m.Context, watchPubKey).Result()
	if err != nil {
		klog.Errorf("get pubsub %s error: %s", watchPubKey, err.Error())
		return
	}
	if num, ok := subNums[watchPubKey]; !ok || num <= 0 {
		klog.Errorf("watch global is not in subscribe")
		return
	}
	respData, err := json.Marshal(data)
	if err != nil {
		klog.Errorf("watch response serializer error: %s", err.Error())
		return
	}
	_, err = m.client.Publish(m.Context, watchPubKey, respData).Result()
	if err != nil {
		klog.Error("publish watch response error: %s", err.Error())
		return
	}
	//klog.V(1).Infof("send global watch %s", string(respData))
}

func (m *MiddleMessage) ClusterWatchQueueKey(cluster string) string {
	return "osp:cluster_watch:" + cluster
}

func (m *MiddleMessage) HasWatchReceive(cluster string) bool {
	watchKey := m.ClusterWatchQueueKey(cluster)
	subNums, err := m.client.PubSubNumSub(m.Context, watchKey).Result()
	if err != nil {
		klog.Errorf("get pubsub %s error: %s", watchKey, err.Error())
		return false
	}
	if num, ok := subNums[watchKey]; !ok || num <= 0 {
		klog.Errorf("watch cluster %s is not in subscribe", cluster)
		return false
	}
	return true
}

func (m *MiddleMessage) ReceiveWatch(cluster string, reqHandle func(string)) error {
	reqSubKey := m.ClusterWatchQueueKey(cluster)
	pubsub := m.client.Subscribe(m.Context, reqSubKey)
	defer pubsub.Close()
	klog.V(1).Infof("start receive pubsub %s message", cluster)
	for {
		data, err := pubsub.ReceiveMessage(m.Context)
		if err != nil {
			klog.Errorf("receive message error: %s", err.Error())
			return err
		}
		klog.V(1).Info(data.Payload)
		reqHandle(data.Payload)
	}
}

func (m *MiddleMessage) SendWatch(cluster string, midRes *MiddleResponse) {
	watchPubKey := m.ClusterWatchQueueKey(cluster)
	subNums, err := m.client.PubSubNumSub(m.Context, watchPubKey).Result()
	if err != nil {
		klog.Errorf("get pubsub %s error: %s", watchPubKey, err.Error())
		return
	}
	if num, ok := subNums[watchPubKey]; !ok || num <= 0 {
		klog.Infof("watch cluster %s is not in subscribe", cluster)
		req := NewMiddleRequest(cluster, WatchType, GetAction, map[string]interface{}{"action": "close"}, 0)
		m.SendRequest(req)
		return
	}
	respData, err := midRes.Serializer()
	if err != nil {
		klog.Errorf("watch response serializer error: %s", err.Error())
		return
	}
	_, err = m.client.Publish(m.Context, watchPubKey, respData).Result()
	if err != nil {
		klog.Error("publish watch response error: %s", err.Error())
		return
	}
}

func (m *MiddleMessage) ClusterTermQueueKey(sessionId string) string {
	return "osp:term:" + sessionId
}

func (m *MiddleMessage) ReceiveTerm(sessionId string, reqHandle func(string)) error {
	termKey := m.ClusterTermQueueKey(sessionId)
	klog.V(1).Infof("start receive term %s message", termKey)
	for {
		data, err := m.client.BRPop(m.Context, time.Duration(0), termKey).Result()
		if err != nil {
			klog.Error("receive term data error: ", err.Error())
			return err
		}
		klog.Info("receive data ", data[1])
		reqHandle(data[1])
	}
}

func (m *MiddleMessage) SendTerm(midRes *MiddleResponse) {
	reqId := midRes.RequestId
	sessionKey := m.ClusterTermQueueKey(reqId)
	pipeLine := m.client.Pipeline()
	pipeLine.LPush(m.Context, sessionKey, midRes.Data)
	pipeLine.Expire(m.Context, sessionKey, time.Second*3)
	pipeLine.Exec(m.Context)
}

func (m *MiddleMessage) ClusterLogQueueKey(sessionId string) string {
	return "osp:log:" + sessionId
}

func (m *MiddleMessage) ReceiveLog(sessionId string, reqHandle func(string)) error {
	termKey := m.ClusterLogQueueKey(sessionId)
	klog.V(1).Infof("start receive log %s message", termKey)
	for {
		data, err := m.client.BRPop(m.Context, time.Duration(0), termKey).Result()
		if err != nil {
			klog.Error("receive log data error: ", err.Error())
			return err
		}
		klog.V(5).Info("receive log ", data[1])
		reqHandle(data[1])
	}
}

func (m *MiddleMessage) SendLog(midRes *MiddleResponse) {
	reqId := midRes.RequestId
	sessionKey := m.ClusterLogQueueKey(reqId)
	pipeLine := m.client.Pipeline()
	pipeLine.LPush(m.Context, sessionKey, midRes.Data)
	pipeLine.Expire(m.Context, sessionKey, time.Second*3)
	pipeLine.Exec(m.Context)
}
