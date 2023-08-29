package resource

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/kubernetes/resource"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/service/cluster"
	"github.com/kubespace/kubespace/pkg/utils"
)

type patchHandler struct {
	models     *model.Models
	kubeClient *cluster.KubeClient
}

func PatchHandler(conf *config.ServerConfig) api.Handler {
	return &patchHandler{
		models:     conf.Models,
		kubeClient: conf.ServiceFactory.Cluster.KubeClient,
	}
}

func (h *patchHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	clusterId, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopeCluster,
		ScopeId: clusterId,
		Role:    types.RoleEditor,
	}, nil
}

func (h *patchHandler) Handle(c *api.Context) *utils.Response {
	var ser resource.PatchParams
	if err := c.ShouldBind(&ser); err != nil {
		return c.ResponseError(errors.New(code.ParseError, err))
	}

	scope, scopeName, scopeId, err := GetAuditScope(h.models, c.Query("project_id"), c.Param("id"))
	if err != nil {
		return c.ResponseError(err)
	}

	resp := h.kubeClient.Patch(c.Param("id"), c.Param("resType"), &ser)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationPatch,
		OperateDetail:        fmt.Sprintf("Patch %s %s/%s", c.Param("resType"), ser.Namespace, ser.Name),
		Scope:                scope,
		ScopeId:              scopeId,
		ScopeName:            scopeName,
		Namespace:            ser.Namespace,
		ResourceType:         c.Param("resType"),
		ResourceName:         ser.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: &ser,
	})
	return resp
}
