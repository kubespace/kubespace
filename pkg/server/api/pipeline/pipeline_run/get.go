package pipeline_run

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
)

type getHandler struct {
	models *model.Models
}

func GetHandler(conf *config.ServerConfig) api.Handler {
	return &getHandler{models: conf.Models}
}

type getPipelineRunWorkspaceData struct {
	Id      uint   `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	CodeUrl string `json:"code_url"`
}

type getPipelineRunData struct {
	Pipeline    *types.Pipeline              `json:"pipeline"`
	PipelineRun *types.PipelineRun           `json:"pipeline_run"`
	StagesRun   []*types.PipelineRunStage    `json:"stages_run"`
	Workspace   *getPipelineRunWorkspaceData `json:"workspace"`
}

func (h *getHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	workspaceId, _ := utils.ParseUint(c.Query("workspace_id"))
	return true, &api.AuthPerm{
		Scope:   types.ScopePipeline,
		ScopeId: workspaceId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *getHandler) Handle(c *api.Context) *utils.Response {
	pipelineRunId, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	pipelineRun, err := h.models.PipelineRunManager.Get(pipelineRunId)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, err))
	}
	stagesRun, err := h.models.PipelineRunManager.StagesRun(pipelineRunId)
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}
	pipelineObj, err := h.models.PipelineManager.GetById(pipelineRun.PipelineId)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, err))
	}
	workspace, err := h.models.PipelineWorkspaceManager.Get(pipelineObj.WorkspaceId)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, err))
	}
	cloneUrl := ""
	if workspace.Code != nil {
		cloneUrl = workspace.Code.CloneUrl
	}
	return c.ResponseOK(&getPipelineRunData{
		Pipeline:    pipelineObj,
		PipelineRun: pipelineRun,
		StagesRun:   stagesRun,
		Workspace: &getPipelineRunWorkspaceData{
			Id:      workspace.ID,
			Name:    workspace.Name,
			Type:    workspace.Type,
			CodeUrl: cloneUrl,
		},
	})
}
