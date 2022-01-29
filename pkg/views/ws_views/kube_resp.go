package ws_views

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/redis"
	kubewebsocket "github.com/kubespace/kubespace/pkg/websockets"
	"k8s.io/klog"
)

type KubeResp struct {
	redisOptions *redis.Options
	models       *model.Models
}

func NewKubeResp(op *redis.Options, models *model.Models) *KubeResp {
	return &KubeResp{
		redisOptions: op,
		models:       models,
	}
}

func (k *KubeResp) Connect(c *gin.Context) {

	token := c.GetHeader("token")
	klog.V(1).Info(token)

	cluster, err := k.models.ClusterManager.GetByToken(token)
	if err != nil {
		klog.Errorf("get cluster error")
		return
	}
	if cluster == nil {
		klog.Errorf("can not get cluster")
		return
	}

	upGrader := &websocket.Upgrader{}
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		klog.Errorf("upgrader agent conn error: %s", err)
		return
	}
	kubeWebsocket := kubewebsocket.NewKubeRespWebsocket(cluster.Name, ws, k.redisOptions, k.models)
	kubeWebsocket.Consume()
	klog.V(1).Infof("cluster %s kube response connect finish", cluster.Name)
}
