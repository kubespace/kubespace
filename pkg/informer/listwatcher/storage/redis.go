package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/kubespace/kubespace/pkg/core/datatype"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
	"sync"
	"time"
)

var ErrNotWatch = errors.New("watch key is not subscribed")
var ErrTimeout = errors.New("result timeout")

type RedisStorage struct {
	id         string
	client     *redis.Client
	watchKey   string
	pubSub     *redis.PubSub
	resultChan chan interface{}
	watchErrCh chan error
	resyncSec  int
	listFunc   ListFunc
	filterFunc FilterFunc
	stopCh     chan struct{}
	stopped    bool
	dataType   datatype.DataType
}

func NewRedisStorage(redisClient *redis.Client, watchKey string, listFunc ListFunc, filterFunc FilterFunc, resyncSec int, dataType datatype.DataType) Storage {
	return &RedisStorage{
		id:         utils.CreateUUID(),
		client:     redisClient,
		watchKey:   watchKey,
		resultChan: make(chan interface{}),
		watchErrCh: make(chan error),
		listFunc:   listFunc,
		filterFunc: filterFunc,
		resyncSec:  resyncSec,
		stopCh:     make(chan struct{}),
		dataType:   dataType,
	}
}

func (r *RedisStorage) Id() string {
	return r.id
}

func (r *RedisStorage) Delegate(obj interface{}) {
	if r.filterFunc == nil {
		r.resultChan <- obj
	} else if r.filterFunc(obj) {
		r.resultChan <- obj
	}
}

func (r *RedisStorage) DelegateErr(err error) {
	r.watchErrCh <- err
}

func (r *RedisStorage) Run() {
	if r.stopped {
		r.stopped = false
		r.stopCh = make(chan struct{})
	}
	if r.watchKey != "" {
		// 监听watchKey
		sharedWatch.Watch(r.client, r.watchKey, r, r.dataType)
	}
	if r.resyncSec > 0 && r.listFunc != nil {
		go r.list()
	}
}

// 定时查询listFunc定义的方法
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
	if r.watchKey != "" {
		sharedWatch.Stop(r.watchKey, r)
	}
	close(r.stopCh)
	return nil
}

var sharedWatch = newSharedWatch()

// SharedWatch 在一个实例中共享同一个watchKey的监听
type SharedWatch struct {
	// key是watchKey，value是redis client对watchKey的监听对象
	sharedKeyWatchMap map[string]*sharedWatchStorage
	// 对sharedKeyWatchMap进行操作时加锁，保证原子性
	mu *sync.Mutex
}

func newSharedWatch() *SharedWatch {
	return &SharedWatch{
		sharedKeyWatchMap: make(map[string]*sharedWatchStorage),
		mu:                &sync.Mutex{},
	}
}

// Watch 对watchKey进行监听，如果已监听，则直接添加委托
func (s *SharedWatch) Watch(client *redis.Client, watchKey string, delegate sharedWatchDelegate, dataType datatype.DataType) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ws, ok := s.sharedKeyWatchMap[watchKey]
	if ok {
		// 已经存在监听实例，直接添加委托对象
		ws.AddDelegate(delegate)
		return
	}
	ws = newSharedWatchStorage(client, watchKey, delegate, dataType)
	s.sharedKeyWatchMap[watchKey] = ws
	// 开启一个协程后台监听该watchKey
	go ws.Watch()
}

// Stop 停止对该委托对象的监听
func (s *SharedWatch) Stop(watchKey string, delegate sharedWatchDelegate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ws, ok := s.sharedKeyWatchMap[watchKey]
	if ok {
		ws.RemoveDelegate(delegate)
		if ws.Stopped() {
			// 如果已经停止了，从map中删除
			delete(s.sharedKeyWatchMap, watchKey)
		}
	}
}

// sharedWatchDelegate 对watchKey监听到对象或错误后，委托发送给该接口
type sharedWatchDelegate interface {
	Id() string
	Delegate(obj interface{})
	// DelegateErr 监听失败的错误转发
	DelegateErr(err error)
}

// sharedWatchStorage 监听watchKey的redis存储
type sharedWatchStorage struct {
	client   *redis.Client
	watchKey string
	// 每个listwatcher实例对watchKey的监听注册到这里，当监听到对象时，发送给每个listwatcher
	delegates map[string]sharedWatchDelegate
	stopped   bool
	dataType  datatype.DataType
	mu        *sync.Mutex
	pubsub    *redis.PubSub
	cancel    context.CancelFunc
	ctx       context.Context
}

func newSharedWatchStorage(client *redis.Client, watchKey string, delegate sharedWatchDelegate, dataType datatype.DataType) *sharedWatchStorage {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &sharedWatchStorage{
		client:    client,
		watchKey:  watchKey,
		delegates: map[string]sharedWatchDelegate{delegate.Id(): delegate},
		stopped:   false,
		dataType:  dataType,
		mu:        &sync.Mutex{},
		cancel:    cancelFunc,
		ctx:       ctx,
	}
}

func (s *sharedWatchStorage) AddDelegate(delegate sharedWatchDelegate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegates[delegate.Id()] = delegate
}

func (s *sharedWatchStorage) RemoveDelegate(delegate sharedWatchDelegate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.delegates, delegate.Id())
	if len(s.delegates) == 0 {
		// 如果没有要监听发送的委托对象了，则退出监听
		s.Stop()
	}
}

func (s *sharedWatchStorage) Watch() {
	if s.pubsub == nil {
		s.pubsub = s.client.Subscribe(context.Background(), s.watchKey)
	}
	klog.Infof("start share watch key=%s", s.watchKey)
	for {
		data, err := s.pubsub.ReceiveMessage(s.ctx)
		if s.stopped {
			break
		}
		if err != nil {
			klog.Errorf("receive message error: %s", err.Error())
			for _, delegate := range s.delegates {
				// 发送错误给每个委托对象
				delegate.DelegateErr(err)
			}
			// 5s后重试
			tick := time.NewTicker(5 * time.Second)
			klog.Infof("retry watch key=%s after 5 seconds", s.watchKey)
			select {
			case <-tick.C:
				continue
			case <-s.ctx.Done():
				break
			}
		}
		// 对监听到的数据统一进行转换
		obj, err := s.dataType.Unmarshal([]byte(data.Payload))
		if err != nil {
			klog.Errorf("unmarshal receive message to data type error: %s, datatype=%v, message=%s", err.Error(), s.dataType, data.Payload)
		} else {
			for _, delegate := range s.delegates {
				// 发送数据给每个委托对象
				delegate.Delegate(obj)
			}
		}
	}
	klog.Infof("stopped share watch key=%s", s.watchKey)
}

func (s *sharedWatchStorage) Stopped() bool {
	return s.stopped
}

func (s *sharedWatchStorage) Stop() {
	s.stopped = true
	if s.pubsub != nil {
		s.pubsub.Close()
	}
	s.cancel()
}
