package cluster

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kubespace/kubespace/pkg/informer"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/cluster"
	kubetypes "github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/config"
	"k8s.io/klog/v2"
	"net/http"
)

func resolveHost(r *http.Request) (host string) {
	switch {
	case r.Header.Get("X-Host") != "":
		return r.Header.Get("X-Host")
	case r.Host != "":
		return r.Host
	case r.URL.Host != "":
		return r.URL.Host
	default:
		return r.Host
	}
}

type AgentViews struct {
	models          *model.Models
	informerFactory informer.Factory
	AgentVersion    string
	AgentRepository string
}

func NewAgentViews(conf *config.ServerConfig) *AgentViews {
	return &AgentViews{
		models:          conf.Models,
		informerFactory: conf.InformerFactory,
		AgentVersion:    conf.AgentVersion,
		AgentRepository: conf.AgentRepository,
	}
}

func (a *AgentViews) AgentYaml(c *gin.Context) {
	token := c.Param("token")
	serverUrl := resolveHost(c.Request)
	agentYaml := fmt.Sprintf(clusterAgentYaml, a.AgentRepository, a.AgentVersion, token, serverUrl)
	c.String(200, agentYaml)
}

func (a *AgentViews) processWebsocket(c *gin.Context) (*types.Cluster, *websocket.Conn, error) {
	token := c.GetHeader("token")

	clusterObj, err := a.models.ClusterManager.GetByToken(token)
	if err != nil {
		c.Data(404, "", []byte("not found cluster with token "+token))
		return nil, nil, fmt.Errorf("not found cluster with token " + token)
	}
	if clusterObj == nil {
		c.Data(404, "", []byte("not found cluster with token "+token))
		return nil, nil, fmt.Errorf("not found cluster with token " + token)
	}

	upGrader := &websocket.Upgrader{}
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.Data(500, "", []byte("upgrader agent conn error: "+err.Error()))
		return nil, nil, fmt.Errorf("upgrader agent conn error: " + err.Error())
	}
	return clusterObj, ws, nil
}

func (a *AgentViews) Connect(c *gin.Context) {
	clusterObj, ws, err := a.processWebsocket(c)
	if err != nil {
		klog.Errorf("process websocket error: %s", err.Error())
		return
	}

	tunnel, err := newAgentTunnel(ws, a.informerFactory.ClusterAgentInformer(clusterObj.Token))
	if err != nil {
		klog.Errorf("create websocket error: %s", err)
		c.Data(500, "", []byte(err.Error()))
		return
	}
	tunnel.Consume()
	clusterObj.Status = types.ClusterConnect
	a.models.ClusterManager.Update(clusterObj)
	klog.V(1).Infof("cluster %s kube connect finish", clusterObj.Name)
}

func (a *AgentViews) Response(c *gin.Context) {
	clusterObj, ws, err := a.processWebsocket(c)
	if err != nil {
		klog.Errorf("process websocket error: %s", err.Error())
		return
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
			agentListWatcher := cluster.NewAgentListWatcher(clusterObj.Token, a.models.ListWatcherConfig)
			if err = agentListWatcher.NotifyResponse(&resp); err != nil {
				klog.Errorf("notify response error: %s", err.Error())
			}
		}
	}()
}
