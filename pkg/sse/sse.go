package sse

import (
	"encoding/json"
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/redis"
	"k8s.io/klog"
)

var Stream *stream

type StreamClient struct {
	// 客户端唯一标识id
	ClientId string
	// 客户端监听的对象类型
	WatchType EventType
	// 客户端监听事件的对象key
	WatchKey string
	// 发送到客户端的channel
	ClientChan chan Event
}

type stream struct {
	// 新的客户端监听channel
	NewClient chan StreamClient
	// 客户端关闭监听channel
	CloseClient chan StreamClient
	// 所有的监听客户端，key为clientId
	Clients map[string]StreamClient

	DBClients map[string]StreamClient
	DBEvents  chan Event

	ClusterClients map[string]StreamClient
	ClusterEvent   chan Event

	ClusterWatch map[string]*kube_resource.MiddleMessage

	redisOptions *redis.Options
}

func NewStream(redisOp *redis.Options) *stream {
	s := &stream{
		NewClient:      make(chan StreamClient),
		CloseClient:    make(chan StreamClient),
		Clients:        make(map[string]StreamClient),
		DBClients:      make(map[string]StreamClient),
		DBEvents:       make(chan Event),
		ClusterClients: make(map[string]StreamClient),
		ClusterEvent:   make(chan Event),
		redisOptions:   redisOp,
	}
	go s.dbWatch()
}

func (s *stream) Listen() {
	for {
		select {
		case client := <-s.NewClient:
			if client.ClientId != "" {
				s.Clients[client.ClientId] = client
			} else {
				klog.Error("receive new sse clients, but no client id")
			}
		case client := <-s.CloseClient:
			if client.ClientId != "" {
				delete(s.Clients, client.ClientId)
			} else {
				klog.Error("receive close sse clients, but no client id")
			}
		case event := <-s.DBEvents:
			for _, client := range s.DBClients {
				if client.WatchType == event.Type && client.WatchKey == event.Key {
					client.ClientChan <- event
				}
			}
		case event := <-s.ClusterEvent:
			for _, client := range s.ClusterClients {
				if client.WatchType == event.Type && client.WatchKey == event.Key {
					client.ClientChan <- event
				}
			}
		}
	}
}

func (s *stream) dbWatch() {
	dbMiddleMsg := kube_resource.NewMiddleMessage(s.redisOptions)
	for {
		err := dbMiddleMsg.ReceiveGlobalWatch(func(res string) {
			var event Event
			err := json.Unmarshal([]byte(res), &event)
			if err != nil {
				klog.Errorf("unmarshal db message [%s] error: %s", res, err.Error())
			} else {
				s.DBEvents <- event
			}
		})
		if err != nil {
			klog.Errorf("receive global watch error: %s", err.Error())
		}
	}
}

func (s *stream) clusterWatch(cluster string) {
	for {
		err := dbMiddleMsg.ReceiveWatch(cluster, func(res string) {
			var event Event
			err := json.Unmarshal([]byte(res), &event)
			if err != nil {
				klog.Errorf("unmarshal db message [%s] error: %s", res, err.Error())
			} else {
				s.DBEvents <- event
			}
		})
		if err != nil {
			klog.Errorf("receive global watch error: %s", err.Error())
		}
	}
}
