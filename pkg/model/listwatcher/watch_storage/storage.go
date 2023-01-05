package watch_storage

type Storage interface {
	Stop() error
	Watch() <-chan []byte
	Notify(interface{}) error
}
