package pipeline

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	pipelineservice "github.com/kubespace/kubespace/pkg/service/pipeline"
	"github.com/kubespace/kubespace/pkg/utils"
)

type getHandler struct {
	models          *model.Models
	pipelineService *pipelineservice.PipelineService
}

func GetHandler(conf *config.ServerConfig) api.Handler {
	return &getHandler{
		models:          conf.Models,
		pipelineService: conf.ServiceFactory.Pipeline.PipelineService,
	}
}

func (h *getHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	pipelineId, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	pipeline, err := h.models.PipelineManager.GetById(pipelineId)
	if err != nil {
		return true, nil, errors.New(code.DataNotExists, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopePipeline,
		ScopeId: pipeline.WorkspaceId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *getHandler) Handle(c *api.Context) *utils.Response {
	pipelineId, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	return h.pipelineService.GetPipeline(pipelineId)
}
