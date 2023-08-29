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
	var ser createImageRegistryBody
	if err := c.ShouldBind(&ser); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	id, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	obj, err := h.models.ImageRegistryManager.Get(id)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, "获取镜像仓库失败: "+err.Error()))
	}
	obj.User = ser.User
	obj.Password = ser.Password
	obj.UpdateUser = c.User.Name
	obj.UpdateTime = time.Now()
	_, err = h.models.ImageRegistryManager.Update(obj)
	if err != nil {
		err = errors.New(code.DBError, "更新镜像仓库失败: "+err.Error())
	}
	resp := c.Response(err, nil)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationUpdate,
		OperateDetail:        fmt.Sprintf("更新镜像仓库：%s", ser.Registry),
		Scope:                types.ScopePlatform,
		ResourceId:           obj.ID,
		ResourceType:         types.AuditResourcePlatformRegistry,
		ResourceName:         obj.Registry,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: ser,
	})
	return resp
}
