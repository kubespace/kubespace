package secret

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
	var body createSecretBody
	if err := c.ShouldBind(&body); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	secretId, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	secret, err := h.models.SettingsSecretManager.Get(secretId)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, "获取密钥失败: "+err.Error()))
	}
	secret.Description = body.Description
	secret.Type = body.SecretType
	secret.User = body.User
	secret.Password = body.Password
	secret.PrivateKey = body.PrivateKey
	secret.AccessToken = body.AccessToken
	secret.UpdateUser = c.User.Name
	secret.UpdateTime = time.Now()
	_, err = h.models.SettingsSecretManager.Update(secret)
	if err != nil {
		err = errors.New(code.DBError, "更新密钥失败: "+err.Error())
	}
	resp := c.ResponseError(err)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationUpdate,
		OperateDetail:        fmt.Sprintf("更新密钥：%s", secret.Name),
		Scope:                types.ScopePlatform,
		ResourceId:           secret.ID,
		ResourceType:         types.AuditResourcePlatformSecret,
		ResourceName:         secret.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: body,
	})
	return resp
}
