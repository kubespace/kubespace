package role

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/manager/user"
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

type listUserRoleForm struct {
	UserId  uint   `json:"user_id" form:"user_id"`
	Scope   string `json:"scope" form:"scope"`
	ScopeId uint   `json:"scope_id" form:"scope_id"`
	Role    string `json:"role" form:"from"`
}

func (h *listHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	var ser listUserRoleForm
	if err := c.ShouldBind(&ser); err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   ser.Scope,
		ScopeId: ser.ScopeId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *listHandler) Handle(c *api.Context) *utils.Response {
	var ser listUserRoleForm
	if err := c.ShouldBind(&ser); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	if ser.Scope == "" {
		return c.ResponseError(errors.New(code.ParamsError, "scope参数错误"))
	}
	if ser.Scope == types.ScopePlatform {
		ser.ScopeId = 0
	}
	if ser.ScopeId == 0 && ser.Scope != types.ScopePlatform {
		return c.ResponseError(errors.New(code.ParamsError, "scopeId参数错误"))
	}
	userRoles, err := h.models.UserRoleManager.List(&user.ListUserRoleCondition{
		Scope:        ser.Scope,
		ScopeId:      &ser.ScopeId,
		WithUsername: true,
	})
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, "获取用户权限列表失败:"+err.Error()))
	}
	return c.ResponseOK(userRoles)
}
