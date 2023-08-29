package user

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/manager"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
)

type hasAdminHandler struct {
	models *model.Models
}

func HasAdminHandler(conf *config.ServerConfig) api.Handler {
	return &hasAdminHandler{models: conf.Models}
}

type hasAdminData struct {
	Has int `json:"has"`
}

func (h *hasAdminHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return false, nil, nil
}

func (h *hasAdminHandler) Handle(c *api.Context) *utils.Response {
	data := hasAdminData{Has: 1}
	admin, err := h.models.UserManager.GetByName(types.ADMIN, manager.NotFoundReturnNil)
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}
	if admin == nil {
		data.Has = 0
	}
	return c.ResponseOK(data)
}
