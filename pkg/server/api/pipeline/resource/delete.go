package resource

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
)

type deleteHandler struct {
	models *model.Models
}

func DeleteHandler(conf *config.ServerConfig) api.Handler {
	return &deleteHandler{models: conf.Models}
}

func (h *deleteHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
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

func (h *deleteHandler) Handle(c *api.Context) *utils.Response {
	id, _ := utils.ParseUint(c.Param("id"))
	res, _ := h.models.PipelineResourceManager.Get(id)
	workspace, err := h.models.PipelineWorkspaceManager.Get(res.WorkspaceId)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, "获取流水线空间失败："+err.Error()))
	}
	err = h.models.PipelineResourceManager.Delete(res)
	if err != nil {
		err = errors.New(code.DBError, "删除流水线资源失败: "+err.Error())
	}
	resp := c.ResponseError(err)
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
