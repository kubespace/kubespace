package user

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/manager/user"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
)

type tokenHandler struct {
	models *model.Models
}

func TokenHandler(conf *config.ServerConfig) api.Handler {
	return &tokenHandler{models: conf.Models}
}

func (h *tokenHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, nil, nil
}

func (h *tokenHandler) Handle(c *api.Context) *utils.Response {
	userObj, err := h.models.UserManager.GetById(c.User.ID)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, err))
	}
	roles, err := h.models.UserRoleManager.List(&user.ListUserRoleCondition{UserId: &c.User.ID})
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}
	return c.ResponseOK(map[string]interface{}{
		"id":          c.User.ID,
		"name":        c.User.Name,
		"permissions": roles,
		"is_super":    userObj.IsSuper,
	})
}
