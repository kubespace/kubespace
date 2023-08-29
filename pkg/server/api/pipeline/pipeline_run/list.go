package pipeline_run

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

func ListHandler(conf *config.ServerConfig) api.Handler {
	return &listHandler{models: conf.Models}
}

type listPipelineRunForm struct {
	PipelineId      uint   `json:"pipeline_id" form:"pipeline_id"`
	LastBuildNumber int    `json:"last_build_number" form:"last_build_number"`
	Status          string `json:"status" form:"status"`
	Limit           int    `json:"limit" form:"limit" default:"20"`
}

type listPipelineRunData struct {
	PipelineRun *types.PipelineRun        `json:"pipeline_run"`
	StagesRun   []*types.PipelineRunStage `json:"stages_run"`
}

func (h *listHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	var form listPipelineRunForm
	if err := c.ShouldBindQuery(&form); err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	pipelineObj, err := h.models.PipelineManager.GetById(form.PipelineId)
	if err != nil {
		return true, nil, errors.New(code.DataNotExists, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopePipeline,
		ScopeId: pipelineObj.WorkspaceId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *listHandler) Handle(c *api.Context) *utils.Response {
	var form listPipelineRunForm
	if err := c.ShouldBindQuery(&form); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	if form.Limit == 0 {
		form.Limit = 20
	}
	pipelineRuns, err := h.models.PipelineRunManager.ListPipelineRun(pipeline.ListPipelineRunCondition{
		PipelineId:      form.PipelineId,
		LastBuildNumber: form.LastBuildNumber,
		Status:          form.Status,
		Limit:           form.Limit,
	})
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}
	var retData []*listPipelineRunData
	for i, pipelineRun := range pipelineRuns {
		stagesRun, err := h.models.PipelineRunManager.StagesRun(pipelineRun.ID)
		if err != nil {
			return c.ResponseError(errors.New(code.DBError, err))
		}
		retData = append(retData, &listPipelineRunData{
			PipelineRun: pipelineRuns[i],
			StagesRun:   stagesRun,
		})
	}
	return c.ResponseOK(retData)
}
