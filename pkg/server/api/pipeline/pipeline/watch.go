package pipeline

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/informer"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/pipeline"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	pipelineservice "github.com/kubespace/kubespace/pkg/service/pipeline"
	"github.com/kubespace/kubespace/pkg/utils"
	"sync"
)

type watchHandler struct {
	models          *model.Models
	pipelineService *pipelineservice.PipelineService
	informerFactory informer.Factory
}

func WatchHandler(conf *config.ServerConfig) api.Handler {
	return &watchHandler{
		models:          conf.Models,
		pipelineService: conf.ServiceFactory.Pipeline.PipelineService,
		informerFactory: conf.InformerFactory,
	}
}

func (h *watchHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	pipelineId, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	pipelineObj, err := h.models.PipelineManager.GetById(pipelineId)
	if err != nil {
		return true, nil, errors.New(code.DataNotExists, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopePipeline,
		ScopeId: pipelineObj.WorkspaceId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *watchHandler) Handle(c *api.Context) *utils.Response {

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Add("Cache-Control", "no-transform")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	c.SSEvent("message", "{}")
	c.Writer.Flush()

	pipelineId, _ := utils.ParseUint(c.Param("id"))

	pipelineRunInformer := h.informerFactory.PipelineRunInformer(&pipeline.PipelineRunWatchCondition{
		PipelineId: pipelineId,
		WithList:   false,
	})
	stopCh := make(chan struct{})
	mu := sync.Mutex{}
	pipelineRunInformer.AddHandler(&informer.ResourceHandler{HandleFunc: func(obj interface{}) error {
		mu.Lock()
		defer mu.Unlock()
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
