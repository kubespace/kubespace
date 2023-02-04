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

func (m *memoryLock) Acquire(key string) (bool, time.Time) {
	obj, loaded := m.muMap.LoadOrStore(key, time.Now())
	// loaded为true表示已存在，未获取到锁，返回false
	return !loaded, obj.(time.Time)
}

func (m *memoryLock) Release(key string) {
	m.muMap.Delete(key)
}
