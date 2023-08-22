package pipeline

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/service/pipeline"
	"github.com/kubespace/kubespace/pkg/service/pipeline/schemas"
	"github.com/kubespace/kubespace/pkg/utils"
	"net/http"
	"strconv"
	"time"
)

type PipelineWorkspace struct {
	Views            []*views.View
	models           *model.Models
	workspaceService *pipeline.WorkspaceService
}

func NewPipelineWorkspace(config *config.ServerConfig) *PipelineWorkspace {
	pipelineWs := &PipelineWorkspace{
		models:           config.Models,
		workspaceService: config.ServiceFactory.Pipeline.WorkspaceService,
	}
	pipelineWs.Views = []*views.View{
		views.NewView(http.MethodGet, "", pipelineWs.list),
		views.NewView(http.MethodGet, "/latest_release", pipelineWs.latestReleaseVersion),
		views.NewView(http.MethodGet, "/exists_release", pipelineWs.existsReleaseVersion),
		views.NewView(http.MethodGet, "/:id", pipelineWs.get),
		views.NewView(http.MethodPost, "", pipelineWs.create),
		views.NewView(http.MethodPut, "/:id", pipelineWs.update),
		views.NewView(http.MethodDelete, "/:id", pipelineWs.delete),
		views.NewView(http.MethodGet, "/list_git_repos", pipelineWs.listGitRepos),
	}
	return pipelineWs
}

func (p *PipelineWorkspace) create(c *views.Context) *utils.Response {
	var ser serializers.WorkspaceSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	workspace, err := p.workspaceService.Create(&ser, c.User)
	resp := c.GenerateResponse(err, workspace)
	var scopeId uint
	if workspace != nil {
		scopeId = workspace.ID
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationCreate,
		OperateDetail:        "创建流水线空间:" + ser.Name,
		Scope:                types.ScopePipeline,
		ScopeId:              scopeId,
		ScopeName:            ser.Name,
		ResourceId:           scopeId,
		ResourceType:         types.AuditResourcePipeSpace,
		ResourceName:         ser.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: ser,
	})
	return resp
}

func (p *PipelineWorkspace) update(c *views.Context) *utils.Response {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	workspace, err := p.models.PipelineWorkspaceManager.Get(uint(id))
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: "获取流水线空间失败: " + err.Error()}
	}
	var ser serializers.WorkspaceUpdateSerializer
	if err = c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	if ser.CodeSecretId != 0 && workspace.Code != nil {
		workspace.Code.SecretId = ser.CodeSecretId
	}
	if ser.Description != "" {
		workspace.Description = ser.Description
	}
	workspace.UpdateUser = c.User.Name
	workspace.UpdateTime = time.Now()
	_, err = p.models.PipelineWorkspaceManager.Update(workspace)
	if err != nil {
		err = errors.New(code.DBError, "更新流水线空间失败: "+err.Error())
	}
	resp := c.GenerateResponse(err, nil)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationUpdate,
		OperateDetail:        "更新流水线空间:" + workspace.Name,
		Scope:                types.ScopePipeline,
		ScopeId:              workspace.ID,
		ScopeName:            workspace.Name,
		ResourceId:           workspace.ID,
		ResourceType:         types.AuditResourcePipeSpace,
		ResourceName:         workspace.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: ser,
	})
	return resp
}

func (p *PipelineWorkspace) list(c *views.Context) *utils.Response {
	var ser serializers.WorkspaceListSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	resp := &utils.Response{Code: code.Success}
	workspaces, err := p.models.PipelineWorkspaceManager.List()
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = err.Error()
		return resp
	}
	var data []types.PipelineWorkspace
	for i, w := range workspaces {
		if !p.models.UserRoleManager.HasScopeRole(c.User, types.ScopePipeline, w.ID, types.RoleTypeViewer) {
			continue
		}
		if ser.Type != "" && w.Type != ser.Type {
			continue
		}
		if ser.WithPipeline {
			workspaces[i].Pipelines, err = p.models.PipelineManager.List(w.ID)
			if err != nil {
				return &utils.Response{Code: code.DBError, Msg: err.Error()}
			}
		}
		data = append(data, workspaces[i])
	}
	resp.Data = data
	return resp
}

func (p *PipelineWorkspace) get(c *views.Context) *utils.Response {
	workspaceId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	workspace, err := p.models.PipelineWorkspaceManager.Get(uint(workspaceId))
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: "获取流水线空间失败：" + err.Error()}
	}
	return &utils.Response{Code: code.Success, Data: workspace}
}

func (p *PipelineWorkspace) delete(c *views.Context) *utils.Response {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.GenerateResponse(errors.New(code.ParamsError, err.Error()), nil)
	}
	workspace, err := p.models.PipelineWorkspaceManager.Get(uint(id))
	if err != nil {
		return c.GenerateResponse(errors.New(code.DataNotExists, "获取流水线空间失败: "+err.Error()), nil)
	}
	err = p.models.PipelineWorkspaceManager.Delete(workspace)
	if err != nil {
		err = errors.New(code.DeleteError, "删除流水线空间失败: "+err.Error())
	}
	resp := c.GenerateResponse(err, nil)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationDelete,
		OperateDetail:        "删除流水线空间:" + workspace.Name,
		Scope:                types.ScopePipeline,
		ScopeId:              workspace.ID,
		ScopeName:            workspace.Name,
		ResourceId:           workspace.ID,
		ResourceType:         types.AuditResourcePipeSpace,
		ResourceName:         workspace.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: nil,
	})
	return resp
}

func (p *PipelineWorkspace) latestReleaseVersion(c *views.Context) *utils.Response {
	var ser serializers.WorkspaceReleaseSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	rel, err := p.models.PipelineReleaseManager.GetLatestRelease(ser.WorkspaceId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success, Data: rel}
}

func (p *PipelineWorkspace) existsReleaseVersion(c *views.Context) *utils.Response {
	var ser serializers.WorkspaceReleaseSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	exists, err := p.models.PipelineReleaseManager.ExistsRelease(ser.WorkspaceId, ser.Version)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success, Data: map[string]interface{}{"exists": exists}}
}

func (p *PipelineWorkspace) listGitRepos(c *views.Context) *utils.Response {
	var params schemas.ListGitReposParams
	if err := c.ShouldBind(&params); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return p.workspaceService.ListGitRepos(&params)
}
