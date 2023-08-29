package pipeline_run

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	pipelinerunservice "github.com/kubespace/kubespace/pkg/service/pipeline/pipeline_run"
	"github.com/kubespace/kubespace/pkg/utils"
)

type buildHandler struct {
	models             *model.Models
	pipelineRunService *pipelinerunservice.PipelineRunService
}

func BuildHandler(conf *config.ServerConfig) api.Handler {
	return &buildHandler{
		models:             conf.Models,
		pipelineRunService: conf.ServiceFactory.Pipeline.PipelineRunService,
	}
}

type pipelineBuildBody struct {
	*types.PipelineBuildConfig `json:",inline"`
}

func (h *buildHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	pipelineId, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	pipelineObj, err := h.models.PipelineManager.GetById(pipelineId)
	if err != nil {
		return true, nil, errors.New(code.DataNotExists, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopePipeline,
		ScopeId: pipelineObj.WorkspaceId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *buildHandler) Handle(c *api.Context) *utils.Response {
	var body pipelineBuildBody
	if err := c.ShouldBind(&body); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	pipelineId, _ := utils.ParseUint(c.Param("id"))
	_, err := h.pipelineRunService.Build(pipelineId, body.PipelineBuildConfig, c.User.Name)
	return c.ResponseError(err)
}
