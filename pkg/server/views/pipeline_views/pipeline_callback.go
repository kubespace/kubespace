package pipeline_views

import (
	"github.com/gin-gonic/gin"
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	pipeline2 "github.com/kubespace/kubespace/pkg/service/pipeline"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"net/http"
)

type PipelineCallback struct {
	models             *model.Models
	pipelineService    *pipeline2.ServicePipeline
	pipelineRunService *pipeline2.ServicePipelineRun
}

func NewPipelineCallback(models *model.Models, kr *kube_resource.KubeResources) *PipelineCallback {
	pc := &PipelineCallback{
		models:             models,
		pipelineService:    pipeline2.NewPipelineService(models),
		pipelineRunService: pipeline2.NewPipelineRunService(models, kr),
	}
	return pc
}

func (p *PipelineCallback) Callback(c *gin.Context) {
	var ser serializers.PipelineCallbackSerializer
	if err := c.ShouldBind(&ser); err != nil {
		c.JSON(http.StatusOK, &utils.Response{Code: code.ParamsError, Msg: err.Error()})
		return
	}
	resp := p.pipelineRunService.Callback(ser)
	c.JSON(http.StatusOK, resp)
}
