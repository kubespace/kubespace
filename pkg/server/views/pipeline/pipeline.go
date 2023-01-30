package pipeline

import (
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	pipelineservice "github.com/kubespace/kubespace/pkg/service/pipeline"
	"github.com/kubespace/kubespace/pkg/sse"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"k8s.io/klog/v2"
	"net/http"
	"strconv"
	"time"
)

type Pipeline struct {
	Views              []*views.View
	models             *model.Models
	pipelineService    *pipelineservice.ServicePipeline
	pipelineRunService *pipelineservice.ServicePipelineRun
}

func NewPipeline(config *config.ServerConfig) *Pipeline {
	pw := &Pipeline{
		models:             config.Models,
		pipelineService:    config.ServiceFactory.Pipeline.PipelineService,
		pipelineRunService: config.ServiceFactory.Pipeline.PipelineRunService,
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
	c.SSEvent("message", "{}")
	c.Writer.Flush()

	tick := time.NewTicker(30 * time.Second)

	for {
		klog.Infof("select for channel")
		select {
		case <-c.Writer.CloseNotify():
			klog.Info("client gone")
			return nil
		case event := <-streamClient.ClientChan:
			c.SSEvent("message", event)
			c.Writer.Flush()
		case <-tick.C:
			c.SSEvent("message", "{}")
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
