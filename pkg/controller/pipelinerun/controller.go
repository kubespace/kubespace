package pipelinerun

import (
	"github.com/kubespace/kubespace/pkg/controller"
	"github.com/kubespace/kubespace/pkg/controller/pipelinerun/job_run"
	"github.com/kubespace/kubespace/pkg/informer"
	pipelinelistwatcher "github.com/kubespace/kubespace/pkg/informer/listwatcher/pipeline"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
)

type PipelineRunController struct {
	models              *model.Models
	pipelineRunInformer informer.Informer
}

func NewPipelineRunController(config *controller.Config) *PipelineRunController {
	jobRun := job_run.NewJobRun(config.Models, config.ServiceFactory.Cluster.KubeClient, config.InformerFactory)
	pipelineRunInformer := config.InformerFactory.PipelineRunInformer(&pipelinelistwatcher.PipelineRunWatchCondition{
		StatusIn: []string{types.PipelineStatusWait, types.PipelineStatusDoing, types.PipelineStatusCancel},
		WithList: true,
	})
	// 添加流水线构建handler
	pipelineRunInformer.AddHandler(NewRunHandler(config.Models, jobRun))
	// 添加流水线取消handler
	pipelineRunInformer.AddHandler(NewCancelHandler(config.Models, jobRun))

	return &PipelineRunController{
		models:              config.Models,
		pipelineRunInformer: pipelineRunInformer,
	}
}

func (p *PipelineRunController) Run(stopCh <-chan struct{}) {
	p.pipelineRunInformer.Run(stopCh)
}
