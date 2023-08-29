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
	"strconv"
)

type deleteHandler struct {
	models *model.Models
}

func DeleteHandler(conf *config.ServerConfig) api.Handler {
	return &deleteHandler{models: conf.Models}
}

func (h *deleteHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, &api.AuthPerm{
		Scope:   types.ScopePlatform,
		ScopeId: 0,
		Role:    types.RoleEditor,
	}, nil
}

func (h *deleteHandler) Handle(c *api.Context) *utils.Response {
	secretId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	obj, err := h.models.ImageRegistryManager.Get(uint(secretId))
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, "获取镜像仓库失败: "+err.Error()))
	}
	err = h.models.ImageRegistryManager.Delete(obj)
	if err != nil {
		err = errors.New(code.DBError, "删除镜像仓库失败: "+err.Error())
	}
	resp := c.ResponseError(err)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationDelete,
		OperateDetail:        fmt.Sprintf("删除镜像仓库：%s", obj.Registry),
		Scope:                types.ScopePlatform,
		ResourceId:           obj.ID,
		ResourceType:         types.AuditResourcePlatformRegistry,
		ResourceName:         obj.Registry,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: nil,
	})
	return resp
}
