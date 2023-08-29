package pipeline

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	pipelineservice "github.com/kubespace/kubespace/pkg/service/pipeline"
	"github.com/kubespace/kubespace/pkg/utils"
)

type listRepoBranchHandler struct {
	models          *model.Models
	pipelineService *pipelineservice.PipelineService
}

func ListRepoBranchHandler(conf *config.ServerConfig) api.Handler {
	return &listRepoBranchHandler{
		models:          conf.Models,
		pipelineService: conf.ServiceFactory.Pipeline.PipelineService,
	}
}

func (h *listRepoBranchHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	id, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParseError, err)
	}
	pipelineObj, err := h.models.PipelineManager.GetById(id)
	if err != nil {
		return true, nil, errors.New(code.DataNotExists, fmt.Sprintf("获取流水线id=%d失败：%s", id, err.Error()))
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopePipeline,
		ScopeId: pipelineObj.WorkspaceId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *listRepoBranchHandler) Handle(c *api.Context) *utils.Response {
	pipelineId, _ := utils.ParseUint(c.Param("id"))
	branches, err := h.pipelineService.ListRepoBranches(pipelineId)
	if err != nil {
		return c.ResponseError(err)
	}
	return c.ResponseOK(branches)
}
