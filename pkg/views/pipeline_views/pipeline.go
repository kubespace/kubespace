package pipeline_views

import (
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/pipeline"
	"github.com/kubespace/kubespace/pkg/sse"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"k8s.io/klog"
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
		views.NewView(http.MethodGet, "/:pipelineId/sse", pw.sse),
		views.NewView(http.MethodPost, "", pw.create),
		views.NewView(http.MethodPut, "", pw.update),
		views.NewView(http.MethodDelete, "/:pipelineId", pw.delete),
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

func (p *Pipeline) sse(c *views.Context) *utils.Response {
	if c.Param("pipelineId") == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "get param pipeline run id error"}
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	//c.Writer.Header().Set("Transfer-Encoding", "chunked")

	streamClient := sse.StreamClient{
		ClientId: utils.CreateUUID(),
		Catalog:  sse.CatalogDatabase,
		WatchSelector: map[string]string{
			sse.EventTypePipeline: c.Param("pipelineId"),
		},
		ClientChan: make(chan sse.Event),
	}
	sse.Stream.AddClient(streamClient)
	defer sse.Stream.RemoveClient(streamClient)
	w := c.Writer
	clientGone := w.CloseNotify()
	c.SSEvent("message", "")
	w.Flush()
	//c.Stream()

	for {
		klog.Infof("select for channel")
		select {
		case <-clientGone:
			klog.Info("client gone")
			return nil
		case event := <-streamClient.ClientChan:
			c.SSEvent("message", event)
			c.Writer.Flush()
		}
	}
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
	err = p.models.ManagerPipeline.Delete(uint(id))
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: "删除流水线失败：" + err.Error()}
	}
	return resp
}
