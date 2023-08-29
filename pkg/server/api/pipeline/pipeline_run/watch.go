package pipeline_run

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/informer"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/pipeline"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
)

type watchHandler struct {
	models          *model.Models
	informerFactory informer.Factory
}

func WatchHandler(conf *config.ServerConfig) api.Handler {
	return &watchHandler{
		models:          conf.Models,
		informerFactory: conf.InformerFactory,
	}
}

func (h *watchHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	workspaceId, _ := utils.ParseUint(c.Query("workspace_id"))
	return true, &api.AuthPerm{
		Scope:   types.ScopePipeline,
		ScopeId: workspaceId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *watchHandler) Handle(c *api.Context) *utils.Response {
	pipelineRunId, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, "get param pipeline run id error: "+err.Error()))
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	c.SSEvent("message", "{}")
	c.Writer.Flush()

	pipelineRunInformer := h.informerFactory.PipelineRunInformer(&pipeline.PipelineRunWatchCondition{
		Id: pipelineRunId,
	})
	stopCh := make(chan struct{})
	pipelineRunInformer.AddHandler(&informer.ResourceHandler{HandleFunc: func(obj interface{}) error {
		pipelineRun, ok := obj.(types.PipelineRun)
		if !ok {
			return nil
		}
		stageRuns, _ := h.models.PipelineRunManager.StagesRun(pipelineRun.ID)
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
