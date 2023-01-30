package cluster

import (
	"fmt"
	kubetypes "github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/service/cluster"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"k8s.io/klog/v2"
	"net/http"
	"strconv"
	"time"
)

type Cluster struct {
	Views      []*views.View
	models     *model.Models
	kubeClient *cluster.KubeClient
}

func NewCluster(config *config.ServerConfig) *Cluster {
	clu := &Cluster{
		models:     config.Models,
		kubeClient: config.ServiceFactory.Cluster.KubeClient,
	}
	clu.Views = []*views.View{
		views.NewView(http.MethodGet, "", clu.list),
		views.NewView(http.MethodPost, "", clu.create),
		views.NewView(http.MethodPost, "/members", clu.members),
		views.NewView(http.MethodPost, "/delete", clu.delete),
	}
	return clu
}

func (clu *Cluster) list(c *views.Context) *utils.Response {
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
		res := clu.kubeClient.Get(du.Name, kubetypes.ClusterType, map[string]interface{}{"only_version": true})
		if res.IsSuccess() {
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

func (clu *Cluster) create(c *views.Context) *utils.Response {
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
	clusterObj := &types.Cluster{
		Name1:      ser.Name,
		Token:      utils.CreateUUID(),
		Status:     types.ClusterPending,
		CreatedBy:  c.User.Name,
		Members:    ser.Members,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	if err := clu.models.ClusterManager.Create(clusterObj); err != nil {
		resp.Code = code.CreateError
		resp.Msg = err.Error()
		return resp
	}
	d := map[string]interface{}{
		"id":          clusterObj.ID,
		"name1":       clusterObj.Name1,
		"name":        clusterObj.Name,
		"token":       clusterObj.Token,
		"status":      clusterObj.Status,
		"create_time": clusterObj.CreateTime,
		"update_time": clusterObj.UpdateTime,
	}
	resp.Data = d
	return resp
}

func (clu *Cluster) members(c *views.Context) *utils.Response {
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
	clusterObj, err := clu.models.ClusterManager.GetByName(ser.Name)
	if err != nil {
		resp.Code = code.GetError
		resp.Msg = fmt.Sprintf("get cluster %s error: %s", ser.Name, err.Error())
		return resp
	}
	clusterObj.Members = ser.Members
	clusterObj.UpdateTime = time.Now()
	if err := clu.models.ClusterManager.Update(clusterObj); err != nil {
		resp.Code = code.UpdateError
		resp.Msg = err.Error()
		return resp
	}
	d := map[string]interface{}{
		"name":        clusterObj.Name,
		"token":       clusterObj.Token,
		"status":      clusterObj.Status,
		"create_time": clusterObj.CreateTime,
		"update_time": clusterObj.UpdateTime,
	}
	resp.Data = d
	return resp
}

func (clu *Cluster) delete(c *views.Context) *utils.Response {
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
