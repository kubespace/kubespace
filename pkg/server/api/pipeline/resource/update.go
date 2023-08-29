package resource

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
	"time"
)

type updateHandler struct {
	models *model.Models
}

func UpdateHandler(conf *config.ServerConfig) api.Handler {
	return &updateHandler{models: conf.Models}
}

func (h *updateHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	resId, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	resource, err := h.models.PipelineResourceManager.Get(resId)
	if err != nil {
		return true, nil, errors.New(code.DataNotExists, "获取资源失败: "+err.Error())
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopePipeline,
		ScopeId: resource.WorkspaceId,
		Role:    types.RoleEditor,
	}, nil
}

func (h *updateHandler) Handle(c *api.Context) *utils.Response {
	var ser pipelineResourceBody
	if err := c.ShouldBind(&ser); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	resId, _ := utils.ParseUint(c.Param("id"))
	resource, err := h.models.PipelineResourceManager.Get(resId)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, "获取资源失败: "+err.Error()))
	}
	workspace, err := h.models.PipelineWorkspaceManager.Get(resource.WorkspaceId)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, "获取流水线空间失败："+err.Error()))
	}
	resource.Value = ser.Value
	resource.Description = ser.Description
	resource.Global = ser.Global
	resource.SecretId = ser.SecretId
	resource.UpdateUser = c.User.Name
	resource.UpdateTime = time.Now()

	resp := c.ResponseOK(nil)
	if _, err = h.models.PipelineResourceManager.Update(resource); err != nil {
		resp = c.ResponseError(errors.New(code.DBError, "更新资源失败: "+err.Error()))
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
