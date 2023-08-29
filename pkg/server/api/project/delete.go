package project

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	projectservice "github.com/kubespace/kubespace/pkg/service/project"
	"github.com/kubespace/kubespace/pkg/utils"
)

type deleteHandler struct {
	models         *model.Models
	projectService *projectservice.ProjectService
}

func DeleteHandler(conf *config.ServerConfig) api.Handler {
	return &deleteHandler{
		models:         conf.Models,
		projectService: conf.ServiceFactory.Project.ProjectService,
	}
}

type projectDeleteBody struct {
	DelResource bool `json:"del_resource" form:"del_resource"`
}

func (h *deleteHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	projectId, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopeProject,
		ScopeId: projectId,
		Role:    types.RoleAdmin,
	}, nil
}

func (h *deleteHandler) Handle(c *api.Context) *utils.Response {
	var body projectDeleteBody
	if err := c.ShouldBind(&body); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	projectId, _ := utils.ParseUint(c.Param("id"))

	project, err := h.projectService.Delete(projectId, body.DelResource)
	resp := c.ResponseError(err)

	if project != nil {
		c.CreateAudit(&types.AuditOperate{
			Operation:            types.AuditOperationDelete,
			OperateDetail:        "删除工作空间：" + project.Name,
			Scope:                types.ScopeProject,
			ScopeId:              project.ID,
			ScopeName:            project.Name,
			ResourceId:           project.ID,
			ResourceType:         types.AuditResourceProject,
			ResourceName:         project.Name,
			Code:                 resp.Code,
			Message:              resp.Msg,
			OperateDataInterface: body,
		})
	}
	return resp
}
