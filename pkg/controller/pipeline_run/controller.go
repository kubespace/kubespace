package pipeline_run

import (
	"github.com/kubespace/kubespace/pkg/controller"
	"github.com/kubespace/kubespace/pkg/controller/pipeline_run/job_run"
	"github.com/kubespace/kubespace/pkg/core/lock"
	"github.com/kubespace/kubespace/pkg/informer"
	pipelinelistwatcher "github.com/kubespace/kubespace/pkg/informer/listwatcher/pipeline"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
)

// PipelineRunController 流水线构建controller
type PipelineRunController struct {
	models              *model.Models
	pipelineRunInformer informer.Informer
	// 流水线构建时对其进行加锁，保证只有一个进行处理
	lock lock.Lock
	// 任务执行处理
	jobRun *job_run.JobRun
}

func NewPipelineRunController(config *controller.Config) *PipelineRunController {

	jobRun := job_run.NewJobRun(config.Models, config.ServiceFactory.Cluster.KubeClient, config.InformerFactory)

	// 监听未构建完成以及要取消的pipelineRun
	pipelineRunInformer := config.InformerFactory.PipelineRunInformer(&pipelinelistwatcher.PipelineRunWatchCondition{
		StatusIn: []string{types.PipelineStatusWait, types.PipelineStatusDoing, types.PipelineStatusCancel},
		WithList: true,
	})

	c := &PipelineRunController{
		models:              config.Models,
		pipelineRunInformer: pipelineRunInformer,
		lock:                lock.NewMemLock(),
		jobRun:              jobRun,
	}
	// 流水线构建handler
	pipelineRunInformer.AddHandler(&informer.ResourceHandler{
		CheckFunc:  c.buildCheck,
		HandleFunc: c.build,
	})
	// 流水线取消handler
	pipelineRunInformer.AddHandler(&informer.ResourceHandler{
		CheckFunc:  c.cancelCheck,
		HandleFunc: c.cancel,
	})

	return c
}

func (p *PipelineRunController) Run(stopCh <-chan struct{}) {
	go p.pipelineRunInformer.Run(stopCh)
}
