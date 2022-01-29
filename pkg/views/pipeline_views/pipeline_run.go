package pipeline_views

import (
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/pipeline"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"net/http"
	"strconv"
)

type PipelineRun struct {
	Views              []*views.View
	models             *model.Models
	pipelineService    *pipeline.ServicePipeline
	pipelineRunService *pipeline.ServicePipelineRun
}

func NewPipelineRun(models *model.Models) *PipelineRun {
	pw := &PipelineRun{
		models:             models,
		pipelineService:    pipeline.NewPipelineService(models),
		pipelineRunService: pipeline.NewPipelineRunService(models),
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "list", pw.list),
		views.NewView(http.MethodGet, "/:pipelineRunId", pw.get),
		views.NewView(http.MethodPost, "", pw.build),
		views.NewView(http.MethodPost, "/manual_execute", pw.manual),
		views.NewView(http.MethodPost, "/retry", pw.retry),
	}
	pw.Views = vs
	return pw
}

func (p *PipelineRun) build(c *views.Context) *utils.Response {
	var ser serializers.PipelineBuildSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg:  err.Error()}
	}
	return p.pipelineRunService.Build(&ser, c.User)
}

func (p *PipelineRun) list(c *views.Context) *utils.Response {
	var ser serializers.PipelineBuildListSerializer
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg:  err.Error(),}
	}
	return p.pipelineRunService.ListPipelineRun(ser.PipelineId)
}

func (p *PipelineRun) get(c *views.Context) *utils.Response {
	pipelineRunId, err := strconv.ParseUint(c.Param("pipelineRunId"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return p.pipelineRunService.GetPipelineRun(uint(pipelineRunId))
}

func (p *PipelineRun) manual(c *views.Context) *utils.Response {
	var ser serializers.PipelineStageManualSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{
			Code: code.ParamsError,
			Msg:  err.Error(),
		}
	}
	return p.pipelineRunService.ManualExecuteStage(&ser)
}

func (p *PipelineRun) retry(c *views.Context) *utils.Response {
	var ser serializers.PipelineStageRetrySerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{
			Code: code.ParamsError,
			Msg:  err.Error(),
		}
	}
	return p.pipelineRunService.RetryStage(&ser)
}
