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
	"k8s.io/klog/v2"
	"strconv"
)

type deleteHandler struct {
	models *model.Models
}

func DeleteHandler(conf *config.ServerConfig) api.Handler {
	return &deleteHandler{models: conf.Models}
}

func (h *deleteHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	clusterId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	// 删除集群需要有集群管理员权限
	return true, &api.AuthPerm{
		Scope:   types.ScopeCluster,
		ScopeId: uint(clusterId),
		Role:    types.RoleAdmin,
	}, nil
}

func (h *deleteHandler) Handle(c *api.Context) *utils.Response {
	id, _ := utils.ParseUint(c.Param("id"))
	clusterObj, err := h.models.ClusterManager.GetById(id)
	if err != nil {
		return &utils.Response{Code: code.DataNotExists, Msg: fmt.Sprintf("not found cluster id=%d", id)}
	}
	err = h.models.ClusterManager.Delete(id)
	if err != nil {
		klog.Errorf("delete cluster %s error: %s", c, err.Error())
		err = errors.New(code.DBError, err)
	}
	resp := c.Response(err, nil)

	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationDelete,
		OperateDetail:        "删除集群：" + clusterObj.Name1,
		Scope:                types.ScopeCluster,
		ScopeId:              clusterObj.ID,
		ScopeName:            clusterObj.Name1,
		ResourceId:           clusterObj.ID,
		ResourceType:         types.AuditResourceCluster,
		ResourceName:         clusterObj.Name1,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: clusterObj,
	})
	return resp
}
