package image_registry

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
	objs, err := h.models.ImageRegistryManager.List()
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}
	var data []map[string]interface{}

	for _, obj := range objs {
		data = append(data, map[string]interface{}{
			"id":          obj.ID,
			"registry":    obj.Registry,
			"user":        obj.User,
			"create_user": obj.CreateUser,
			"update_user": obj.UpdateUser,
			"create_time": obj.CreateTime,
			"update_time": obj.UpdateTime,
		})
	}
	return c.ResponseOK(data)
}
