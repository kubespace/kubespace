package pipeline

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/manager/pipeline"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	pipelineservice "github.com/kubespace/kubespace/pkg/service/pipeline"
	"github.com/kubespace/kubespace/pkg/utils"
)

type listHandler struct {
	models          *model.Models
	pipelineService *pipelineservice.PipelineService
}

type listPipelineForm struct {
	WorkspaceId uint `json:"workspace_id" form:"workspace_id"`
}

type listPipelineData struct {
	Pipeline  *types.Pipeline    `json:"pipeline"`
	LastBuild *types.PipelineRun `json:"last_build"`
}

func ListHandler(conf *config.ServerConfig) api.Handler {
	return &listHandler{
		models:          conf.Models,
		pipelineService: conf.ServiceFactory.Pipeline.PipelineService,
	}
}

func (h *listHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	var form listPipelineForm
	if err := c.ShouldBindQuery(&form); err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopePipeline,
		ScopeId: form.WorkspaceId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *listHandler) Handle(c *api.Context) *utils.Response {
	var form listPipelineForm
	if err := c.ShouldBindQuery(&form); err != nil {
		return c.ResponseError(errors.New(code.ParseError, err))
	}
	pipelines, err := h.models.PipelineManager.List(pipeline.ListPipelineCondition{
		WorkspaceId: &form.WorkspaceId,
	})
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, fmt.Sprintf("获取流水线列表错误: %v", err)))
	}
	var retData []*listPipelineData
	for i, obj := range pipelines {
		lastPipelineRun, err := h.models.PipelineRunManager.GetLastPipelineRun(obj.ID)
		if err != nil {
			return c.ResponseError(errors.New(code.DBError, fmt.Sprintf("获取流水线构建列表错误: %v", err)))
		}
		retData = append(retData, &listPipelineData{
			Pipeline:  pipelines[i],
			LastBuild: lastPipelineRun,
		})
	}
	return c.ResponseOK(retData)
}
