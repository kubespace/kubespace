package pipespace

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
	id, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopePipeline,
		ScopeId: id,
		Role:    types.RoleAdmin,
	}, nil
}

func (h *deleteHandler) Handle(c *api.Context) *utils.Response {
	id, _ := utils.ParseUint(c.Param("id"))
	workspace, err := h.models.PipelineWorkspaceManager.Get(id)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, "获取流水线空间失败: "+err.Error()))
	}
	err = h.models.PipelineWorkspaceManager.Delete(workspace)
	if err != nil {
		err = errors.New(code.DeleteError, "删除流水线空间失败: "+err.Error())
	}
	resp := c.Response(err, nil)
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
