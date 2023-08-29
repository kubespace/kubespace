package user

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
	"time"
)

type createAdminHandler struct {
	models *model.Models
}

func CreateAdminHandler(conf *config.ServerConfig) api.Handler {
	return &createAdminHandler{models: conf.Models}
}

func (h *createAdminHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return false, nil, nil
}

func (h *createAdminHandler) Handle(c *api.Context) *utils.Response {
	var ser userBody
	if err := c.ShouldBind(&ser); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}

	user := &types.User{
		Name:       types.ADMIN,
		Email:      ser.Email,
		Password:   utils.Encrypt(ser.Password),
		IsSuper:    true,
		Status:     "normal",
		LastLogin:  time.Now(),
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	if err := h.models.UserManager.Create(user); err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}
	return c.ResponseOK(nil)
}
