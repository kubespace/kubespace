package informer

import (
	"github.com/kubespace/kubespace/pkg/informer/listwatcher"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
	"sync"
	"time"
)

type Handler interface {
	Check(interface{}) bool
	Handle(interface{}) error
}

type CommonHandler struct {
	HandleFunc func(interface{}) error
}

func (h *CommonHandler) Check(obj interface{}) bool {
	return true
}

func (h *CommonHandler) Handle(obj interface{}) error {
	return h.HandleFunc(obj)
}

type Informer interface {
	Run(stopCh <-chan struct{})
	AddHandler(Handler)
}

type informer struct {
	listWatcher listwatcher.Interface
	handlers    []Handler
	mu          sync.Mutex
}

func NewInformer(listWatcher listwatcher.Interface) Informer {
	return &informer{
		listWatcher: listWatcher,
		mu:          sync.Mutex{},
	}
}

func (b *informer) Run(stopCh <-chan struct{}) {
	for {
		klog.Infof("start run informer %v", b)
		if err := b.run(stopCh); err != nil {
			klog.Errorf("run informer %v error: %s", b, err.Error())
			// 5s后重试
			tick := time.NewTicker(5 * time.Second)
			select {
			case <-tick.C:
				continue
			case <-stopCh:
				break
			}
		}
		klog.Infof("stop run informer=%v", b)
		break
	}
}

func (b *informer) run(stopCh <-chan struct{}) error {
	b.listWatcher.Run()
	defer b.listWatcher.Stop()
	for {
		select {
		case obj := <-b.listWatcher.Result():
			go b.handle(obj)
		case watchErr := <-b.listWatcher.WatchErr():
			klog.Errorf("informer watch error: %v", watchErr)
		case <-stopCh:
			return nil
		}
	}
}

func (b *informer) AddHandler(handler Handler) {
	b.mu.Lock()
	b.handlers = append(b.handlers, handler)
	b.mu.Unlock()
}

func (b *informer) handle(obj interface{}) {
	defer utils.HandleCrash(func(r interface{}) { klog.Errorf("crashed object: %v", obj) })
	for _, handler := range b.handlers {
		if !handler.Check(obj) {
			continue
		}
		if err := handler.Handle(obj); err != nil {
			klog.Errorf("handle object error=%s, object=%v", err.Error(), obj)
		}
	}
}
