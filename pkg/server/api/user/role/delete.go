package role

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
	"strconv"
)

type deleteHandler struct {
	models *model.Models
}

func DeleteHandler(conf *config.ServerConfig) api.Handler {
	return &deleteHandler{models: conf.Models}
}

func (h *deleteHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	id, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	userRole, err := h.models.UserRoleManager.GetById(id)
	if err != nil {
		return true, nil, errors.New(code.DataNotExists, err)
	}
	return true, &api.AuthPerm{
		Scope:   userRole.Scope,
		ScopeId: userRole.ScopeId,
		Role:    types.RoleAdmin,
	}, nil
}

func (h *deleteHandler) Handle(c *api.Context) *utils.Response {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	userRole, err := h.models.UserRoleManager.GetById(uint(id))
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, err))
	}
	userObj, err := h.models.UserManager.GetById(userRole.UserId)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, err))
	}
	scopeName, err := getScopeName(h.models, userRole.Scope, userRole.ScopeId)
	if err != nil {
		return c.ResponseError(err)
	}
	err = h.models.UserRoleManager.Delete(uint(id))
	if err != nil {
		err = errors.New(code.DBError, err)
	}
	resp := c.ResponseError(err)
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
	return resp
}
