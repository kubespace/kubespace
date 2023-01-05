package watch_storage

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"k8s.io/klog"
)

type RedisStorage struct {
	client     *redis.Client
	watchKey   string
	pubSub     *redis.PubSub
	watchCh    chan []byte
	watchErrCh chan error
}

func NewRedisStorage(redisClient *redis.Client, watchKey string) Storage {
	return &RedisStorage{
		client:     redisClient,
		watchKey:   watchKey,
		watchCh:    make(chan []byte),
		watchErrCh: make(chan error),
	}
}

func (r *RedisStorage) watch() {
	if r.pubSub != nil {
		return
	}
	defer r.Stop()
	r.pubSub = r.client.Subscribe(context.Background(), r.watchKey)
	klog.V(1).Infof("start watch key=%s", r.watchKey)
	for {
		data, err := r.pubSub.ReceiveMessage(context.Background())
		if err != nil {
			klog.Errorf("receive message error: %s", err.Error())
			if errors.Is(err, redis.ErrClosed) {
				return
			}
			r.watchErrCh <- err
			return
		}
		r.watchCh <- []byte(data.Payload)
	}
}

func (r *RedisStorage) Watch() <-chan []byte {
	go r.watch()
	return r.watchCh
}

func (r *RedisStorage) WatchError() <-chan error {
	return r.watchErrCh
}

func (r *RedisStorage) Notify(obj interface{}) error {
	subNums, err := r.client.PubSubNumSub(context.Background(), r.watchKey).Result()
	if err != nil {
		klog.Errorf("notify key=%s watcher error: get pubsub num error: %s", r.watchKey, err.Error())
		return err
	}
	if num, ok := subNums[r.watchKey]; !ok || num <= 0 {
		return nil
	}
	msg, err := json.Marshal(obj)
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

func (r *RedisStorage) Stop() error {
	if r.pubSub != nil {
		err := r.pubSub.Close()
		r.pubSub = nil
		if err != nil && errors.Is(err, redis.ErrClosed) {
			return nil
		}
		return err
	}
	return nil
}
