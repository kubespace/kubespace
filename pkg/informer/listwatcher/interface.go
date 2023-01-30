package listwatcher

type Interface interface {
	Run()
	Stop() error
	Result() <-chan interface{}
	WatchErr() <-chan error
	Notify(interface{}) error
}
