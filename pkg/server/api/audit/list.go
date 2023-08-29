package audit

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/manager/audit"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
)

type listHandler struct {
	models *model.Models
}

func ListHandler(conf *config.ServerConfig) api.Handler {
	return &listHandler{models: conf.Models}
}

func (h *listHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	var cond audit.AuditOperateListCondition
	if err := c.ShouldBindQuery(&cond); err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   cond.Scope,
		ScopeId: cond.ScopeId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *listHandler) Handle(c *api.Context) *utils.Response {
	var listCond audit.AuditOperateListCondition
	if err := c.ShouldBindQuery(&listCond); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	aos, page, err := h.models.AuditOperateManager.List(&listCond)
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}

	return c.ResponseOK(map[string]interface{}{
		"data":       aos,
		"pagination": page,
	})
}
