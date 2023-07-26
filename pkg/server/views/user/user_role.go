package user

import (
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"net/http"
	"strconv"
)

type UserRole struct {
	Views  []*views.View
	models *model.Models
}

func NewUserRole(models *model.Models) *UserRole {
	role := &UserRole{
		models: models,
	}
	role.Views = []*views.View{
		views.NewView(http.MethodGet, "", role.list),
		views.NewView(http.MethodPost, "", role.update),
		views.NewView(http.MethodDelete, "/:id", role.delete),
	}
	role.models.RoleManager.Init()
	return role
}

func (r *UserRole) list(c *views.Context) *utils.Response {
	var ser serializers.UserRoleSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	if ser.Scope == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "scope参数错误"}
	}
	if ser.Scope == types.RoleScopePlatform {
		ser.ScopeId = 0
	}
	if ser.ScopeId == 0 && ser.Scope != types.RoleScopePlatform {
		return &utils.Response{Code: code.ParamsError, Msg: "scopeId参数错误"}
	}
	userRoles, err := r.models.UserRoleManager.List(ser.Scope, ser.ScopeId)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: "获取用户权限列表失败:" + err.Error()}
	}
	return &utils.Response{Code: code.Success, Data: userRoles}
}

func (r *UserRole) update(c *views.Context) *utils.Response {
	var ser serializers.UserRoleUpdateSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	if ser.Scope == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "scope参数错误"}
	}
	if ser.Role == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "role参数错误"}
	}
	if ser.Scope == types.RoleScopePlatform {
		ser.ScopeId = 0
	}
	if ser.ScopeId == 0 && ser.Scope != types.RoleScopePlatform {
		return &utils.Response{Code: code.ParamsError, Msg: "scopeId参数错误"}
	}
	if len(ser.UserIds) == 0 {
		return &utils.Response{Code: code.ParamsError, Msg: "用户列表为空"}
	}
	if err := r.models.UserRoleManager.CreateOrUpdate(ser.Scope, ser.ScopeId, ser.UserIds, ser.Role); err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success, Msg: "创建用户权限成功"}
}

func (r *UserRole) delete(c *views.Context) *utils.Response {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	if err = r.models.UserRoleManager.Delete(uint(id)); err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success}
}
