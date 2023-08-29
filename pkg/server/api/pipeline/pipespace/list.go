package pipespace

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/manager/pipeline"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
)

type listHandler struct {
	models *model.Models
}

type listPipespaceForm struct {
	WithPipeline bool   `json:"with_pipeline" form:"with_pipeline"`
	Type         string `json:"type" form:"type"`
}

func ListHandler(conf *config.ServerConfig) api.Handler {
	return &listHandler{models: conf.Models}
}

func (h *listHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, nil, nil
}

func (h *listHandler) Handle(c *api.Context) *utils.Response {
	var form listPipespaceForm
	if err := c.ShouldBind(&form); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	workspaces, err := h.models.PipelineWorkspaceManager.List()
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}
	var data []types.PipelineWorkspace
	for i, w := range workspaces {
		if !h.models.UserRoleManager.AuthRole(c.User, types.ScopePipeline, w.ID, types.RoleViewer) {
			continue
		}
		if form.Type != "" && w.Type != form.Type {
			continue
		}
		if form.WithPipeline {
			workspaces[i].Pipelines, err = h.models.PipelineManager.List(pipeline.ListPipelineCondition{
				WorkspaceId: &w.ID,
			})
			if err != nil {
				return &utils.Response{Code: code.DBError, Msg: err.Error()}
			}
		}
		data = append(data, workspaces[i])
	}
	return c.ResponseOK(data)
}
