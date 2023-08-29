package agent

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/kubespace/kubespace/pkg/informer"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/cluster"
	kubetypes "github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
	"net/http"
)

// 处理agent发送的请求返回
type responseHandler struct {
	models          *model.Models
	informerFactory informer.Factory
}

func ResponseHandler(conf *config.ServerConfig) api.Handler {
	return &responseHandler{
		models:          conf.Models,
		informerFactory: conf.InformerFactory,
	}
}

func (h *responseHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	// 不需要用户认证
	return false, nil, nil
}

func (h *responseHandler) Handle(c *api.Context) *utils.Response {
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

	go func() {
		defer ws.Close()
		for {
			_, data, err := ws.ReadMessage()
			if err != nil {
				break
			}
			var resp kubetypes.Response
			if err = json.Unmarshal(data, &resp); err != nil {
				klog.Errorf("json unmarshal response error: %s", err.Error())
				return
			}
			agentListWatcher := cluster.NewAgentListWatcher(clusterObj.Token, h.models.ListWatcherConfig)
			if err = agentListWatcher.NotifyResponse(&resp); err != nil {
				klog.Errorf("notify response error: %s", err.Error())
			}
		}
	}()
	return nil
}
