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

type deleteHandler struct {
	models *model.Models
}

func DeleteHandler(conf *config.ServerConfig) api.Handler {
	return &deleteHandler{models: conf.Models}
}

type deleteUserBody struct {
	Name string `json:"name"`
}

func (h *deleteHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, &api.AuthPerm{
		Scope:   types.ScopePlatform,
		ScopeId: 0,
		Role:    types.RoleEditor,
	}, nil
}

func (h *deleteHandler) Handle(c *api.Context) *utils.Response {
	var ser []deleteUserBody
	if err := c.ShouldBind(&ser); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	for _, du := range ser {
		userObj, err := h.models.UserManager.GetByName(du.Name)
		if err != nil {
			return c.ResponseError(errors.New(code.DataNotExists, err))
		}
		err = h.models.UserManager.Delete(du.Name)
		if err != nil {
			err = errors.New(code.DeleteError, err)
		}
		resp := c.ResponseError(err)
		c.CreateAudit(&types.AuditOperate{
			Operation:            types.AuditOperationDelete,
			OperateDetail:        fmt.Sprintf("删除用户：%s", userObj.Name),
			Scope:                types.ScopePlatform,
			ResourceId:           userObj.ID,
			ResourceType:         types.AuditResourcePlatformUser,
			ResourceName:         userObj.Name,
			Code:                 resp.Code,
			Message:              resp.Msg,
			OperateDataInterface: userObj,
		})
		if err != nil {
			return resp
		}
	}
	return c.ResponseOK(nil)
}
