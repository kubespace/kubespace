package listwatcher

type Interface interface {
	Key() string
	Run()
	Stop() error
	Result() <-chan interface{}
	WatchErr() <-chan error
	Notify(interface{}) error
}
