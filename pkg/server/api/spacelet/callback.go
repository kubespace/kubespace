package spacelet

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/service/pipeline/pipeline_run"
	"github.com/kubespace/kubespace/pkg/utils"
)

type callbackHandler struct {
	models             *model.Models
	pipelineRunService *pipeline_run.PipelineRunService
}

func CallbackHandler(conf *config.ServerConfig) api.Handler {
	return &callbackHandler{
		models:             conf.Models,
		pipelineRunService: conf.ServiceFactory.Pipeline.PipelineRunService,
	}
}

// JobCallbackParams 任务执行完成回调参数
type jobCallbackBody struct {
	JobId  uint   `json:"job_id"`
	Status string `json:"status"`
}

func (h *callbackHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return false, nil, nil
}

func (h *callbackHandler) Handle(c *api.Context) *utils.Response {
	var form jobCallbackBody
	if err := c.ShouldBind(&form); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	err := h.pipelineRunService.JobCallback(form.JobId, form.Status)
	return c.ResponseError(err)
}
