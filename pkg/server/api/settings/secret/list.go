package secret

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
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
	return true, &api.AuthPerm{
		Scope:   types.ScopePlatform,
		ScopeId: 0,
		Role:    types.RoleViewer,
	}, nil
}

func (h *listHandler) Handle(c *api.Context) *utils.Response {
	secrets, err := h.models.SettingsSecretManager.List()
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	var data []map[string]interface{}

	for _, secret := range secrets {
		data = append(data, map[string]interface{}{
			"id":          secret.ID,
			"name":        secret.Name,
			"description": secret.Description,
			"type":        secret.Type,
			"user":        secret.User,
			"create_user": secret.CreateUser,
			"update_user": secret.UpdateUser,
			"create_time": secret.CreateTime,
			"update_time": secret.UpdateTime,
		})
	}
	return c.ResponseOK(data)
}
