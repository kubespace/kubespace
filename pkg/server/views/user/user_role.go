package user

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/manager/user"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/utils"
	"net/http"
	"strconv"
	"strings"
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
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	if ser.Scope == "" {
		return c.GenerateResponseError(errors.New(code.ParamsError, "scope参数错误"))
	}
	if ser.Scope == types.ScopePlatform {
		ser.ScopeId = 0
	}
	if ser.ScopeId == 0 && ser.Scope != types.ScopePlatform {
		return c.GenerateResponseError(errors.New(code.ParamsError, "scopeId参数错误"))
	}
	userRoles, err := r.models.UserRoleManager.List(ser.Scope, ser.ScopeId)
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DBError, "获取用户权限列表失败:"+err.Error()))
	}
	return c.GenerateResponseOK(userRoles)
}

func (r *UserRole) update(c *views.Context) *utils.Response {
	var ser serializers.UserRoleUpdateSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	if ser.Scope == "" {
		return c.GenerateResponseError(errors.New(code.ParamsError, "scope参数错误"))
	}
	if ser.Role == "" {
		return c.GenerateResponseError(errors.New(code.ParamsError, "role参数错误"))
	}
	if ser.Scope == types.ScopePlatform {
		ser.ScopeId = 0
	}
	if ser.ScopeId == 0 && ser.Scope != types.ScopePlatform {
		return c.GenerateResponseError(errors.New(code.ParamsError, "scopeId参数错误"))
	}
	if len(ser.UserIds) == 0 {
		return c.GenerateResponseError(errors.New(code.ParamsError, "用户列表为空"))
	}
	scopeName, err := getScopeName(r.models, ser.Scope, ser.ScopeId)
	if err != nil {
		return c.GenerateResponseError(err)
	}
	users, err := r.models.UserManager.List(user.UserListCondition{Ids: ser.UserIds})
	var userNames []string
	for _, u := range users {
		userNames = append(userNames, u.Name)
	}
	err = r.models.UserRoleManager.CreateOrUpdate(ser.Scope, ser.ScopeId, ser.UserIds, ser.Role)
	if err != nil {
		err = errors.New(code.DBError, err)
	}
	resp := c.GenerateResponse(err, nil)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationUpdate,
		OperateDetail:        fmt.Sprintf("更新用户「%s」权限为：%s", strings.Join(userNames, ","), ser.Role),
		Scope:                ser.Scope,
		ScopeId:              ser.ScopeId,
		ScopeName:            scopeName,
		ResourceType:         types.AuditResourcePermission,
		ResourceName:         strings.Join(userNames, ","),
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: ser,
	})
	return resp
}

func (r *UserRole) delete(c *views.Context) *utils.Response {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	userRole, err := r.models.UserRoleManager.GetById(uint(id))
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DataNotExists, err))
	}
	userObj, err := r.models.UserManager.GetById(userRole.UserId)
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DataNotExists, err))
	}
	scopeName, err := getScopeName(r.models, userRole.Scope, userRole.ScopeId)
	if err != nil {
		return c.GenerateResponseError(err)
	}
	err = r.models.UserRoleManager.Delete(uint(id))
	if err != nil {
		err = errors.New(code.DBError, err)
	}
	resp := c.GenerateResponse(err, nil)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationDelete,
		OperateDetail:        fmt.Sprintf("删除用户「%s」权限：%s", userObj.Name, userRole.Role),
		Scope:                userRole.Scope,
		ScopeId:              userRole.ScopeId,
		ScopeName:            scopeName,
		ResourceType:         types.AuditResourcePermission,
		ResourceId:           userRole.ID,
		ResourceName:         userObj.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: nil,
	})
	return &utils.Response{Code: code.Success}
}

func getScopeName(models *model.Models, scope string, scopeId uint) (string, error) {
	var scopeName string
	switch scope {
	case types.ScopeCluster:
		if clusterObj, err := models.ClusterManager.Get(scopeId); err != nil {
			return "", errors.New(code.DataNotExists, fmt.Sprintf("获取集群id=%d失败：%s", scopeId, err.Error()))
		} else {
			scopeName = clusterObj.Name1
		}
	case types.ScopeProject:
		if projectObj, err := models.ProjectManager.Get(scopeId); err != nil {
			return "", errors.New(code.DataNotExists, fmt.Sprintf("获取工作空间id=%d失败：%s", scopeId, err.Error()))
		} else {
			scopeName = projectObj.Name
		}
	case types.ScopePipeline:
		if pipespace, err := models.PipelineWorkspaceManager.Get(scopeId); err != nil {
			return "", errors.New(code.DataNotExists, fmt.Sprintf("获取流水线空间id=%d失败：%s", scopeId, err.Error()))
		} else {
			scopeName = pipespace.Name
		}
	}
	return scopeName, nil
}
