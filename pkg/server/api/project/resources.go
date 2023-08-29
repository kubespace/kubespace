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

// 工作空间下的资源
type resourcesHandler struct {
	models         *model.Models
	projectService *projectservice.ProjectService
}

func ResourcesHandler(conf *config.ServerConfig) api.Handler {
	return &resourcesHandler{
		models:         conf.Models,
		projectService: conf.ServiceFactory.Project.ProjectService,
	}
}

type projectResourcesForm struct {
	ProjectId uint `json:"project_id" form:"project_id"`
}

func (h *resourcesHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	var form projectResourcesForm
	if err := c.ShouldBindQuery(&form); err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopeProject,
		ScopeId: form.ProjectId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *resourcesHandler) Handle(c *api.Context) *utils.Response {
	var form projectResourcesForm
	if err := c.ShouldBindQuery(&form); err != nil {
		return c.ResponseError(errors.New(code.ParseError, err))
	}
	return h.projectService.GetProjectNamespaceResources(form.ProjectId)
}
