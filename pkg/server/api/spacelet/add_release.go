package spacelet

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	spaceletservice "github.com/kubespace/kubespace/pkg/service/spacelet"
	"github.com/kubespace/kubespace/pkg/utils"
)

type addReleaseHandler struct {
	models          *model.Models
	spaceletService *spaceletservice.SpaceletService
}

func AddReleaseHandler(conf *config.ServerConfig) api.Handler {
	return &addReleaseHandler{
		models:          conf.Models,
		spaceletService: conf.ServiceFactory.Pipeline.SpaceletService,
	}
}

// addReleaseVersionParams 发布阶段执行时添加版本
type addReleaseVersionBody struct {
	WorkspaceId uint   `json:"workspace_id"`
	JobId       uint   `json:"job_id"`
	Version     string `json:"version"`
}

func (h *addReleaseHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return false, nil, nil
}

func (h *addReleaseHandler) Handle(c *api.Context) *utils.Response {
	var body addReleaseVersionBody
	if err := c.ShouldBind(&body); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	if err := h.models.PipelineReleaseManager.Add(body.WorkspaceId, body.Version, body.JobId); err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}
	return c.ResponseOK(nil)
}
