package lock

import "time"

type Lock interface {
	Acquire(key string) (bool, time.Time)
	Release(key string)
}
