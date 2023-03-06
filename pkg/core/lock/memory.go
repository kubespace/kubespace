package lock

import (
	"sync"
	"time"
)

type memoryLock struct {
	muMap sync.Map
}

func NewMemLock() Lock {
	return &memoryLock{muMap: sync.Map{}}
}

// Locked 是否已存在锁
func (m *memoryLock) Locked(key string) (bool, *time.Time) {
	val, ok := m.muMap.Load(key)
	if ok {
		// ok为true表示已存在
		t := val.(time.Time)
		return ok, &t
	}
	return ok, nil
}

// Acquire 争抢锁
func (m *memoryLock) Acquire(key string) (bool, time.Time) {
	obj, loaded := m.muMap.LoadOrStore(key, time.Now())
	// loaded为true表示已存在，未获取到锁，返回false
	return !loaded, obj.(time.Time)
}

// Release 释放锁
func (m *memoryLock) Release(key string) {
	m.muMap.Delete(key)
}
