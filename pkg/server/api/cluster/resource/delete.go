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
	"strings"
)

type deleteHandler struct {
	models     *model.Models
	kubeClient *cluster.KubeClient
}

func DeleteHandler(conf *config.ServerConfig) api.Handler {
	return &deleteHandler{
		models:     conf.Models,
		kubeClient: conf.ServiceFactory.Cluster.KubeClient,
	}
}

func (h *deleteHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	if c.Query("project_id") != "" {
		projectId, err := utils.ParseUint(c.Query("project_id"))
		if err != nil {
			return true, nil, errors.New(code.ParamsError, err)
		}
		return true, &api.AuthPerm{
			Scope:   types.ScopeProject,
			ScopeId: projectId,
			Role:    types.RoleEditor,
		}, nil
	}
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

func (h *deleteHandler) Handle(c *api.Context) *utils.Response {
	var params resource.DeleteParams
	if err := c.ShouldBind(&params); err != nil {
		return c.ResponseError(errors.New(code.ParseError, err))
	}

	scope, scopeName, scopeId, err := GetAuditScope(h.models, c.Query("project_id"), c.Param("id"))
	if err != nil {
		return c.ResponseError(err)
	}

	resp := h.kubeClient.Delete(c.Param("id"), c.Param("resType"), &params)

	namespace := ""
	namespaceMap := make(map[string]struct{})
	var delNames []string
	var delNamesWithNamespace []string
	var opDetail = ""
	var delNameStr string
	if len(params.Resources) > 0 {
		for _, r := range params.Resources {
			if r.Namespace != "" {
				namespaceMap[r.Namespace] = struct{}{}
			}
			delNames = append(delNames, r.Name)
			delNamesWithNamespace = append(delNamesWithNamespace, r.Namespace+"/"+r.Name)
		}
		var namespaces []string
		for n := range namespaceMap {
			namespaces = append(namespaces, n)
		}
		namespace = strings.Join(namespaces, ",")
		if len(namespaces) > 1 {
			delNameStr = strings.Join(delNamesWithNamespace, ",")
		} else {
			delNameStr = strings.Join(delNames, ",")
		}
		opDetail = fmt.Sprintf("删除%s:%s", c.Param("resType"), delNameStr)
	} else {
		namespace = params.Namespace
		opDetail = fmt.Sprintf("删除%s:%s", c.Param("resType"), params.LabelSelector.String())
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationDelete,
		OperateDetail:        opDetail,
		Scope:                scope,
		ScopeId:              scopeId,
		ScopeName:            scopeName,
		Namespace:            namespace,
		ResourceType:         c.Param("resType"),
		ResourceName:         delNameStr,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: &params,
	})
	return resp
}
