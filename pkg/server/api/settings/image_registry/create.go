package image_registry

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

type createImageRegistryBody struct {
	Registry string `json:"registry" form:"registry"`
	User     string `json:"user" form:"user"`
	Password string `json:"password" form:"password"`
}

func (h *createHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, &api.AuthPerm{
		Scope:   types.ScopePlatform,
		ScopeId: 0,
		Role:    types.RoleEditor,
	}, nil
}

func (h *createHandler) Handle(c *api.Context) *utils.Response {
	var body createImageRegistryBody
	if err := c.ShouldBind(&body); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	registry := &types.SettingsImageRegistry{
		Registry:   body.Registry,
		User:       body.User,
		Password:   body.Password,
		CreateUser: c.User.Name,
		UpdateUser: c.User.Name,
		CreateTime: time.Time{},
		UpdateTime: time.Time{},
	}
	_, err := h.models.ImageRegistryManager.Create(registry)
	if err != nil {
		err = errors.New(code.DBError, "创建镜像仓库失败: "+err.Error())
	}
	resp := c.ResponseError(err)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationCreate,
		OperateDetail:        fmt.Sprintf("创建镜像仓库：%s", body.Registry),
		Scope:                types.ScopePlatform,
		ResourceId:           registry.ID,
		ResourceType:         types.AuditResourcePlatformRegistry,
		ResourceName:         body.Registry,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: body,
	})
	return resp
}
