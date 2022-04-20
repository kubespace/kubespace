package sse

import (
	"encoding/json"
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/redis"
	"k8s.io/klog"
	"time"
)

var Stream *stream

type StreamClient struct {
	// 客户端唯一标识id
	ClientId      string
	Catalog       string
	WatchSelector map[string]string
	Cluster       string
	//// 客户端监听的对象类型
	//WatchType EventType
	//// 客户端监听事件的对象key
	//WatchKey string
	// 发送到客户端的channel
	ClientChan chan Event
}

type stream struct {
	// 新的客户端监听channel
	NewClient chan StreamClient
	// 客户端关闭监听channel
	CloseClient chan StreamClient
	// 所有的监听客户端，key为clientId
	Clients    map[string]StreamClient
	EventsChan chan Event

	ClusterWatch map[string]*kube_resource.MiddleMessage

	redisOptions  *redis.Options
	kubeResources *kube_resource.KubeResources
}

func NewStream(redisOp *redis.Options, kr *kube_resource.KubeResources) *stream {
	s := &stream{
		NewClient:     make(chan StreamClient),
		CloseClient:   make(chan StreamClient),
		Clients:       make(map[string]StreamClient),
		EventsChan:    make(chan Event),
		redisOptions:  redisOp,
		kubeResources: kr,
	}
	go s.Listen()
	go s.dbWatch()
	return s
}

func (s *stream) Stream() {

}

func (s *stream) AddClient(client StreamClient) {
	klog.Infof("add new client %s", client.ClientId)
	s.NewClient <- client
}

func (s *stream) RemoveClient(client StreamClient) {
	klog.Infof("delete client %s", client.ClientId)
	s.CloseClient <- client
}

func (s *stream) Listen() {
	for {
		select {
		case client := <-s.NewClient:
			if client.ClientId == "" {
				klog.Errorf("receive new sse clients, but no client id")
				break
			}
			if client.WatchSelector == nil {
				klog.Errorf("receive new sse clients, but no watch selector")
				break
			}
			if client.Catalog == CatalogCluster {
				if client.Cluster == "" {
					klog.Errorf("receive new sse clients, but no cluster")
					break
				}
				if _, ok := client.WatchSelector[EventLabelType]; !ok {
					klog.Errorf("receive new sse clients, but no watch type")
					break
				}
				if _, ok := s.ClusterWatch[client.Cluster]; !ok {
					s.ClusterWatch[client.Cluster] = kube_resource.NewMiddleMessage(s.redisOptions)
					go s.clusterWatch(client.Cluster)
				}
				s.Clients[client.ClientId] = client
				s.clusterTypeWatch(client.Cluster)
			} else {
				s.Clients[client.ClientId] = client
			}
			klog.Infof("add new sse client %s", client.ClientId)
			klog.Infof("%v", s.Clients)

		case client := <-s.CloseClient:
			if client.ClientId == "" {
				klog.Errorf("receive new sse clients, but no client id")
				break
			}
			if _, ok := s.Clients[client.ClientId]; ok {
				delete(s.Clients, client.ClientId)
				if client.Catalog == CatalogCluster {
					hasSameCluster := false
					for _, c := range s.Clients {
						if c.Cluster == client.Cluster {
							hasSameCluster = true
							break
						}
					}
					if !hasSameCluster {
						s.ClusterWatch[client.Cluster].Close()
						delete(s.ClusterWatch, client.Cluster)
					}
					s.clusterTypeWatch(client.Cluster)

				}
			}
			klog.Infof("delete sse client %s", client.ClientId)
			klog.Infof("%v", s.Clients)

		case event := <-s.EventsChan:
			for _, client := range s.Clients {
				flag := true
				for sk, sv := range client.WatchSelector {
					found := false
					for lk, lv := range event.Labels {
						if sk == lk && sv == lv {
							found = true
							break
						}
					}
					if !found {
						flag = false
						break
					}
				}
				if flag {
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
				s.EventsChan <- event
			}
		})
		if err != nil {
			klog.Errorf("receive global watch error: %s", err.Error())
		}
		time.Sleep(10 * time.Second)
	}
}

func (s *stream) clusterTypeWatch(cluster string) {
	watchTypes := make(map[string]struct{})
	for _, client := range s.Clients {
		if client.Catalog != CatalogCluster {
			continue
		}
		if client.Cluster == cluster {
			if _, ok := client.WatchSelector[EventLabelType]; ok {
				watchTypes[client.WatchSelector[EventLabelType]] = struct{}{}
			}
		}
	}
	var types []string
	for t, _ := range watchTypes {
		types = append(types, t)
	}
	resp := s.kubeResources.Watch.OpenWatch(cluster, types)
	if !resp.IsSuccess() {
		klog.Errorf("open watch error: %s", resp.Msg)
		return
	}
	//s.kubeResources.Watch
	klog.Infof("cluster %s watch types %v", cluster, types)
}

func (s *stream) clusterWatch(cluster string) {
	for {
		if clusterMiddle, ok := s.ClusterWatch[cluster]; ok {
			err := clusterMiddle.ReceiveWatch(cluster, func(res string) {
				var event Event
				err := json.Unmarshal([]byte(res), &event)
				if err != nil {
					klog.Errorf("unmarshal db message [%s] error: %s", res, err.Error())
				} else {
					s.EventsChan <- event
				}
			})
			if err != nil {
				klog.Errorf("receive global watch error: %s", err.Error())
			}
		} else {
			klog.Infof("close cluster %s watch", cluster)
			break
		}
	}
}
