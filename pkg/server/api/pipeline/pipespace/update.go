package pipespace

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

type updateWorkspaceBody struct {
	Description  string `json:"description" form:"description"`
	CodeSecretId uint   `json:"code_secret_id" form:"code_secret_id"`
}

func (h *updateHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	id, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopePipeline,
		ScopeId: id,
		Role:    types.RoleEditor,
	}, nil
}

func (h *updateHandler) Handle(c *api.Context) *utils.Response {
	var body updateWorkspaceBody
	if err := c.ShouldBind(&body); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}

	id, _ := utils.ParseUint(c.Param("id"))
	workspace, err := h.models.PipelineWorkspaceManager.Get(id)
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, "获取流水线空间失败: "+err.Error()))
	}

	if body.CodeSecretId != 0 && workspace.Code != nil {
		workspace.Code.SecretId = body.CodeSecretId
	}
	if body.Description != "" {
		workspace.Description = body.Description
	}
	workspace.UpdateUser = c.User.Name
	workspace.UpdateTime = time.Now()
	_, err = h.models.PipelineWorkspaceManager.Update(workspace)
	if err != nil {
		err = errors.New(code.DBError, "更新流水线空间失败: "+err.Error())
	}
	resp := c.Response(err, nil)
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
		OperateDataInterface: body,
	})
	return resp
}
