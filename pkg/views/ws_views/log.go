package ws_views

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/redis"
	kubewebsocket "github.com/kubespace/kubespace/pkg/websockets"
	"k8s.io/klog"
	"net/http"
)

type LogWs struct {
	redisOptions *redis.Options
	models       *model.Models
	*kube_resource.KubeResources
}

func NewLogWs(op *redis.Options, models *model.Models, kr *kube_resource.KubeResources) *LogWs {
	return &LogWs{
		redisOptions:  op,
		models:        models,
		KubeResources: kr,
	}
}

func (l *LogWs) Connect(c *gin.Context) {
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
	_, err = l.models.TokenManager.Get(token)
	if err != nil {
		klog.Errorf("auth token error: %s", err.Error())
		ws.WriteMessage(websocket.TextMessage, []byte("auth token error"))
		ws.Close()
		return
	}

	container := c.Query("container")

	cluster, ok := c.Params.Get("cluster")
	if !ok {
		ws.WriteMessage(websocket.TextMessage, []byte("get cluster params error"))
		ws.Close()
		return
	}
	namespace, ok := c.Params.Get("namespace")
	if !ok {
		ws.WriteMessage(websocket.TextMessage, []byte("get namespace params error"))
		ws.Close()
		return
	}
	pod, ok := c.Params.Get("pod")
	if !ok {
		ws.WriteMessage(websocket.TextMessage, []byte("get pod params error"))
		ws.Close()
		return
	}

	logWebsocket := kubewebsocket.NewLogWebsocket(cluster, ws, l.redisOptions, l.KubeResources,
		namespace, pod, container)
	go logWebsocket.Consume()
	klog.V(1).Info("log websocket connect finish")
}
