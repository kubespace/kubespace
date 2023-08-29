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

type logoutHandler struct {
	models *model.Models
}

func LogoutHandler(conf *config.ServerConfig) api.Handler {
	return &logoutHandler{models: conf.Models}
}

func (h *logoutHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return false, nil, nil
}

func (h *logoutHandler) Handle(c *api.Context) *utils.Response {
	sessionId, err := c.Cookie("sessionId")
	if err != nil {
		return c.ResponseError(errors.New(code.CookieError, fmt.Sprintf("get auth cookie session error: %v", err)))
	}
	if sessionId != "" {
		if err := h.models.SessionManager.Delete(sessionId); err != nil {
			return c.ResponseError(errors.New(code.DeleteError, err))
		}
	}

	return c.ResponseOK(nil)
}
