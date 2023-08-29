package pipeline

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	pipelineservice "github.com/kubespace/kubespace/pkg/service/pipeline"
	"github.com/kubespace/kubespace/pkg/utils"
)

type deleteHandler struct {
	models          *model.Models
	pipelineService *pipelineservice.PipelineService
}

func DeleteHandler(conf *config.ServerConfig) api.Handler {
	return &deleteHandler{
		models:          conf.Models,
		pipelineService: conf.ServiceFactory.Pipeline.PipelineService,
	}
}

func (h *deleteHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	id, err := utils.ParseUint(c.Param("pipelineId"))
	if err != nil {
		return true, nil, errors.New(code.ParseError, err)
	}
	pipelineObj, err := h.models.PipelineManager.GetById(id)
	if err != nil {
		return true, nil, errors.New(code.DataNotExists, fmt.Sprintf("获取流水线id=%d失败：%s", id, err.Error()))
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopePipeline,
		ScopeId: pipelineObj.WorkspaceId,
		Role:    types.RoleEditor,
	}, nil
}

func (h *deleteHandler) Handle(c *api.Context) *utils.Response {
	id, _ := utils.ParseUint(c.Param("id"))
	pipelineObj, err := h.models.PipelineManager.GetById(id)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, fmt.Sprintf("获取流水线id=%d失败：%s", id, err.Error())))
	}
	workspace, err := h.models.PipelineWorkspaceManager.Get(pipelineObj.WorkspaceId)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, fmt.Sprintf("获取流水线空间id=%d失败：%s", pipelineObj.WorkspaceId, err.Error())))
	}
	err = h.models.PipelineManager.Delete(id)
	if err != nil {
		err = errors.New(code.DeleteError, "删除流水线失败："+err.Error())
	}
	resp := c.Response(err, nil)
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
