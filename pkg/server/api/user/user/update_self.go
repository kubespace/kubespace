package user

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
)

type updateSelfHandler struct {
	models *model.Models
}

func UpdateSelfHandler(conf *config.ServerConfig) api.Handler {
	return &updateSelfHandler{models: conf.Models}
}

func (h *updateSelfHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, nil, nil
}

func (h *updateSelfHandler) Handle(c *api.Context) *utils.Response {
	var body userBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}

	userObj, err := h.models.UserManager.GetById(c.User.ID)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, err))
	}

	if body.Status != "" {
		userObj.Status = body.Status
	}
	if body.Password != "" {
		userObj.Password = utils.Encrypt(body.Password)
	}
	if body.Email != "" {
		if ok := utils.VerifyEmailFormat(body.Email); !ok {
			return c.ResponseError(errors.New(code.ParamsError, fmt.Sprintf("email:%s format error for user:%s", body.Email, userObj.Name)))
		}
		userObj.Email = body.Email
	}

	if err := h.models.UserManager.Update(userObj); err != nil {
		return c.ResponseError(errors.New(code.UpdateError, err))
	}
	return c.ResponseOK(nil)
}
