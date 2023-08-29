package project

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	projectservice "github.com/kubespace/kubespace/pkg/service/project"
	"github.com/kubespace/kubespace/pkg/utils"
	"time"
)

type cloneHandler struct {
	models         *model.Models
	projectService *projectservice.ProjectService
}

func CloneHandler(conf *config.ServerConfig) api.Handler {
	return &cloneHandler{
		models:         conf.Models,
		projectService: conf.ServiceFactory.Project.ProjectService,
	}
}

type projectCloneAppBody struct {
	Id uint `json:"id" form:"id"`
}

type projectCloneResourceBody struct {
	Type string `json:"type" form:"type"`
	Name string `json:"name" form:"name"`
}

type projectCloneBody struct {
	OriginProjectId uint                        `json:"origin_project_id" form:"origin_project_id"`
	Name            string                      `json:"name" form:"name"`
	Description     string                      `json:"description" form:"description"`
	ClusterId       string                      `json:"cluster_id" form:"cluster_id"`
	Namespace       string                      `json:"namespace" form:"namespace"`
	Owner           string                      `json:"owner" form:"owner"`
	Resources       []*projectCloneResourceBody `json:"resources" form:"resources"`
	Apps            []*projectCloneAppBody      `json:"apps" form:"apps"`
}

func (h *cloneHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, &api.AuthPerm{
		Scope:   types.ScopePlatform,
		ScopeId: 0,
		Role:    types.RoleEditor,
	}, nil
}

func (h *cloneHandler) Handle(c *api.Context) *utils.Response {
	var body projectCloneBody
	if err := c.ShouldBind(&body); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	sourceProject, err := h.models.ProjectManager.Get(body.OriginProjectId)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, "获取源工作空间失败:"+err.Error()))
	}
	newProject := &types.Project{
		Name:        body.Name,
		Description: body.Description,
		ClusterId:   body.ClusterId,
		Namespace:   body.Namespace,
		Owner:       body.Owner,
		CreateUser:  c.User.Name,
		UpdateUser:  c.User.Name,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
	newProject, err = h.projectService.Clone(sourceProject.ID, newProject)

	resp := c.ResponseError(err)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationClone,
		OperateDetail:        fmt.Sprintf("克隆源工作空间：%s，到目标工作空间：%s", sourceProject.Name, body.Name),
		Scope:                types.ScopeProject,
		ScopeId:              sourceProject.ID,
		ScopeName:            sourceProject.Name,
		ResourceId:           sourceProject.ID,
		ResourceType:         types.AuditResourceProject,
		ResourceName:         sourceProject.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: body,
	})
	return resp
}
