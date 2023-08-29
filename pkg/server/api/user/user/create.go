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
	"time"
)

type createHandler struct {
	models *model.Models
}

func CreateHandler(conf *config.ServerConfig) api.Handler {
	return &createHandler{models: conf.Models}
}

type userBody struct {
	Name     string   `json:"name"`
	Password string   `json:"password"`
	Email    string   `json:"email"`
	Status   string   `json:"status"`
	Roles    []string `json:"roles"`
}

func (h *createHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, &api.AuthPerm{
		Scope:   types.ScopePlatform,
		ScopeId: 0,
		Role:    types.RoleEditor,
	}, nil
}

func (h *createHandler) Handle(c *api.Context) *utils.Response {
	var ser userBody

	if err := c.ShouldBind(&ser); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	if ser.Name == "" {
		return c.ResponseError(errors.New(code.ParamsError, "params error, user name is empty"))
	}
	if ok := utils.VerifyEmailFormat(ser.Email); !ok {
		return c.ResponseError(errors.New(code.ParamsError, fmt.Sprintf("email:%s format error", ser.Email)))
	}

	userObj := types.User{
		Name:       ser.Name,
		Password:   utils.Encrypt(ser.Password),
		Email:      ser.Email,
		IsSuper:    false,
		Status:     "normal",
		LastLogin:  time.Now(),
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	err := h.models.UserManager.Create(&userObj)
	if err != nil {
		err = errors.New(code.CreateError, err)
	}
	resp := c.ResponseError(err)

	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationCreate,
		OperateDetail:        fmt.Sprintf("创建用户：%s", ser.Name),
		Scope:                types.ScopePlatform,
		ResourceId:           userObj.ID,
		ResourceType:         types.AuditResourcePlatformUser,
		ResourceName:         ser.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: ser,
	})
	return resp
}
