package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/kubespace/kubespace/pkg/core/datatype"
	"k8s.io/klog/v2"
	"time"
)

var ErrNotWatch = errors.New("watch key is not subscribed")
var ErrTimeout = errors.New("result timeout")

type RedisStorage struct {
	client     *redis.Client
	watchKey   string
	pubSub     *redis.PubSub
	resultChan chan interface{}
	watchErrCh chan error
	resyncSec  int
	listFunc   ListFunc
	stopCh     chan struct{}
	stopped    bool
	dataType   datatype.DataType
}

func NewRedisStorage(redisClient *redis.Client, watchKey string, listFunc ListFunc, resyncSec int, dataType datatype.DataType) Storage {
	return &RedisStorage{
		client:     redisClient,
		watchKey:   watchKey,
		resultChan: make(chan interface{}),
		watchErrCh: make(chan error),
		listFunc:   listFunc,
		resyncSec:  resyncSec,
		stopCh:     make(chan struct{}),
		dataType:   dataType,
	}
}

func (r *RedisStorage) Run() {
	if r.stopped {
		r.stopped = false
		r.stopCh = make(chan struct{})
	}
	if r.watchKey != "" {
		go r.watch()
	}
	if r.resyncSec > 0 && r.listFunc != nil {
		go r.list()
	}
}

func (r *RedisStorage) list() {
	ticker := time.NewTicker(time.Second * time.Duration(r.resyncSec))
	for {
		select {
		case <-ticker.C:
			objects, err := r.listFunc()
			if err != nil {
				klog.Errorf("list")
			}
			for i, _ := range objects {
				r.resultChan <- objects[i]
			}
		case <-r.stopCh:
			return
		}
	}
}

func (r *RedisStorage) watch() {
	if r.pubSub != nil {
		return
	}
	defer r.Stop()
	r.pubSub = r.client.Subscribe(context.Background(), r.watchKey)
	klog.Infof("start watch key=%s", r.watchKey)
	for {
		data, err := r.pubSub.ReceiveMessage(context.Background())
		if err != nil {
			klog.Errorf("receive message error: %s", err.Error())
			if r.stopped {
				break
			}
			// 这里还需要再思考下，watch失败不退出，让list还能维持工作
			r.watchErrCh <- err
			return
		}
		obj, err := r.dataType.Unmarshal([]byte(data.Payload))
		if err != nil {
			klog.Errorf("unmarshal receive message to data type error: %s, datatye=%v, message=%s", err.Error(), r.dataType, data.Payload)
		} else {
			r.resultChan <- obj
		}
	}
	klog.Errorf("stopped watch key=%s", r.watchKey)
}

func (r *RedisStorage) Result() <-chan interface{} {
	return r.resultChan
}

func (r *RedisStorage) WatchErr() <-chan error {
	return r.watchErrCh
}

func (r *RedisStorage) Watched() (bool, error) {
	if r.watchKey == "" {
		return false, fmt.Errorf("watch key is empty")
	}
	subNums, err := r.client.PubSubNumSub(context.Background(), r.watchKey).Result()
	if err != nil {
		klog.Errorf("notify key=%s watcher error: get pubsub num error: %s", r.watchKey, err.Error())
		return false, err
	}
	if num, ok := subNums[r.watchKey]; !ok || num <= 0 {
		return false, nil
	}
	return true, nil
}

func (r *RedisStorage) Notify(data interface{}) error {
	err := r.notify(data)
	if err != nil && errors.Is(err, ErrNotWatch) {
		return nil
	}
	return err
}

func (r *RedisStorage) notify(data interface{}) error {
	if watched, err := r.Watched(); err != nil {
		return err
	} else if !watched {
		// 没有订阅该key，不发送通知直接返回
		return ErrNotWatch
	}
	msg, err := json.Marshal(data)
	if err != nil {
		klog.Errorf("notify key=%s watcher error: json marshal object error: %s", r.watchKey, err.Error())
		return err
	}
	_, err = r.client.Publish(context.Background(), r.watchKey, msg).Result()
	if err != nil {
		klog.Errorf("notify key=%s watcher error: publish message error: %s", r.watchKey, err.Error())
		return err
	}
	return nil
}

func (r *RedisStorage) NotifyResult(traceId string, timeout int, data interface{}) ([]byte, error) {
	if err := r.notify(data); err != nil {
		return nil, err
	}
	resData, err := r.client.BLPop(context.Background(), time.Duration(timeout)*time.Second, traceId).Result()
	if len(resData) < 2 {
		return nil, ErrTimeout
	}
	if err != nil {
		klog.Errorf("get response error: %s", err.Error())
		return nil, err
	}
	return []byte(resData[1]), nil
}

func (r *RedisStorage) NotifyWatch(traceId string, stopCh <-chan struct{}) <-chan []byte {
	resCh := make(chan []byte)
	cancelCtx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			resData, err := r.client.BLPop(cancelCtx, 0, traceId).Result()
			if errors.Is(err, context.Canceled) {
				klog.V(1).Infof("watch stopped")
				close(resCh)
				return
			}
			if err != nil {
				klog.Errorf("get response error: %s", err.Error())
				continue
			}
			if len(resData) < 2 {
				klog.Errorf("not found data: %v", resData)
			} else {
				resCh <- []byte(resData[1])
			}
		}
	}()
	go func() {
		select {
		case <-stopCh:
			klog.V(1).Infof("stop notify watch")
			cancel()
		}
	}()

	return resCh
}

func (r *RedisStorage) NotifyResponse(traceId string, resp []byte) error {
	pipeLine := r.client.Pipeline()
	pipeLine.LPush(context.Background(), traceId, resp)
	pipeLine.Expire(context.Background(), traceId, time.Second*3)
	if _, err := pipeLine.Exec(context.Background()); err != nil {
		return err
	}
	return nil
}

func (r *RedisStorage) Stop() error {
	if r.stopped {
		return nil
	}
	r.stopped = true
	if r.pubSub != nil {
		err := r.pubSub.Close()
		r.pubSub = nil
		if err != nil && !errors.Is(err, redis.ErrClosed) {
			klog.Errorf("stop pubsub connection error: %s", err.Error())
			return err
		}
	}
	close(r.stopCh)
	return nil
}
