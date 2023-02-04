package cluster

import (
	"github.com/gorilla/websocket"
	"github.com/kubespace/kubespace/pkg/kubernetes/resource"
	kubetypes "github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/service/cluster"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"k8s.io/klog/v2"
	"net/http"
)

type KubeResource struct {
	Views  []*views.View
	client *cluster.KubeClient
}

func NewKubeResource(config *config.ServerConfig) *KubeResource {
	res := &KubeResource{
		client: config.ServiceFactory.Cluster.KubeClient,
	}
	res.Views = []*views.View{
		views.NewView(http.MethodPost, "/apply", res.apply),
		views.NewView(http.MethodPost, "/:resType/list", res.list),
		views.NewView(http.MethodGet, "/:resType/watch", res.watch),
		views.NewView(http.MethodGet, "/:resType/namespace/:namespace/:name", res.get),
		views.NewView(http.MethodGet, "/:resType/:name", res.get),
		views.NewView(http.MethodPost, "/:resType/delete", res.delete),
		views.NewView(http.MethodGet, "/pod/exec/:namespace/:pod", res.podExec),
		views.NewView(http.MethodGet, "/pod/log/:namespace/:pod", res.podLog),
		views.NewView(http.MethodPut, "/:resType/namespace/:namespace/:name", res.update),
		views.NewView(http.MethodPut, "/:resType/:name", res.update),
		views.NewView(http.MethodPost, "/:resType/patch", res.patch),
	}
	return res
}

func (p *KubeResource) list(c *views.Context) *utils.Response {
	var params interface{}
	if c.Param("resType") == kubetypes.CustomResourceType {
		params = &resource.CustomResourceQueryParams{}
	} else {
		params = &resource.QueryParams{}
	}
	if err := c.ShouldBind(params); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return p.client.List(c.Param("cluster"), c.Param("resType"), params)
}

func (p *KubeResource) podExec(c *views.Context) *utils.Response {
	upGrader := &websocket.Upgrader{}
	upGrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		klog.Errorf("upgrader agent conn error: %s", err)
		return nil
	}
	podexec, err := newPodExec(ws, p.client, c.Param("cluster"), &podExecParams{
		Namespace: c.Param("namespace"),
		Name:      c.Param("pod"),
		Container: c.Query("container"),
		Rows:      c.Query("rows"),
		Cols:      c.Query("cols"),
	})
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		ws.Close()
		return nil
	}
	go podexec.consume()
	return nil
}

func (p *KubeResource) podLog(c *views.Context) *utils.Response {
	upGrader := &websocket.Upgrader{}
	upGrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		klog.Errorf("upgrader agent conn error: %s", err)
		return nil
	}
	podlog, err := newPodLog(ws, p.client, c.Param("cluster"), &podLogParams{
		Namespace: c.Param("namespace"),
		Name:      c.Param("pod"),
		Container: c.Query("container"),
	})
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		ws.Close()
		return nil
	}
	go podlog.consume()
	return nil
}

func (p *KubeResource) get(c *views.Context) *utils.Response {
	var params interface{}
	if c.Param("resType") == kubetypes.CustomResourceType {
		params = &resource.CustomResourceQueryParams{
			Name:      c.Param("name"),
			Namespace: c.Param("namespace"),
		}
	} else {
		params = &resource.QueryParams{
			Name:      c.Param("name"),
			Namespace: c.Param("namespace"),
		}
	}
	if err := c.ShouldBindQuery(params); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return p.client.Get(c.Param("cluster"), c.Param("resType"), params)
}

func (p *KubeResource) watch(c *views.Context) *utils.Response {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	//c.Writer.Header().Set("Connection", "keep-alive")
	if c.Param("cluster") == "" {
		c.SSEvent("message", "get param cluster error")
		return nil
	}
	var ser resource.QueryParams
	if err := c.ShouldBindQuery(&ser); err != nil {
		c.SSEvent("message", err.Error())
		return nil
	}
	watchOuter, err := p.client.Watch(c.Param("cluster"), c.Param("resType"), &ser)
	if err != nil {
		c.SSEvent("message", err.Error())
		return nil
	}
	c.SSEvent("message", "\n")
	c.Writer.Flush()
	defer watchOuter.Close()
	for {
		select {
		case <-c.Writer.CloseNotify():
			klog.Info("select for cluster %s resource %s client gone", c.Param("cluster"), c.Param("resType"))
			return nil
		case event := <-watchOuter.OutCh():
			c.SSEvent("message", event)
			c.Writer.Flush()
		case <-watchOuter.StopCh():
			return nil
		}
	}
}

func (p *KubeResource) delete(c *views.Context) *utils.Response {
	var ser resource.DeleteParams
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return p.client.Delete(c.Param("cluster"), c.Param("resType"), &ser)
}

func (p *KubeResource) update(c *views.Context) *utils.Response {
	var ser resource.UpdateParams
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	ser.Namespace = c.Param("namespace")
	ser.Name = c.Param("name")
	return p.client.Update(c.Param("cluster"), c.Param("resType"), &ser)
}

func (p *KubeResource) patch(c *views.Context) *utils.Response {
	var ser resource.PatchParams
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return p.client.Patch(c.Param("cluster"), c.Param("resType"), &ser)
}

func (p *KubeResource) apply(c *views.Context) *utils.Response {
	var ser resource.ApplyParams
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return p.client.Apply(c.Param("cluster"), &ser)
}
