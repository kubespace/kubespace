package ws_views

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kubespace/kubespace/pkg/core/redis"
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/model"
	kubewebsocket "github.com/kubespace/kubespace/pkg/server/websockets"
	"k8s.io/klog"
	"net/http"
)

type ApiWs struct {
	redisOptions *redis.Options
	models       *model.Models
	*kube_resource.KubeResources
}

func NewApiWs(op *redis.Options, models *model.Models, kr *kube_resource.KubeResources) *ApiWs {
	return &ApiWs{
		redisOptions:  op,
		models:        models,
		KubeResources: kr,
	}
}

func (a *ApiWs) Connect(c *gin.Context) {
	upGrader := &websocket.Upgrader{}
	upGrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		klog.Errorf("upgrader agent conn error: %s", err)
		return
	}
	token, err := c.Cookie("osp-token")
	if err != nil {
		klog.Errorf("auth token error: %s", err.Error())
		ws.WriteMessage(websocket.TextMessage, []byte("auth token error"))
		ws.Close()
		return
	}
	klog.V(1).Info(token)
	_, err = a.models.TokenManager.Get(token)
	if err != nil {
		klog.Errorf("auth token error: %s", err.Error())
		ws.WriteMessage(websocket.TextMessage, []byte("auth token error"))
		ws.Close()
		return
	}

	apiWebsocket := kubewebsocket.NewApiWebsocket(ws, a.redisOptions, a.KubeResources)
	go apiWebsocket.Consume()
	klog.V(1).Info("cluster api connect finish")
}
