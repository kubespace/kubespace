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

type createHandler struct {
	models *model.Models
}

func CreateHandler(conf *config.ServerConfig) api.Handler {
	return &createHandler{models: conf.Models}
}

type createProjectBody struct {
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
	ClusterId   string `json:"cluster_id" form:"cluster_id"`
	Namespace   string `json:"namespace" form:"namespace"`
	Owner       string `json:"owner" form:"owner"`
}

func (h *createHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, &api.AuthPerm{
		Scope:   types.ScopePlatform,
		ScopeId: 0,
		Role:    types.RoleEditor,
	}, nil
}

func (h *createHandler) Handle(c *api.Context) *utils.Response {
	var body createProjectBody
	if err := c.ShouldBind(&body); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	_, err := h.models.ClusterManager.GetByName(body.ClusterId)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, "get cluster error: "+err.Error()))
	}
	project := &types.Project{
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
	project, err = h.models.ProjectManager.Create(project)
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, "创建项目空间失败: "+err.Error()))
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationCreate,
		OperateDetail:        "创建工作空间，名称=" + project.Name,
		Scope:                types.ScopeProject,
		ScopeId:              project.ID,
		ScopeName:            project.Name,
		ResourceId:           project.ID,
		ResourceType:         types.AuditResourceProject,
		ResourceName:         project.Name,
		Code:                 code.Success,
		OperateDataInterface: project,
	})
	return c.ResponseOK(project)
}
