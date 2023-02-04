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
	"k8s.io/klog/v2"
	"net/http"
	"strconv"
	"time"
)

type PipelineRun struct {
	Views              []*views.View
	models             *model.Models
	pipelineService    *pipelineservice.ServicePipeline
	pipelineRunService *pipelineservice.ServicePipelineRun
	informerFactory    informer.Factory
}

func NewPipelineRun(config *config.ServerConfig) *PipelineRun {
	pw := &PipelineRun{
		models:             config.Models,
		pipelineService:    config.ServiceFactory.Pipeline.PipelineService,
		pipelineRunService: config.ServiceFactory.Pipeline.PipelineRunService,
		informerFactory:    config.InformerFactory,
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "list", pw.list),
		views.NewView(http.MethodGet, "/:pipelineRunId", pw.get),
		views.NewView(http.MethodGet, "/:pipelineRunId/sse", pw.watch),
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
	if ser.Limit == 0 {
		ser.Limit = 20
	}
	return p.pipelineRunService.ListPipelineRun(ser.PipelineId, ser.LastBuildNumber, ser.Status, ser.Limit)
}

func (p *PipelineRun) get(c *views.Context) *utils.Response {
	pipelineRunId, err := strconv.ParseUint(c.Param("pipelineRunId"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return p.pipelineRunService.GetPipelineRun(uint(pipelineRunId))
}

func (p *PipelineRun) watch(c *views.Context) *utils.Response {
	if c.Param("pipelineRunId") == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "get param pipeline run id error"}
	}

	pipelineRunId, err := strconv.Atoi(c.Param("pipelineRunId"))
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
		Id: uint(pipelineRunId),
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
	jobLog, err := p.models.PipelineRunManager.GetJobRunLog(uint(jobRunId), true)
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
	lastJobLog, err := p.models.PipelineRunManager.GetJobRunLog(uint(jobRunId), true)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	c.SSEvent("message", "{}")
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
		select {
		case <-clientGone:
			klog.Info("log client gone")
			return nil
		case <-tick.C:
			//c.SSEvent("message", event)
			currJobLog, err := p.models.PipelineRunManager.GetJobRunLog(uint(jobRunId), false)
			if err == nil {
				if currJobLog != nil && (lastJobLog == nil || currJobLog.UpdateTime != lastJobLog.UpdateTime) {
					lastJobLog, err = p.models.PipelineRunManager.GetJobRunLog(uint(jobRunId), true)
					if err != nil {
						klog.Errorf("get job id=%s log error: %s", jobRunId, err.Error())
					} else {
						c.SSEvent("message", lastJobLog.Logs)
						w.Flush()
					}
				}
			} else {
				klog.Errorf("get job id=%s log error: %s", jobRunId, err.Error())
				c.SSEvent("message", "{}")
				w.Flush()
			}
		}
	}
}
