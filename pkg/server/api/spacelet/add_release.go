package spacelet

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/service/pipeline/schemas"
	"github.com/kubespace/kubespace/pkg/utils"
)

type addReleaseHandler struct {
	models *model.Models
}

func AddReleaseHandler(conf *config.ServerConfig) api.Handler {
	return &addReleaseHandler{models: conf.Models}
}

func (h *addReleaseHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return false, nil, nil
}

func (h *addReleaseHandler) Handle(c *api.Context) *utils.Response {
	var body schemas.AddReleaseVersionParams
	if err := c.ShouldBind(&body); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	if err := h.models.PipelineReleaseManager.Add(body.WorkspaceId, body.Version, body.JobId); err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}
	return c.ResponseOK(nil)
}
