package manager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
	"time"
)

type CommonManager struct {
	client *redis.Client
	DB     *gorm.DB

	modelKey string

	context.Context
	*kube_resource.MiddleMessage
	watch bool
}

func NewCommonManager(redisClient *redis.Client, db *gorm.DB, key string, watch bool) *CommonManager {
	middleMessage := kube_resource.NewMiddleMessageWithClient(nil, redisClient)
	return &CommonManager{
		client:        redisClient,
		DB:            db,
		modelKey:      key,
		Context:       context.Background(),
		MiddleMessage: middleMessage,
		watch:         watch,
	}
}

func (cm *CommonManager) Get(key string, keyObj interface{}) error {
	data, err := cm.client.HGetAll(cm.Context, cm.PrimaryKey(key)).Result()
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return errors.New(fmt.Sprintf("not found key:%s", key))
	}
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(jsonBody, keyObj); err != nil {
		return err
	}
	return nil
}

func (cm *CommonManager) Save(key string, keyObj interface{}, expire time.Duration, sets bool) error {

	if sets {
		if err := cm.AddSets(key); err != nil {
			return err
		}
	}

	jsonbody, _ := json.Marshal(keyObj)
	var m map[string]interface{}
	err := json.Unmarshal(jsonbody, &m)
	if err != nil {
		klog.Error(err)
		return err
	}

	if expire < 1 {
		err = cm.client.HMSet(cm.Context, cm.PrimaryKey(key), m).Err()
		if err != nil {
			klog.Error(err)
			return err
		}
	} else {
		pipeline := cm.client.Pipeline()
		pipeline.HMSet(cm.Context, cm.PrimaryKey(key), m)
		pipeline.Expire(cm.Context, cm.PrimaryKey(key), expire)
		_, err := pipeline.Exec(cm.Context)
		if err != nil {
			return err
		}
	}
	if cm.watch {

		eventObj := NewEventObj(AddEvent, cm.modelKey, keyObj)
		cm.MiddleMessage.SendGlobalWatch(eventObj)
	}

	return nil
}

func (cm *CommonManager) Update(key string, keyObj interface{}, expire time.Duration, sets bool) error {

	if sets {
		if err := cm.AddSets(key); err == nil {
			return fmt.Errorf("duplicate key")
		}
	}

	jsonbody, _ := json.Marshal(keyObj)
	var m map[string]interface{}
	err := json.Unmarshal(jsonbody, &m)
	if err != nil {
		return err
	}

	if expire < 1 {
		err = cm.client.HMSet(cm.Context, cm.PrimaryKey(key), m).Err()
		if err != nil {
			return err
		}
	} else {
		pipeline := cm.client.Pipeline()
		pipeline.HMSet(cm.Context, cm.PrimaryKey(key), m)
		pipeline.Expire(cm.Context, cm.PrimaryKey(key), expire)
		_, err := pipeline.Exec(cm.Context)
		if err != nil {
			return err
		}
	}

	if cm.watch {
		obj := make(map[string]interface{})
		cm.Get(key, &obj)

		eventObj := NewEventObj(UpdateEvent, cm.modelKey, obj)
		cm.MiddleMessage.SendGlobalWatch(eventObj)
	}

	return nil
}

func (cm *CommonManager) Delete(key string) error {
	pipeline := cm.client.Pipeline()
	pipeline.SRem(cm.Context, cm.SetsKey(), key)
	pipeline.Del(cm.Context, cm.PrimaryKey(key))
	_, err := pipeline.Exec(cm.Context)
	if err != nil {
		return err
	}

	if cm.watch {
		eventObj := NewEventObj(DeleteEvent, cm.modelKey, map[string]interface{}{"name": key})
		cm.MiddleMessage.SendGlobalWatch(eventObj)
	}
	return nil
}

func (cm *CommonManager) List(filters map[string]interface{}) ([]map[string]string, error) {
	pipeline := cm.client.Pipeline()
	es, _ := cm.client.SMembers(cm.Context, cm.SetsKey()).Result()
	for _, pk := range es {
		pipeline.HGetAll(cm.Context, cm.PrimaryKey(pk))
	}
	cms, err := pipeline.Exec(cm.Context)
	if err != nil {
		return nil, err
	}
	var dList []map[string]string
	for _, cmd := range cms {
		x, err := cmd.(*redis.StringStringMapCmd).Result()
		if err != nil {
			return nil, err
		}

		ad := true
		for k, v := range filters {
			if x[k] != v {
				ad = false
				break
			}
		}
		if ad {
			dList = append(dList, x)
		}
	}
	return dList, nil
}

func (cm *CommonManager) PrimaryKey(key string) string {
	return fmt.Sprintf("%s:%s", cm.modelKey, key)
}

func (cm *CommonManager) SetsKey() string {
	return fmt.Sprintf("%s:_sets", cm.modelKey)
}

func (cm *CommonManager) AddSets(name string) error {
	setKey := cm.SetsKey()
	if ok, _ := cm.client.SIsMember(cm.Context, setKey, name).Result(); ok {
		return errors.New(fmt.Sprintf("have added %s to sets：%s", name, setKey))
	}
	err := cm.client.SAdd(cm.Context, setKey, name).Err()
	if err != nil {
		return err
	}
	return nil
}

type EventObj struct {
	Event    string      `json:"event"`
	Obj      string      `json:"obj"`
	Resource interface{} `json:"resource"`
}

const (
	AddEvent    = "add"
	UpdateEvent = "update"
	DeleteEvent = "delete"
)

func NewEventObj(event, objKey string, obj interface{}) *EventObj {
	return &EventObj{
		Event:    event,
		Obj:      objKey,
		Resource: obj,
	}
}
