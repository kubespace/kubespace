package cluster

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
	"time"
)

type createHandler struct {
	models *model.Models
}

type createClusterBody struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

func CreateHandler(conf *config.ServerConfig) api.Handler {
	return &createHandler{models: conf.Models}
}

func (h *createHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	// 创建集群需要有平台编辑权限
	return true, &api.AuthPerm{
		Scope:   types.ScopePlatform,
		ScopeId: 0,
		Role:    types.RoleEditor,
	}, nil
}

func (h *createHandler) Handle(c *api.Context) *utils.Response {
	var body createClusterBody
	if err := c.ShouldBind(&body); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	if body.Name == "" {
		return c.ResponseError(errors.New(code.ParamsError, "cluster name is blank"))
	}
	clusterObj := &types.Cluster{
		Name1:      body.Name,
		Token:      utils.ShortUUID(),
		Status:     types.ClusterPending,
		CreatedBy:  c.User.Name,
		Members:    body.Members,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	err := h.models.ClusterManager.Create(clusterObj)
	if err != nil {
		err = errors.New(code.CreateError, err)
	}
	resp := c.Response(err, clusterObj)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationCreate,
		OperateDetail:        "创建集群：" + clusterObj.Name1,
		Scope:                types.ScopeCluster,
		ScopeId:              clusterObj.ID,
		ScopeName:            clusterObj.Name1,
		ResourceId:           clusterObj.ID,
		ResourceType:         types.AuditResourceCluster,
		ResourceName:         clusterObj.Name1,
		Code:                 code.Success,
		OperateDataInterface: clusterObj,
	})
	return resp
}
