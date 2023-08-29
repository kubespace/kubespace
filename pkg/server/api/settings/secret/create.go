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

type createHandler struct {
	models *model.Models
}

func CreateHandler(conf *config.ServerConfig) api.Handler {
	return &createHandler{models: conf.Models}
}

type createSecretBody struct {
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
	SecretType  string `json:"secret_type" form:"secret_type"`
	User        string `json:"user" form:"user"`
	Password    string `json:"password" form:"password"`
	PrivateKey  string `json:"private_key" form:"private_key"`
	AccessToken string `json:"access_token" form:"access_token"`
}

func (h *createHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, &api.AuthPerm{
		Scope:   types.ScopePlatform,
		ScopeId: 0,
		Role:    types.RoleEditor,
	}, nil
}

func (h *createHandler) Handle(c *api.Context) *utils.Response {
	var body createSecretBody
	if err := c.ShouldBind(&body); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	secret := &types.SettingsSecret{
		Name:        body.Name,
		Description: body.Description,
		Type:        body.SecretType,
		User:        body.User,
		Password:    body.Password,
		PrivateKey:  body.PrivateKey,
		AccessToken: body.AccessToken,
		CreateUser:  c.User.Name,
		UpdateUser:  c.User.Name,
		CreateTime:  time.Time{},
		UpdateTime:  time.Time{},
	}
	_, err := h.models.SettingsSecretManager.Create(secret)
	if err != nil {
		err = errors.New(code.DBError, err)
	}
	resp := c.ResponseError(err)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationCreate,
		OperateDetail:        fmt.Sprintf("创建密钥：%s", body.Name),
		Scope:                types.ScopePlatform,
		ResourceId:           secret.ID,
		ResourceType:         types.AuditResourcePlatformSecret,
		ResourceName:         body.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: body,
	})
	return resp
}
