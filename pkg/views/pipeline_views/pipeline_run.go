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
	"time"
)

type PipelineRun struct {
	Views              []*views.View
	models             *model.Models
	pipelineService    *pipeline.ServicePipeline
	pipelineRunService *pipeline.ServicePipelineRun
}

func NewPipelineRun(models *model.Models, pipelineRunService *pipeline.ServicePipelineRun) *PipelineRun {
	pw := &PipelineRun{
		models:             models,
		pipelineService:    pipeline.NewPipelineService(models),
		pipelineRunService: pipelineRunService,
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "list", pw.list),
		views.NewView(http.MethodGet, "/:pipelineRunId", pw.get),
		views.NewView(http.MethodGet, "/:pipelineRunId/sse", pw.sse),
		views.NewView(http.MethodPost, "", pw.build),
		views.NewView(http.MethodPost, "/manual_execute", pw.manual),
		views.NewView(http.MethodPost, "/retry", pw.retry),
		views.NewView(http.MethodGet, "/log/:jobRunId", pw.log),
		views.NewView(http.MethodGet, "/log/:jobRunId/sse", pw.logStream),
	}
	pw.Views = vs
	return pw
}

func (p *PipelineRun) build(c *views.Context) *utils.Response {
	var ser serializers.PipelineBuildSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return p.pipelineRunService.Build(&ser, c.User)
}

func (p *PipelineRun) list(c *views.Context) *utils.Response {
	var ser serializers.PipelineBuildListSerializer
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return p.pipelineRunService.ListPipelineRun(ser.PipelineId, ser.LastBuildNumber)
}

func (p *PipelineRun) get(c *views.Context) *utils.Response {
	pipelineRunId, err := strconv.ParseUint(c.Param("pipelineRunId"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return p.pipelineRunService.GetPipelineRun(uint(pipelineRunId))
}

func (p *PipelineRun) sse(c *views.Context) *utils.Response {
	if c.Param("pipelineRunId") == "" {
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
			sse.EventTypePipelineRun: c.Param("pipelineRunId"),
		},
		ClientChan: make(chan sse.Event),
	}
	sse.Stream.AddClient(streamClient)
	defer sse.Stream.RemoveClient(streamClient)
	c.SSEvent("message", "{}")
	c.Writer.Flush()

	for {
		klog.Infof("select for channel")
		select {
		case <-c.Writer.CloseNotify():
			klog.Info("client gone")
			return nil
		case event := <-streamClient.ClientChan:
			c.SSEvent("message", event)
			c.Writer.Flush()
		}
	}
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

func (p *PipelineRun) log(c *views.Context) *utils.Response {
	jobRunId, err := strconv.ParseUint(c.Param("jobRunId"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	jobLog, err := p.models.ManagerPipelineRun.GetJobRunLog(uint(jobRunId), true)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if jobLog == nil {
		return &utils.Response{Code: code.Success, Data: ""}
	}
	return &utils.Response{Code: code.Success, Data: jobLog.Logs}
}

func (p *PipelineRun) logStream(c *views.Context) *utils.Response {
	if c.Param("jobRunId") == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "get param job run id error"}
	}
	jobRunId, err := strconv.ParseUint(c.Param("jobRunId"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	lastJobLog, err := p.models.ManagerPipelineRun.GetJobRunLog(uint(jobRunId), true)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	c.SSEvent("message", "\n")
	w := c.Writer
	w.Flush()
	clientGone := w.CloseNotify()
	if lastJobLog != nil {
		c.SSEvent("message", lastJobLog.Logs)
	}
	w.Flush()
	//c.Stream()
	tick := time.NewTicker(5 * time.Second)
	for {
		klog.Infof("select for log channel")
		select {
		case <-clientGone:
			klog.Info("log client gone")
			return nil
		case <-tick.C:
			//c.SSEvent("message", event)
			currJobLog, err := p.models.ManagerPipelineRun.GetJobRunLog(uint(jobRunId), false)
			if err == nil {
				if currJobLog != nil && (lastJobLog == nil || currJobLog.UpdateTime != lastJobLog.UpdateTime) {
					lastJobLog, err = p.models.ManagerPipelineRun.GetJobRunLog(uint(jobRunId), true)
					if err != nil {
						klog.Errorf("get job id=%s log error: %s", jobRunId, err.Error())
					} else {
						c.SSEvent("message", lastJobLog.Logs)
						w.Flush()
					}
				}
			} else {
				klog.Errorf("get job id=%s log error: %s", jobRunId, err.Error())
			}
		}
	}
}
