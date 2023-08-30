package spacelet

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	spaceletmanager "github.com/kubespace/kubespace/pkg/model/manager/spacelet"
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

func (h *listHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, nil, nil
}

func (h *listHandler) Handle(c *api.Context) *utils.Response {
	spacelets, err := h.models.SpaceletManager.List(&spaceletmanager.SpaceletListCondition{})
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}
	for _, sp := range spacelets {
		sp.Token = ""
	}
	return c.ResponseOK(spacelets)
}
