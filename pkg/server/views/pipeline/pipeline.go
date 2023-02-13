package pipeline

import (
	"github.com/kubespace/kubespace/pkg/informer"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/pipeline"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	pipelineservice "github.com/kubespace/kubespace/pkg/service/pipeline"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"net/http"
	"strconv"
)

type Pipeline struct {
	Views              []*views.View
	models             *model.Models
	pipelineService    *pipelineservice.ServicePipeline
	pipelineRunService *pipelineservice.ServicePipelineRun
	informerFactory    informer.Factory
}

func NewPipeline(config *config.ServerConfig) *Pipeline {
	pw := &Pipeline{
		models:             config.Models,
		pipelineService:    config.ServiceFactory.Pipeline.PipelineService,
		pipelineRunService: config.ServiceFactory.Pipeline.PipelineRunService,
		informerFactory:    config.InformerFactory,
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "", pw.list),
		views.NewView(http.MethodGet, "/:pipelineId", pw.get),
		views.NewView(http.MethodGet, "/:pipelineId/sse", pw.watch),
		views.NewView(http.MethodPost, "", pw.create),
		views.NewView(http.MethodPut, "", pw.update),
		views.NewView(http.MethodDelete, "/:pipelineId", pw.delete),
		views.NewView(http.MethodGet, "/:pipelineId/list_repo_branch", pw.listRepoBranch),
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

func (p *Pipeline) watch(c *views.Context) *utils.Response {
	if c.Param("pipelineId") == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "get param pipeline run id error"}
	}

	pipelineId, err := strconv.Atoi(c.Param("pipelineId"))
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: "get param pipeline run id error: " + err.Error()}
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	//c.Writer.Header().Set("Transfer-Encoding", "chunked")

	c.SSEvent("message", "{}")
	c.Writer.Flush()

	pipelineRunInformer := p.informerFactory.PipelineRunInformer(&pipeline.PipelineRunWatchCondition{
		PipelineId: uint(pipelineId),
		WithList:   false,
	})
	stopCh := make(chan struct{})
	pipelineRunInformer.AddHandler(&informer.CommonHandler{HandleFunc: func(obj interface{}) error {
		pipelineRun, ok := obj.(types.PipelineRun)
		if !ok {
			return nil
		}
		stageRuns, _ := p.models.PipelineRunManager.StagesRun(pipelineRun.ID)
		c.SSEvent("message", map[string]interface{}{
			"pipeline_run": pipelineRun,
			"stages_run":   stageRuns,
		})
		c.Writer.Flush()
		return nil
	}})
	go pipelineRunInformer.Run(stopCh)

	<-c.Writer.CloseNotify()
	close(stopCh)
	return nil
}

func (p *Pipeline) list(c *views.Context) *utils.Response {
	var ser serializers.PipelineListSerializer
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return p.pipelineService.ListPipeline(ser.WorkspaceId)
}

func (p *Pipeline) create(c *views.Context) *utils.Response {
	var ser serializers.PipelineSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{
			Code: code.ParamsError,
			Msg:  err.Error(),
		}
	}
	return p.pipelineService.Create(&ser, c.User)
}

func (p *Pipeline) update(c *views.Context) *utils.Response {
	var ser serializers.PipelineSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{
			Code: code.ParamsError,
			Msg:  err.Error(),
		}
	}
	return p.pipelineService.Update(&ser, c.User)
}

func (p *Pipeline) delete(c *views.Context) *utils.Response {
	resp := &utils.Response{Code: code.Success}
	id, err := strconv.ParseUint(c.Param("pipelineId"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	err = p.models.PipelineManager.Delete(uint(id))
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: "删除流水线失败：" + err.Error()}
	}
	return resp
}

func (p *Pipeline) listRepoBranch(c *views.Context) *utils.Response {
	if c.Param("pipelineId") == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "get param pipeline run id error"}
	}

	pipelineId, err := strconv.Atoi(c.Param("pipelineId"))
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: "get param pipeline id error: " + err.Error()}
	}

	return p.pipelineService.ListRepoBranches(uint(pipelineId))
}
