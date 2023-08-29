package pipeline_run

import (
	"fmt"
	"github.com/gin-gonic/gin/binding"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/service/pipeline/pipeline_run"
	pipelinerunservice "github.com/kubespace/kubespace/pkg/service/pipeline/pipeline_run"
	"github.com/kubespace/kubespace/pkg/utils"
)

type stageActionHandler struct {
	models             *model.Models
	pipelineRunService *pipelinerunservice.PipelineRunService
}

func StageActionHandler(conf *config.ServerConfig) api.Handler {
	return &stageActionHandler{
		models:             conf.Models,
		pipelineRunService: conf.ServiceFactory.Pipeline.PipelineRunService,
	}
}

type pipelineStageActionBody struct {
	WorkspaceId  uint                            `json:"workspace_id"`
	Action       string                          `json:"action"`
	StageRunId   uint                            `json:"stage_run_id"`
	CustomParams map[string]interface{}          `json:"custom_params"`
	JobParams    map[uint]map[string]interface{} `json:"job_params"`
}

func (h *stageActionHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	var body pipelineStageActionBody
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopePipeline,
		ScopeId: body.WorkspaceId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *stageActionHandler) Handle(c *api.Context) *utils.Response {
	var body pipelineStageActionBody
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	pipelineRun, stageRun, err := h.pipelineRunService.StageAction(body.Action, body.StageRunId, pipeline_run.StageActionParams{
		CustomParams: body.CustomParams,
		JobParams:    body.JobParams,
	})
	resp := c.ResponseError(err)
	if pipelineRun == nil {
		return resp
	}
	pipelineObj, err := h.models.PipelineManager.GetById(pipelineRun.PipelineId)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, err))
	}
	workspace, err := h.models.PipelineWorkspaceManager.Get(pipelineObj.WorkspaceId)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, err))
	}
	actionNameMap := map[string]string{
		pipeline_run.StageActionManualExec:   "手动执行",
		pipeline_run.StageActionCancel:       "取消执行",
		pipeline_run.StageActionCancelReexec: "重新执行取消",
		pipeline_run.StageActionReexec:       "重新执行",
		pipeline_run.StageActionErrorRetry:   "重试失败",
	}
	opDetail := fmt.Sprintf("%s流水线「%s(#%d)」构建阶段：%s", actionNameMap[body.Action], pipelineObj.Name, pipelineRun.BuildNumber, stageRun.Name)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationUpdate,
		OperateDetail:        opDetail,
		Scope:                types.ScopePipeline,
		ScopeId:              workspace.ID,
		ScopeName:            workspace.Name,
		ResourceId:           pipelineRun.ID,
		ResourceType:         types.AuditResourcePipelineBuild,
		ResourceName:         fmt.Sprintf("%s(#%d)", pipelineObj.Name, pipelineRun.BuildNumber),
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: body,
	})
	return resp
}
