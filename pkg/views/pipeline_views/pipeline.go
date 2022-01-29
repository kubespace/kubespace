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

type Pipeline struct {
	Views              []*views.View
	models             *model.Models
	pipelineService    *pipeline.ServicePipeline
	pipelineRunService *pipeline.ServicePipelineRun
}

func NewPipeline(models *model.Models) *Pipeline {
	pw := &Pipeline{
		models:             models,
		pipelineService:    pipeline.NewPipelineService(models),
		pipelineRunService: pipeline.NewPipelineRunService(models),
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "", pw.list),
		views.NewView(http.MethodGet, "/:pipelineId", pw.get),
		views.NewView(http.MethodPost, "", pw.create),
	}
	pw.Views = vs
	return pw
}

func (p *Pipeline) get(c *views.Context) *utils.Response {
	pipelineId, err := strconv.ParseUint(c.Param("pipelineId"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return p.pipelineService.GetPipeline(uint(pipelineId))
}

func (p *Pipeline) list(c *views.Context) *utils.Response {
	var ser serializers.PipelineListSerializer
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return p.pipelineService.ListPipeline(ser.WorkspaceId)
}

func (p *Pipeline) create(c *views.Context) *utils.Response {
	var ser serializers.PipelineCreateSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{
			Code: code.ParamsError,
			Msg:  err.Error(),
		}
	}
	return p.pipelineService.Create(&ser, c.User)
}
