package ws_views

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kubespace/kubespace/pkg/core/redis"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	kubewebsocket "github.com/kubespace/kubespace/pkg/server/websockets"
	"k8s.io/klog"
)

type KubeWs struct {
	redisOptions *redis.Options
	models       *model.Models
}

func NewKubeWs(op *redis.Options, models *model.Models) *KubeWs {
	return &KubeWs{
		redisOptions: op,
		models:       models,
	}
}

func (k *KubeWs) Connect(c *gin.Context) {

	token := c.GetHeader("token")

	cluster, err := k.models.ClusterManager.GetByToken(token)
	if err != nil {
		klog.Errorf("get cluster error")
		c.Data(404, "", []byte("not found cluster with token "+token))
		return
	}
	if cluster == nil {
		klog.Errorf("can not get cluster")
		c.Data(404, "", []byte("not found cluster with token "+token))
		return
	}

	upGrader := &websocket.Upgrader{}
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		klog.Errorf("upgrader agent conn error: %s", err)
		c.Data(500, "", []byte("upgrader agent conn error: "+err.Error()))
		return
	}

	kubeWebsocket, err := kubewebsocket.NewKubeWebsocket(cluster.Name, ws, k.redisOptions, k.models)
	if err != nil {
		klog.Errorf("create websocket error: %s", err)
		c.Data(500, "", []byte(err.Error()))
		return
	}
	kubeWebsocket.Consume()
	cluster.Status = types.ClusterConnect
	k.models.ClusterManager.Update(cluster)
	klog.V(1).Infof("cluster %s kube connect finish", cluster.Name)
}
