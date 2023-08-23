package pipeline

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/informer"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/pipeline"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	pipelineservice "github.com/kubespace/kubespace/pkg/service/pipeline"
	"github.com/kubespace/kubespace/pkg/service/pipeline/pipeline_run"
	"github.com/kubespace/kubespace/pkg/service/pipeline/schemas"
	"github.com/kubespace/kubespace/pkg/utils"
	"net/http"
	"strconv"
	"sync"
)

type Pipeline struct {
	Views              []*views.View
	models             *model.Models
	pipelineService    *pipelineservice.PipelineService
	pipelineRunService *pipeline_run.PipelineRunService
	informerFactory    informer.Factory
}

func NewPipeline(config *config.ServerConfig) *Pipeline {
	pw := &Pipeline{
		models:             config.Models,
		pipelineService:    config.ServiceFactory.Pipeline.PipelineService,
		pipelineRunService: config.ServiceFactory.Pipeline.PipelineRunService,
		informerFactory:    config.InformerFactory,
	}
	pw.Views = []*views.View{
		views.NewView(http.MethodGet, "", pw.list),
		views.NewView(http.MethodGet, "/:pipelineId", pw.get),
		views.NewView(http.MethodGet, "/:pipelineId/sse", pw.watch),
		views.NewView(http.MethodPost, "", pw.create),
		views.NewView(http.MethodPut, "", pw.update),
		views.NewView(http.MethodDelete, "/:pipelineId", pw.delete),
		views.NewView(http.MethodGet, "/:pipelineId/list_repo_branch", pw.listRepoBranch),
	}
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
	c.Writer.Header().Add("Cache-Control", "no-transform")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	c.SSEvent("message", "{}")
	c.Writer.Flush()

	pipelineRunInformer := p.informerFactory.PipelineRunInformer(&pipeline.PipelineRunWatchCondition{
		PipelineId: uint(pipelineId),
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
	var ser schemas.PipelineParams
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	workspace, err := p.models.PipelineWorkspaceManager.Get(ser.WorkspaceId)
	if err != nil {
		return c.GenerateResponse(errors.New(code.DataNotExists, "获取流水线空间失败："+err.Error()), nil)
	}
	pipelineObj, err := p.pipelineService.Create(&ser, c.User)
	var resId uint
	if pipelineObj != nil {
		resId = pipelineObj.ID
	}
	resp := c.GenerateResponse(err, pipelineObj)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationCreate,
		OperateDetail:        "创建流水线:" + ser.Name,
		Scope:                types.ScopePipeline,
		ScopeId:              workspace.ID,
		ScopeName:            workspace.Name,
		ResourceId:           resId,
		ResourceType:         types.AuditResourcePipeline,
		ResourceName:         ser.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: ser,
	})
	return resp
}

func (p *Pipeline) update(c *views.Context) *utils.Response {
	var ser schemas.PipelineParams
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	workspace, err := p.models.PipelineWorkspaceManager.Get(ser.WorkspaceId)
	if err != nil {
		return c.GenerateResponse(errors.New(code.DataNotExists, err), nil)
	}
	pipelineObj, err := p.pipelineService.Update(&ser, c.User)
	resp := c.GenerateResponse(err, nil)
	if pipelineObj == nil {
		return resp
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationUpdate,
		OperateDetail:        "更新流水线:" + ser.Name,
		Scope:                types.ScopePipeline,
		ScopeId:              workspace.ID,
		ScopeName:            workspace.Name,
		ResourceId:           pipelineObj.ID,
		ResourceType:         types.AuditResourcePipeline,
		ResourceName:         ser.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: ser,
	})
	return resp
}

func (p *Pipeline) delete(c *views.Context) *utils.Response {
	id, err := strconv.ParseUint(c.Param("pipelineId"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	pipelineObj, err := p.models.PipelineManager.Get(uint(id))
	if err != nil {
		return c.GenerateResponse(errors.New(code.DataNotExists, fmt.Sprintf("获取流水线id=%d失败：%s", id, err.Error())), nil)
	}
	workspace, err := p.models.PipelineWorkspaceManager.Get(pipelineObj.WorkspaceId)
	if err != nil {
		return c.GenerateResponse(errors.New(code.DataNotExists, fmt.Sprintf("获取流水线空间id=%d失败：%s", pipelineObj.WorkspaceId, err.Error())), nil)
	}
	err = p.models.PipelineManager.Delete(uint(id))
	if err != nil {
		err = errors.New(code.DeleteError, "删除流水线失败："+err.Error())
	}
	resp := c.GenerateResponse(err, nil)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationDelete,
		OperateDetail:        "删除流水线:" + pipelineObj.Name,
		Scope:                types.ScopePipeline,
		ScopeId:              workspace.ID,
		ScopeName:            workspace.Name,
		ResourceId:           pipelineObj.ID,
		ResourceType:         types.AuditResourcePipeline,
		ResourceName:         pipelineObj.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: nil,
	})
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
