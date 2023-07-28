package spacelet

import (
	"github.com/kubespace/kubespace/pkg/controller"
	"github.com/kubespace/kubespace/pkg/core/lock"
	"github.com/kubespace/kubespace/pkg/informer"
	spaceletlistwatcher "github.com/kubespace/kubespace/pkg/informer/listwatcher/spacelet"
	"github.com/kubespace/kubespace/pkg/model"
)

type SpaceletController struct {
	models           *model.Models
	spaceletInformer informer.Informer
	// 流水线构建时对其进行加锁，保证只有一个进行处理
	lock lock.Lock
}

func NewSpaceletController(config *controller.Config) *SpaceletController {
	// 定时监听所有的Spacelet节点
	spaceletInformer := config.InformerFactory.SpaceletInformer(&spaceletlistwatcher.SpaceletWatchCondition{})

	c := &SpaceletController{
		models:           config.Models,
		spaceletInformer: spaceletInformer,
		lock:             lock.NewMemLock(),
	}

	// 定时探测spacelet节点存活
	spaceletInformer.AddHandler(&informer.ResourceHandler{
		CheckFunc:  c.probeCheck,
		HandleFunc: c.probe,
	})

	return c
}

func (s *SpaceletController) Run(stopCh <-chan struct{}) {
	go s.spaceletInformer.Run(stopCh)
}
