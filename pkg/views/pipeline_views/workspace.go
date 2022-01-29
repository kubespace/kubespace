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

type PipelineWorkspace struct {
	Views            []*views.View
	models           *model.Models
	workspaceService *pipeline.WorkspaceService
}

func NewPipelineWorkspace(models *model.Models) *PipelineWorkspace {
	pipelineWs := &PipelineWorkspace{
		models:           models,
		workspaceService: pipeline.NewWorkspaceService(models),
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "", pipelineWs.list),
		views.NewView(http.MethodPost, "", pipelineWs.create),
		views.NewView(http.MethodDelete, "/:id", pipelineWs.delete),
	}
	pipelineWs.Views = vs
	return pipelineWs
}

func (p *PipelineWorkspace) create(c *views.Context) *utils.Response {
	var ser serializers.WorkspaceSerializer
	resp := &utils.Response{Code: code.Success}
	if err := c.ShouldBind(&ser); err != nil {
		resp.Code = code.ParamsError
		resp.Msg = err.Error()
		return resp
	}
	return p.workspaceService.Create(&ser, c.User)
}

func (p *PipelineWorkspace) list(c *views.Context) *utils.Response {
	resp := &utils.Response{Code: code.Success}
	workspaces, err := p.models.PipelineWorkspaceManager.List()
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = err.Error()
		return resp
	}
	resp.Data = workspaces
	return resp
}

func (p *PipelineWorkspace) delete(c *views.Context) *utils.Response {
	resp := &utils.Response{Code: code.Success}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	workspace, err := p.models.PipelineWorkspaceManager.Get(uint(id))
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "获取流水线空间失败: " + err.Error()
		return resp
	}
	err = p.models.PipelineWorkspaceManager.Delete(workspace)
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "删除流水线空间失败: " + err.Error()
		return resp
	}
	return resp
}
