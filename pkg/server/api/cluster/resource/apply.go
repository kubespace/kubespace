package resource

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/kubernetes/resource"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/service/cluster"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
	"strings"
)

type applyHandler struct {
	models     *model.Models
	kubeClient *cluster.KubeClient
}

func ApplyHandler(conf *config.ServerConfig) api.Handler {
	return &applyHandler{
		models:     conf.Models,
		kubeClient: conf.ServiceFactory.Cluster.KubeClient,
	}
}

func (h *applyHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
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

func (h *applyHandler) Handle(c *api.Context) *utils.Response {
	var ser resource.ApplyParams
	if err := c.ShouldBind(&ser); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}

	scope, scopeName, scopeId, err := GetAuditScope(h.models, c.Query("project_id"), c.Param("id"))
	if err != nil {
		return c.ResponseError(err)
	}

	resp := h.kubeClient.Apply(c.Param("id"), &ser)

	var applyResources []*resource.ApplyResource
	if err := utils.ConvertTypeByJson(resp.Data, &applyResources); err != nil {
		klog.Errorf("convert apply resource error: %s", err.Error())
	}
	if len(applyResources) > 0 {
		var resNames []string
		var resNamesWithKind []string
		namespaceMap := make(map[string]struct{})
		resTypeMap := make(map[string]struct{})
		for _, r := range applyResources {
			if r.Namespace != "" {
				namespaceMap[r.Namespace] = struct{}{}
			}
			if r.Kind != "" {
				resTypeMap[strings.ToLower(r.Kind)] = struct{}{}
			}
			resNames = append(resNames, r.Name)
			resNamesWithKind = append(resNamesWithKind, r.Kind+"/"+r.Name)
		}
		var namespaces []string
		var resTypes []string
		for n := range namespaceMap {
			namespaces = append(namespaces, n)
		}
		for t := range resTypeMap {
			resTypes = append(resTypes, t)
		}
		operation := types.AuditOperationApply
		opDetail := "Apply resources: " + strings.Join(resNamesWithKind, ",")
		if ser.Create {
			operation = types.AuditOperationCreate
			opDetail = "创建资源：" + strings.Join(resNamesWithKind, ",")
		}
		c.CreateAudit(&types.AuditOperate{
			Operation:            operation,
			OperateDetail:        opDetail,
			Scope:                scope,
			ScopeId:              scopeId,
			ScopeName:            scopeName,
			Namespace:            strings.Join(namespaces, ","),
			ResourceType:         strings.Join(resTypes, ","),
			ResourceName:         strings.Join(resNames, ","),
			Code:                 resp.Code,
			Message:              resp.Msg,
			OperateDataInterface: &ser,
		})
	}
	return resp
}
