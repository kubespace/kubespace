package project

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
	projectId, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopeProject,
		ScopeId: projectId,
		Role:    types.RoleEditor,
	}, nil
}

func (h *updateHandler) Handle(c *api.Context) *utils.Response {
	projectId, _ := utils.ParseUint(c.Param("id"))
	var body createProjectBody
	if err := c.ShouldBind(&body); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	project, err := h.models.ProjectManager.Get(projectId)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, err))
	}
	project.Name = body.Name
	project.Description = body.Description
	project.Owner = body.Owner
	project.UpdateTime = time.Now()
	project.UpdateUser = c.User.Name
	project, err = h.models.ProjectManager.Update(project)
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, "更新项目空间失败:%s"+err.Error()))
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationUpdate,
		OperateDetail:        "更新工作空间：" + project.Name,
		Scope:                types.ScopeProject,
		ScopeId:              project.ID,
		ScopeName:            project.Name,
		ResourceId:           project.ID,
		ResourceType:         types.AuditResourceProject,
		ResourceName:         project.Name,
		Code:                 code.Success,
		OperateDataInterface: body,
	})
	return c.ResponseOK(project)
}
