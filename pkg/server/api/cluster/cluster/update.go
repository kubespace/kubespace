package cluster

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
)

type updateHandler struct {
	models *model.Models
}

type updateClusterBody struct {
	KubeConfig string `json:"kubeconfig"`
}

func UpdateHandler(conf *config.ServerConfig) api.Handler {
	return &updateHandler{models: conf.Models}
}

func (h *updateHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	clusterId, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	// 更新集群需要有集群编辑权限
	return true, &api.AuthPerm{
		Scope:   types.ScopeCluster,
		ScopeId: clusterId,
		Role:    types.RoleEditor,
	}, nil
}

func (h *updateHandler) Handle(c *api.Context) *utils.Response {
	var ser updateClusterBody

	if err := c.ShouldBind(&ser); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}

	clusterId, _ := utils.ParseUint(c.Param("id"))
	clusterObj, err := h.models.ClusterManager.GetById(clusterId)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, fmt.Sprintf("not found cluster id=%d", clusterId)))
	}

	err = h.models.ClusterManager.UpdateByObject(clusterId, &types.Cluster{KubeConfig: ser.KubeConfig})
	if err != nil {
		err = errors.New(code.DBError, err)
	}
	resp := c.Response(err, nil)

	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationUpdate,
		OperateDetail:        "更新集群：" + clusterObj.Name1,
		Scope:                types.ScopeCluster,
		ScopeId:              clusterObj.ID,
		ScopeName:            clusterObj.Name1,
		ResourceId:           clusterObj.ID,
		ResourceType:         types.AuditResourceCluster,
		ResourceName:         clusterObj.Name1,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: ser,
	})
	return resp
}
