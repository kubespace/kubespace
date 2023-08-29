package cluster

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	kubetypes "github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/model"
	clustermgr "github.com/kubespace/kubespace/pkg/model/manager/cluster"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/service/cluster"
	"github.com/kubespace/kubespace/pkg/utils"
	"sync"
)

type listHandler struct {
	models     *model.Models
	kubeClient *cluster.KubeClient
}

type listClusterObject struct {
	*types.Cluster `json:",inline"`
	Version        string `json:"version"`
	ConnectError   string `json:"connect_error"`
}

func ListHandler(conf *config.ServerConfig) api.Handler {
	return &listHandler{
		models:     conf.Models,
		kubeClient: conf.ServiceFactory.Cluster.KubeClient,
	}
}

func (h *listHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, nil, nil
}

func (h *listHandler) Handle(c *api.Context) *utils.Response {
	clusters, err := h.models.ClusterManager.List(clustermgr.ListClusterCondition{})
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}
	var data []*listClusterObject

	var wg sync.WaitGroup
	for i, du := range clusters {
		// 用户是否有该集群view权限
		if !h.models.UserRoleManager.AuthRole(c.User, types.ScopeCluster, du.ID, types.RoleViewer) {
			continue
		}
		cluObject := &listClusterObject{Cluster: clusters[i]}
		data = append(data, cluObject)

		// 并发获取集群是否联通
		wg.Add(1)
		go func(clu *listClusterObject) {
			defer wg.Done()
			clu.Status = types.ClusterPending

			res := h.kubeClient.Get(clu.Name, kubetypes.ClusterType, map[string]interface{}{"only_version": true})
			if res.IsSuccess() {
				// 集群联通，获取集群版本
				clu.Status = types.ClusterConnect
				clu.Version, _ = res.Data.(string)
				return
			}
			// 连接错误信息
			clu.ConnectError = res.Msg
			// 如果是配置kubeconfig，则状态为连接失败，否则为待连接状态
			if clu.KubeConfig != "" {
				clu.Status = types.ClusterFailed
			}
		}(cluObject)
	}
	wg.Wait()
	return c.ResponseOK(data)
}
