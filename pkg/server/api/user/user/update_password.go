package user

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
)

type updatePasswordHandler struct {
	models *model.Models
}

func UpdatePasswordHandler(conf *config.ServerConfig) api.Handler {
	return &updatePasswordHandler{models: conf.Models}
}

type updatePasswordBody struct {
	OriginPassword string `json:"origin_password" form:"origin_password"`
	NewPassword    string `json:"new_password" form:"new_password"`
}

func (h *updatePasswordHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, nil, nil
}

func (h *updatePasswordHandler) Handle(c *api.Context) *utils.Response {
	var body updatePasswordBody
	if err := c.ShouldBind(&body); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	userObj, err := h.models.UserManager.GetById(c.User.ID)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, err))
	}
	if userObj.Password != utils.Encrypt(body.OriginPassword) {
		return c.ResponseError(errors.New(code.ParamsError, "原密码不正确，请重新输入"))
	}
	userObj.Password = utils.Encrypt(body.NewPassword)

	if err = h.models.UserManager.Update(userObj); err != nil {
		return c.ResponseError(errors.New(code.UpdateError, "更新密码失败："+err.Error()))
	}
	return c.ResponseOK(nil)
}
