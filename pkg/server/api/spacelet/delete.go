package spacelet

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	spaceletservice "github.com/kubespace/kubespace/pkg/service/spacelet"
	"github.com/kubespace/kubespace/pkg/utils"
)

type deleteHandler struct {
	models          *model.Models
	spaceletService *spaceletservice.SpaceletService
}

func DeleteHandler(conf *config.ServerConfig) api.Handler {
	return &deleteHandler{
		models:          conf.Models,
		spaceletService: conf.ServiceFactory.Pipeline.SpaceletService,
	}
}

func (h *deleteHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, &api.AuthPerm{
		Scope:   types.ScopePlatform,
		ScopeId: 0,
		Role:    types.RoleEditor,
	}, nil
}

func (h *deleteHandler) Handle(c *api.Context) *utils.Response {
	id, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	spaceletObj, err := h.models.SpaceletManager.Get(id)
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}
	if spaceletObj.Status == types.SpaceletStatusOnline {
		return c.ResponseError(errors.New(code.StatusError, "当前Spacelet节点在线，不能删除"))
	}
	if err = h.models.SpaceletManager.Delete(id); err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}
	resp := c.ResponseOK(nil)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationDelete,
		OperateDetail:        fmt.Sprintf("删除Spacelet节点：%s", spaceletObj.Hostname),
		Scope:                types.ScopePlatform,
		ResourceId:           spaceletObj.ID,
		ResourceType:         types.AuditResourcePlatformSpacelet,
		ResourceName:         spaceletObj.Hostname,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: nil,
	})
	return resp
}
