package user

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
)

type updateHandler struct {
	models *model.Models
}

func UpdateHandler(conf *config.ServerConfig) api.Handler {
	return &updateHandler{models: conf.Models}
}

func (h *updateHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, &api.AuthPerm{
		Scope:   types.ScopePlatform,
		ScopeId: 0,
		Role:    types.RoleEditor,
	}, nil
}

func (h *updateHandler) Handle(c *api.Context) *utils.Response {
	var body userBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}

	userObj, err := h.models.UserManager.GetByName(c.Param("username"))
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, err))
	}

	if body.Status != "" {
		userObj.Status = body.Status
	}

	if body.Password != "" {
		userObj.Password = utils.Encrypt(body.Password)
	}

	if body.Email != "" {
		if ok := utils.VerifyEmailFormat(body.Email); !ok {
			return c.ResponseError(errors.New(code.ParamsError, fmt.Sprintf("email:%s format error", body.Email)))
		}
		userObj.Email = body.Email
	}

	err = h.models.UserManager.Update(userObj)
	if err != nil {
		err = errors.New(code.UpdateError, err)
	}
	resp := c.ResponseError(err)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationUpdate,
		OperateDetail:        fmt.Sprintf("更新用户：%s", userObj.Name),
		Scope:                types.ScopePlatform,
		ResourceId:           userObj.ID,
		ResourceType:         types.AuditResourcePlatformUser,
		ResourceName:         userObj.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: body,
	})
	return resp
}
