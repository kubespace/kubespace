package storage

type Storage interface {
	Key() string
	Stop() error
	Run()
	Result() <-chan interface{}
	WatchErr() <-chan error
	Notify(interface{}) error
	Watched() (bool, error)
	NotifyResult(traceId string, timeout int, data interface{}) ([]byte, error)
	NotifyWatch(traceId string, stopCh <-chan struct{}) <-chan []byte
	NotifyResponse(traceId string, resp []byte) error
}

type ListFunc func() ([]interface{}, error)
type FilterFunc func(interface{}) bool
