package lock

import "time"

type Lock interface {
	Locked(key string) (bool, *time.Time)
	Acquire(key string) (bool, time.Time)
	Release(key string)
}
