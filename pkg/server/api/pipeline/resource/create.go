package resource

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
	"time"
)

type createHandler struct {
	models *model.Models
}

func CreateHandler(conf *config.ServerConfig) api.Handler {
	return &createHandler{models: conf.Models}
}

type pipelineResourceBody struct {
	WorkspaceId uint   `json:"workspace_id" form:"workspace_id"`
	Global      bool   `json:"global" form:"global"`
	Name        string `json:"name" form:"name"`
	Type        string `json:"type" form:"type"`
	Value       string `json:"value" form:"value"`
	SecretId    uint   `json:"secret_id" form:"secret_id"`
	Description string `json:"description" form:"description"`
}

func (h *createHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	var body pipelineResourceBody
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopePipeline,
		ScopeId: body.WorkspaceId,
		Role:    types.RoleEditor,
	}, nil
}

func (h *createHandler) Handle(c *api.Context) *utils.Response {
	var body pipelineResourceBody
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	workspace, err := h.models.PipelineWorkspaceManager.Get(body.WorkspaceId)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, "获取流水线空间失败："+err.Error()))
	}
	resource := &types.PipelineResource{
		WorkspaceId: body.WorkspaceId,
		Name:        body.Name,
		Type:        body.Type,
		Value:       body.Value,
		Global:      body.Global,
		SecretId:    body.SecretId,
		Description: body.Description,
		CreateUser:  c.User.Name,
		UpdateUser:  c.User.Name,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
	_, err = h.models.PipelineResourceManager.Create(resource);
	if err != nil {
		err = errors.New(code.DBError, "创建资源失败: "+err.Error())
	}
	resp := c.ResponseError(err)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationCreate,
		OperateDetail:        "创建流水线资源:" + body.Name,
		Scope:                types.ScopePipeline,
		ScopeId:              workspace.ID,
		ScopeName:            workspace.Name,
		ResourceId:           resource.ID,
		ResourceType:         types.AuditResourcePipelineResource,
		ResourceName:         body.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: body,
	})
	return resp
}
