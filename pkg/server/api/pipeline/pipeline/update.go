package pipeline

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	pipelineservice "github.com/kubespace/kubespace/pkg/service/pipeline"
	"github.com/kubespace/kubespace/pkg/service/pipeline/schemas"
	"github.com/kubespace/kubespace/pkg/utils"
)

type updateHandler struct {
	models          *model.Models
	pipelineService *pipelineservice.PipelineService
}

func UpdateHandler(conf *config.ServerConfig) api.Handler {
	return &updateHandler{
		models:          conf.Models,
		pipelineService: conf.ServiceFactory.Pipeline.PipelineService,
	}
}

func (h *updateHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	var body schemas.PipelineBody
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopePipeline,
		ScopeId: body.WorkspaceId,
		Role:    types.RoleEditor,
	}, nil
}

func (h *updateHandler) Handle(c *api.Context) *utils.Response {
	var body schemas.PipelineBody
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	workspace, err := h.models.PipelineWorkspaceManager.Get(body.WorkspaceId)
	if err != nil {
		return c.Response(errors.New(code.DataNotExists, err), nil)
	}
	pipelineObj, err := h.pipelineService.Update(&body, c.User)
	resp := c.Response(err, nil)
	if pipelineObj == nil {
		return resp
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationUpdate,
		OperateDetail:        "更新流水线:" + body.Name,
		Scope:                types.ScopePipeline,
		ScopeId:              workspace.ID,
		ScopeName:            workspace.Name,
		ResourceId:           pipelineObj.ID,
		ResourceType:         types.AuditResourcePipeline,
		ResourceName:         body.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: body,
	})
	return resp
}
