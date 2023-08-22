package pipeline

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/utils"
	"net/http"
	"strconv"
	"time"
)

type PipelineResource struct {
	Views  []*views.View
	models *model.Models
}

func NewPipelineResource(config *config.ServerConfig) *PipelineResource {
	pipelineWs := &PipelineResource{
		models: config.Models,
	}
	pipelineWs.Views = []*views.View{
		views.NewView(http.MethodGet, "/:workspaceId", pipelineWs.list),
		views.NewView(http.MethodPost, "", pipelineWs.create),
		views.NewView(http.MethodPut, "/:id", pipelineWs.update),
		views.NewView(http.MethodDelete, "/:id", pipelineWs.delete),
	}
	return pipelineWs
}

func (r *PipelineResource) create(c *views.Context) *utils.Response {
	var ser serializers.PipelineResourceSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	workspace, err := r.models.PipelineWorkspaceManager.Get(ser.WorkspaceId)
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DataNotExists, "获取流水线空间失败："+err.Error()))
	}
	resource := &types.PipelineResource{
		WorkspaceId: ser.WorkspaceId,
		Name:        ser.Name,
		Type:        ser.Type,
		Value:       ser.Value,
		Global:      ser.Global,
		SecretId:    ser.SecretId,
		Description: ser.Description,
		CreateUser:  c.User.Name,
		UpdateUser:  c.User.Name,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
	resp := c.GenerateResponseOK(nil)
	if _, err = r.models.PipelineResourceManager.Create(resource); err != nil {
		resp = c.GenerateResponseError(errors.New(code.DBError, "创建资源失败: "+err.Error()))
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationCreate,
		OperateDetail:        "创建流水线资源:" + ser.Name,
		Scope:                types.ScopePipeline,
		ScopeId:              workspace.ID,
		ScopeName:            workspace.Name,
		ResourceId:           resource.ID,
		ResourceType:         types.AuditResourcePipelineResource,
		ResourceName:         ser.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: ser,
	})
	return resp
}

func (r *PipelineResource) update(c *views.Context) *utils.Response {
	var ser serializers.PipelineResourceSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	resId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	resource, err := r.models.PipelineResourceManager.Get(uint(resId))
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DataNotExists, "获取资源失败: "+err.Error()))
	}
	workspace, err := r.models.PipelineWorkspaceManager.Get(resource.WorkspaceId)
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DataNotExists, "获取流水线空间失败："+err.Error()))
	}
	resource.Value = ser.Value
	resource.Description = ser.Description
	resource.Global = ser.Global
	resource.SecretId = ser.SecretId
	resource.UpdateUser = c.User.Name
	resource.UpdateTime = time.Now()

	resp := c.GenerateResponseOK(nil)
	if _, err = r.models.PipelineResourceManager.Update(resource); err != nil {
		resp = c.GenerateResponseError(errors.New(code.DBError, "更新资源失败: "+err.Error()))
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationUpdate,
		OperateDetail:        "更新流水线资源:" + resource.Name,
		Scope:                types.ScopePipeline,
		ScopeId:              workspace.ID,
		ScopeName:            workspace.Name,
		ResourceId:           resource.ID,
		ResourceType:         types.AuditResourcePipelineResource,
		ResourceName:         resource.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: ser,
	})
	return resp
}

func (r *PipelineResource) list(c *views.Context) *utils.Response {
	workspaceId, err := strconv.ParseUint(c.Param("workspaceId"), 10, 64)
	if err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	resources, err := r.models.PipelineResourceManager.List(uint(workspaceId))
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DBError, err))
	}
	return c.GenerateResponseOK(resources)
}

func (r *PipelineResource) delete(c *views.Context) *utils.Response {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	res, err := r.models.PipelineResourceManager.Get(uint(id))
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DataNotExists, "获取资源失败: "+err.Error()))
	}
	workspace, err := r.models.PipelineWorkspaceManager.Get(res.WorkspaceId)
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DataNotExists, "获取流水线空间失败："+err.Error()))
	}
	resp := c.GenerateResponseOK(nil)
	if err = r.models.PipelineResourceManager.Delete(res); err != nil {
		resp = c.GenerateResponseError(errors.New(code.DBError, "删除流水线资源失败: "+err.Error()))
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationDelete,
		OperateDetail:        "删除流水线资源:" + res.Name,
		Scope:                types.ScopePipeline,
		ScopeId:              workspace.ID,
		ScopeName:            workspace.Name,
		ResourceId:           res.ID,
		ResourceType:         types.AuditResourcePipelineResource,
		ResourceName:         res.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: nil,
	})
	return resp
}
