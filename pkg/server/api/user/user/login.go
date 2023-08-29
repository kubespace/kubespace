package user

import (
	"github.com/google/uuid"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
	"time"
)

type loginHandler struct {
	models *model.Models
}

func LoginHandler(conf *config.ServerConfig) api.Handler {
	return &loginHandler{models: conf.Models}
}

type loginForm struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

func (h *loginHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return false, nil, nil
}

func (h *loginHandler) Handle(c *api.Context) *utils.Response {
	var form loginForm
	if err := c.ShouldBindJSON(&form); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	if form.UserName == "" || form.Password == "" {
		return c.ResponseError(errors.New(code.ParamsError, "用户名或密码为空"))
	}
	password := utils.Encrypt(form.Password)

	userObj, err := h.models.UserManager.GetByName(form.UserName)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, "未找到该用户"))
	}

	if password != userObj.Password {
		return c.ResponseError(errors.New(code.AuthError, "密码错误"))
	}

	tkObj := types.UserSession{
		UserName:  form.UserName,
		SessionId: uuid.New(),
	}
	if err = h.models.SessionManager.Create(&tkObj); err != nil {
		return c.ResponseError(errors.New(code.RedisError, "创建用户session错误："+err.Error()))
	}

	userObj.LastLogin = time.Now()
	if err = h.models.UserManager.Update(userObj); err != nil {
		return c.ResponseError(errors.New(code.UpdateError, err))
	}
	//c.SetCookie(api.SessionId, tkObj.SessionId.String(), 60*60*12, "", "*", false, true)
	return c.ResponseOK(map[string]interface{}{
		"token": tkObj.SessionId.String(),
	})
}
