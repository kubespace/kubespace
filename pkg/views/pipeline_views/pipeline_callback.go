package pipeline_views

import (
	"github.com/gin-gonic/gin"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/pipeline"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"net/http"
)

type PipelineCallback struct {
	models             *model.Models
	pipelineService    *pipeline.ServicePipeline
	pipelineRunService *pipeline.ServicePipelineRun
}

func NewPipelineCallback(models *model.Models) *PipelineCallback {
	pc := &PipelineCallback{
		models:             models,
		pipelineService:    pipeline.NewPipelineService(models),
		pipelineRunService: pipeline.NewPipelineRunService(models),
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
