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
)

type watchHandler struct {
	models     *model.Models
	kubeClient *cluster.KubeClient
}

func WatchHandler(conf *config.ServerConfig) api.Handler {
	return &watchHandler{
		models:     conf.Models,
		kubeClient: conf.ServiceFactory.Cluster.KubeClient,
	}
}

func (h *watchHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	clusterId, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopeCluster,
		ScopeId: clusterId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *watchHandler) Handle(c *api.Context) *utils.Response {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	var ser resource.QueryParams
	if err := c.ShouldBindQuery(&ser); err != nil {
		c.SSEvent("message", err.Error())
		return nil
	}
	watchOuter, err := h.kubeClient.Watch(c.Param("id"), c.Param("resType"), &ser)
	if err != nil {
		c.SSEvent("message", err.Error())
		return nil
	}
	c.SSEvent("message", "{}")
	c.Writer.Flush()
	defer watchOuter.Close()
	for {
		select {
		case <-c.Writer.CloseNotify():
			klog.Info("select for cluster %s resource %s client gone", c.Param("id"), c.Param("resType"))
			return nil
		case event := <-watchOuter.OutCh():
			c.SSEvent("message", event)
			c.Writer.Flush()
		case <-watchOuter.StopCh():
			return nil
		}
	}
}
