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

type updateHandler struct {
	models     *model.Models
	kubeClient *cluster.KubeClient
}

func UpdateHandler(conf *config.ServerConfig) api.Handler {
	return &updateHandler{
		models:     conf.Models,
		kubeClient: conf.ServiceFactory.Cluster.KubeClient,
	}
}

func (h *updateHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
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

func (h *updateHandler) Handle(c *api.Context) *utils.Response {
	var params resource.UpdateParams
	if err := c.ShouldBind(&params); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	params.Namespace = c.Param("namespace")
	params.Name = c.Param("name")

	scope, scopeName, scopeId, err := GetAuditScope(h.models, c.Query("project_id"), c.Param("id"))
	if err != nil {
		return c.ResponseError(err)
	}

	resp := h.kubeClient.Update(c.Param("id"), c.Param("resType"), &params)

	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationUpdate,
		OperateDetail:        "更新" + c.Param("resType") + ":" + params.Name,
		Scope:                scope,
		ScopeId:              scopeId,
		ScopeName:            scopeName,
		Namespace:            params.Namespace,
		ResourceType:         c.Param("resType"),
		ResourceName:         params.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: &params,
	})
	return resp
}

func GetAuditScope(models *model.Models, projectId, clusterId string) (scope, scopeName string, scopeId uint, err error) {
	if projectId != "" {
		id, _ := utils.ParseUint(projectId)
		if id == 0 {
			err = errors.New(code.ParseError, "project_id参数错误")
			return
		}
		project, getErr := models.ProjectManager.Get(id)
		if getErr != nil {
			err = errors.New(code.DataNotExists, fmt.Sprintf("获取工作空间id=%d错误：%v", id, err))
			return
		}
		scope = types.ScopeProject
		scopeId = project.ID
		scopeName = project.Name
	} else {
		id, _ := utils.ParseUint(clusterId)
		clusterObj, getErr := models.ClusterManager.GetById(id)
		if getErr != nil {
			err = errors.New(code.DataNotExists, fmt.Sprintf("获取集群id=%d错误：%v", id, err))
			return
		}
		if clusterObj == nil {
			err = errors.New(code.DataNotExists, fmt.Sprintf("未找到集群id=%s", clusterId))
			return
		}
		scope = types.ScopeCluster
		scopeId = clusterObj.ID
		scopeName = clusterObj.Name1
	}
	return
}
