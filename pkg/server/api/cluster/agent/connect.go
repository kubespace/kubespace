package agent

import (
	"github.com/gorilla/websocket"
	"github.com/kubespace/kubespace/pkg/informer"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
	"net/http"
)

type connectHandler struct {
	models          *model.Models
	informerFactory informer.Factory
}

func ConnectHandler(conf *config.ServerConfig) api.Handler {
	return &connectHandler{
		models:          conf.Models,
		informerFactory: conf.InformerFactory,
	}
}

func (h *connectHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	// 不需要用户认证
	return false, nil, nil
}

func (h *connectHandler) Handle(c *api.Context) *utils.Response {
	token := c.GetHeader("token")
	if token == "" {
		c.String(http.StatusBadRequest, "params error, not found cluster token")
	}

	clusterObj, err := h.models.ClusterManager.GetByToken(token)
	if err != nil {
		c.String(http.StatusForbidden, "get cluster error: %v", err)
		return nil
	}

	upGrader := &websocket.Upgrader{}
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.String(http.StatusBadRequest, "upgrade connection error: %v", err)
		return nil
	}

	tunnel, err := newAgentTunnel(ws, h.informerFactory.ClusterAgentInformer(token))
	if err != nil {
		klog.Errorf("create agent tunnel error: %v", err)
		c.String(http.StatusInternalServerError, "create agent tunnel error: %v", err)
		return nil
	}
	tunnel.Consume()
	clusterObj.Status = types.ClusterConnect
	if err = h.models.ClusterManager.UpdateByObject(clusterObj.ID, &types.Cluster{Status: types.ClusterConnect}); err != nil {
		klog.Warningf("update cluster status error: %s", err.Error())
	}
	klog.V(1).Infof("cluster id=%s name=%s kube connect finish", clusterObj.Name, clusterObj.Name1)
	return nil
}
