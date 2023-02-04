package pipeline

import (
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/server/config"
	pipelineservice "github.com/kubespace/kubespace/pkg/service/pipeline"
)

type PipelineCallback struct {
	models             *model.Models
	pipelineRunService *pipelineservice.ServicePipelineRun
}

func NewPipelineCallback(config *config.ServerConfig) *PipelineCallback {
	pc := &PipelineCallback{
		models:             config.Models,
		pipelineRunService: config.ServiceFactory.Pipeline.PipelineRunService,
	}
	return pc
}

//func (p *PipelineCallback) Callback(c *gin.Context) {
//	var ser serializers.PipelineCallbackSerializer
//	if err := c.ShouldBind(&ser); err != nil {
//		c.JSON(http.StatusOK, &utils.Response{Code: code.ParamsError, Msg: err.Error()})
//		return
//	}
//	resp := p.pipelineRunService.Callback(ser)
//	c.JSON(http.StatusOK, resp)
//}
