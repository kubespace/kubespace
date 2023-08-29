package role

import (
	"fmt"
	"github.com/gin-gonic/gin/binding"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/manager/user"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
	"strings"
)

type updateHandler struct {
	models *model.Models
}

func UpdateHandler(conf *config.ServerConfig) api.Handler {
	return &updateHandler{models: conf.Models}
}

type updateUserRoleBody struct {
	UserIds []uint `json:"user_ids" form:"user_ids"`
	Scope   string `json:"scope" form:"scope"`
	ScopeId uint   `json:"scope_id" form:"scope_id"`
	Role    string `json:"role" form:"from"`
}

func (h *updateHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	var ser updateUserRoleBody
	if err := c.ShouldBindBodyWith(&ser, binding.JSON); err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   ser.Scope,
		ScopeId: ser.ScopeId,
		Role:    types.RoleAdmin,
	}, nil
}

func (h *updateHandler) Handle(c *api.Context) *utils.Response {
	var body updateUserRoleBody
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	if body.Scope == "" {
		return c.ResponseError(errors.New(code.ParamsError, "scope参数错误"))
	}
	if body.Role == "" {
		return c.ResponseError(errors.New(code.ParamsError, "role参数错误"))
	}
	if body.Scope == types.ScopePlatform {
		body.ScopeId = 0
	}
	if body.ScopeId == 0 && body.Scope != types.ScopePlatform {
		return c.ResponseError(errors.New(code.ParamsError, "scopeId参数错误"))
	}
	if len(body.UserIds) == 0 {
		return c.ResponseError(errors.New(code.ParamsError, "用户列表为空"))
	}
	scopeName, err := getScopeName(h.models, body.Scope, body.ScopeId)
	if err != nil {
		return c.ResponseError(err)
	}
	users, err := h.models.UserManager.List(user.UserListCondition{Ids: body.UserIds})
	var userNames []string
	for _, u := range users {
		userNames = append(userNames, u.Name)
	}
	err = h.models.UserRoleManager.CreateOrUpdate(body.Scope, body.ScopeId, body.UserIds, body.Role)
	if err != nil {
		err = errors.New(code.DBError, err)
	}
	resp := c.Response(err, nil)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationUpdate,
		OperateDetail:        fmt.Sprintf("更新用户「%s」权限为：%s", strings.Join(userNames, ","), body.Role),
		Scope:                body.Scope,
		ScopeId:              body.ScopeId,
		ScopeName:            scopeName,
		ResourceType:         types.AuditResourcePermission,
		ResourceName:         strings.Join(userNames, ","),
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: body,
	})
	return resp
}

func getScopeName(models *model.Models, scope string, scopeId uint) (string, error) {
	var scopeName string
	switch scope {
	case types.ScopeCluster:
		if clusterObj, err := models.ClusterManager.GetById(scopeId); err != nil {
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
