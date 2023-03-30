package pipeline_trigger

import (
	"github.com/kubespace/kubespace/pkg/controller"
	"github.com/kubespace/kubespace/pkg/core/lock"
	"github.com/kubespace/kubespace/pkg/informer"
	pipelinelistwatcher "github.com/kubespace/kubespace/pkg/informer/listwatcher/pipeline"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/service/pipeline"
)

// PipelineTriggerController 流水线触发事件controller，对所有的触发配置进行监听，如果有触发条件满足，则生成触发事件
// 并监听未消费的触发事件，生成流水线构建
type PipelineTriggerController struct {
	models                       *model.Models
	pipelineTriggerInformer      informer.Informer
	pipelineTriggerEventInformer informer.Informer
	pipelineCodeCacheInformer    informer.Informer
	// 流水线构建时对其进行加锁，保证只有一个进行处理
	lock               lock.Lock
	pipelineRunService *pipeline.ServicePipelineRun
}

func NewPipelineTriggerController(config *controller.Config) *PipelineTriggerController {

	// 定时监听所有的流水线触发配置
	pipelineTriggerInformer := config.InformerFactory.PipelineTriggerInformer(&pipelinelistwatcher.PipelineTriggerWatchCondition{})
	// 监听流水线触发事件状态为New状态
	pipelineTriggerEventInformer := config.InformerFactory.PipelineTriggerEventInformer(
		&pipelinelistwatcher.PipelineTriggerEventWatchCondition{Status: types.PipelineTriggerEventStatusNew})
	// 监听流水线代码缓存状态为open状态
	pipelineCodeCacheInformer := config.InformerFactory.PipelineCodeCacheInformer(
		&pipelinelistwatcher.PipelineCodeCacheWatchCondition{Status: types.PipelineCodeCacheStatusOpen})

	c := &PipelineTriggerController{
		models:                       config.Models,
		pipelineTriggerInformer:      pipelineTriggerInformer,
		pipelineTriggerEventInformer: pipelineTriggerEventInformer,
		pipelineCodeCacheInformer:    pipelineCodeCacheInformer,
		lock:                         lock.NewMemLock(),
		pipelineRunService:           config.ServiceFactory.Pipeline.PipelineRunService,
	}
	// 流水线触发handler
	pipelineTriggerInformer.AddHandler(&informer.ResourceHandler{
		CheckFunc:  c.triggerCheck,
		HandleFunc: c.triggerHandle,
	})
	// 流水线触发事件handler，构建流水线
	pipelineTriggerEventInformer.AddHandler(&informer.ResourceHandler{
		CheckFunc:  c.eventCheck,
		HandleFunc: c.eventHandle,
	})
	// 流水线代码分支缓存handler
	pipelineCodeCacheInformer.AddHandler(&informer.ResourceHandler{
		CheckFunc:  c.codeCacheCheck,
		HandleFunc: c.codeCacheHandle,
	})

	return c
}

func (p *PipelineTriggerController) Run(stopCh <-chan struct{}) {
	go p.pipelineTriggerInformer.Run(stopCh)
	go p.pipelineTriggerEventInformer.Run(stopCh)
	go p.pipelineCodeCacheInformer.Run(stopCh)
}
