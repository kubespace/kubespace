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
	dList, err := h.models.UserManager.List(user.UserListCondition{})
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}
	var data []map[string]interface{}

	for _, du := range dList {
		data = append(data, map[string]interface{}{
			"id":         du.ID,
			"name":       du.Name,
			"email":      du.Email,
			"status":     du.Status,
			"is_super":   du.IsSuper,
			"last_login": du.LastLogin,
		})
	}
	return c.ResponseOK(data)
}
