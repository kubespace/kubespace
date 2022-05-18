package views

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/sse"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"k8s.io/klog"
	"net/http"
	"strconv"
	"time"
)

type Cluster struct {
	Views  []*View
	models *model.Models
	*kube_resource.KubeResources
}

func NewCluster(models *model.Models, kr *kube_resource.KubeResources) *Cluster {
	cluster := &Cluster{
		models:        models,
		KubeResources: kr,
	}
	views := []*View{
		NewView(http.MethodGet, "", cluster.list),
		NewView(http.MethodPost, "", cluster.create),
		NewView(http.MethodPost, "/members", cluster.members),
		NewView(http.MethodGet, "/:cluster/detail", cluster.detail),
		NewView(http.MethodPost, "/delete", cluster.delete),
		NewView(http.MethodPost, "/apply/:cluster", cluster.apply),
		NewView(http.MethodPost, "/createYaml/:cluster", cluster.createYaml),
		NewView(http.MethodGet, "/:cluster/sse", cluster.resourceSSE),
	}
	cluster.Views = views
	return cluster
}

func (clu *Cluster) list(c *Context) *utils.Response {
	resp := &utils.Response{Code: code.Success}
	var filters map[string]interface{}
	clus, err := clu.models.ClusterManager.List(filters)
	if err != nil {
		resp.Code = code.GetError
		resp.Msg = err.Error()
		return resp
	}
	var data []map[string]interface{}

	for _, du := range clus {
		if !clu.models.UserRoleManager.HasScopeRole(c.User, types.RoleScopeCluster, du.ID, types.RoleTypeViewer) {
			continue
		}
		status := types.ClusterPending
		klog.Info(du.Name)
		clusterConnect := clu.Watch.KubeMessage.ClusterConnected(du.Name)
		if clusterConnect {
			status = types.ClusterConnect
		}
		data = append(data, map[string]interface{}{
			"id":          du.ID,
			"name":        du.Name,
			"name1":       du.Name1,
			"token":       du.Token,
			"status":      status,
			"created_by":  du.CreatedBy,
			"members":     du.Members,
			"create_time": du.CreateTime,
			"update_time": du.UpdateTime,
		})
	}
	resp.Data = data
	return resp
}

func (clu *Cluster) create(c *Context) *utils.Response {
	var ser serializers.ClusterCreateSerializers
	resp := &utils.Response{Code: code.Success}

	if err := c.ShouldBind(&ser); err != nil {
		resp.Code = code.ParamsError
		resp.Msg = err.Error()
		return resp
	}
	if ser.Name == "" {
		resp.Code = code.ParamsError
		resp.Msg = fmt.Sprintf("params cluster name:%s blank", ser.Name)
		return resp
	}
	cluster := &types.Cluster{
		Name1:     ser.Name,
		Token:     utils.CreateUUID(),
		Status:    types.ClusterPending,
		CreatedBy: c.User.Name,
		Members:   ser.Members,
	}
	cluster.CreateTime = time.Now()
	cluster.UpdateTime = time.Now()
	if err := clu.models.ClusterManager.Create(cluster); err != nil {
		resp.Code = code.CreateError
		resp.Msg = err.Error()
		return resp
	}
	d := map[string]interface{}{
		"id":          cluster.ID,
		"name1":       cluster.Name1,
		"name":        cluster.Name,
		"token":       cluster.Token,
		"status":      cluster.Status,
		"create_time": cluster.CreateTime,
		"update_time": cluster.UpdateTime,
	}
	resp.Data = d
	return resp
}

func (clu *Cluster) members(c *Context) *utils.Response {
	var ser serializers.ClusterCreateSerializers
	resp := &utils.Response{Code: code.Success}

	if err := c.ShouldBind(&ser); err != nil {
		resp.Code = code.ParamsError
		resp.Msg = err.Error()
		return resp
	}
	if ser.Name == "" {
		resp.Code = code.ParamsError
		resp.Msg = fmt.Sprintf("params cluster name:%s blank", ser.Name)
		return resp
	}
	cluster, err := clu.models.ClusterManager.GetByName(ser.Name)
	if err != nil {
		resp.Code = code.GetError
		resp.Msg = fmt.Sprintf("get cluster %s error: %s", ser.Name, err.Error())
		return resp
	}
	cluster.Members = ser.Members
	cluster.UpdateTime = time.Now()
	if err := clu.models.ClusterManager.Update(cluster); err != nil {
		resp.Code = code.UpdateError
		resp.Msg = err.Error()
		return resp
	}
	d := map[string]interface{}{
		"name":        cluster.Name,
		"token":       cluster.Token,
		"status":      cluster.Status,
		"create_time": cluster.CreateTime,
		"update_time": cluster.UpdateTime,
	}
	resp.Data = d
	return resp
}

func (clu *Cluster) detail(c *Context) *utils.Response {
	return clu.Cluster.Get(c.Param("cluster"), map[string]interface{}{})
}

func (clu *Cluster) apply(c *Context) *utils.Response {
	var ser serializers.ApplyYamlSerializers
	if err := c.ShouldBind(&ser); err != nil {
		klog.Errorf("bind params error: %s", err.Error())
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return clu.Cluster.Apply(c.Param("cluster"), ser)
}

func (clu *Cluster) createYaml(c *Context) *utils.Response {
	var ser serializers.ApplyYamlSerializers
	if err := c.ShouldBind(&ser); err != nil {
		klog.Errorf("bind params error: %s", err.Error())
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return clu.Cluster.Create(c.Param("cluster"), ser)
}

func (clu *Cluster) delete(c *Context) *utils.Response {
	var ser []serializers.DeleteClusterSerializers
	if err := c.ShouldBind(&ser); err != nil {
		klog.Errorf("bind params error: %s", err.Error())
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	for _, c := range ser {
		id, err := strconv.ParseUint(c.Id, 10, 64)
		if err != nil {
			return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
		}
		err = clu.models.ClusterManager.Delete(uint(id))
		if err != nil {
			klog.Errorf("delete cluster %s error: %s", c, err.Error())
			return &utils.Response{Code: code.DeleteError, Msg: err.Error()}
		}
	}
	return &utils.Response{Code: code.Success}
}

func (clu *Cluster) resourceSSE(c *Context) *utils.Response {
	if c.Param("cluster") == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "get param pipeline run id error"}
	}
	var ser serializers.ClusterSSESerializers
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	if ser.Type == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "参数type不能为空"}
	}
	watchSelector := map[string]string{
		sse.EventLabelType: ser.Type,
	}
	if ser.Namespace != "" {
		watchSelector[sse.EventClusterNamespace] = ser.Namespace
	}
	if ser.Uid != "" {
		watchSelector[sse.EventClusterUid] = ser.Uid
	}
	if ser.Selector != nil {
		for k, v := range ser.Selector {
			watchSelector[k] = v
		}
	}
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	streamClient := sse.StreamClient{
		Cluster:       c.Param("cluster"),
		ClientId:      utils.CreateUUID(),
		Catalog:       sse.CatalogCluster,
		WatchSelector: watchSelector,
		ClientChan:    make(chan sse.Event),
	}
	sse.Stream.AddClient(streamClient)
	defer sse.Stream.RemoveClient(streamClient)
	w := c.Writer
	clientGone := w.CloseNotify()
	c.SSEvent("message", "\n")
	w.Flush()
	tick := time.NewTicker(30 * time.Second)

	for {
		klog.Infof("select for cluster %s resource %s channel", ser.Cluster, ser.Type)
		select {
		case <-clientGone:
			klog.Info("select for cluster %s resource %s client gone", ser.Cluster, ser.Type)
			return nil
		case event := <-streamClient.ClientChan:
			c.SSEvent("message", event.Object)
			c.Writer.Flush()
		case <-tick.C:
			c.SSEvent("message", "\n")
			c.Writer.Flush()
		}
	}
}
