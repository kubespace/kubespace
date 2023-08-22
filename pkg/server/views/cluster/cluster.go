package cluster

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	kubetypes "github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/service/cluster"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
	"net/http"
	"strconv"
	"sync"
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
		views.NewView(http.MethodPut, "/:cluster", clu.update),
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

	var wg sync.WaitGroup
	for _, du := range clus {
		if !clu.models.UserRoleManager.HasScopeRole(c.User, types.ScopeCluster, du.ID, types.RoleTypeViewer) {
			continue
		}
		wg.Add(1)
		go func(du types.Cluster) {
			defer wg.Done()
			status := types.ClusterPending
			clusterVersion := ""
			connectErr := ""
			res := clu.kubeClient.Get(du.Name, kubetypes.ClusterType, map[string]interface{}{"only_version": true})
			if res.IsSuccess() {
				status = types.ClusterConnect
				clusterVersion, _ = res.Data.(string)
			} else {
				if du.KubeConfig != "" {
					status = types.ClusterFailed
				}
				connectErr = res.Msg
			}
			data = append(data, map[string]interface{}{
				"id":            du.ID,
				"name":          du.Name,
				"name1":         du.Name1,
				"token":         du.Token,
				"status":        status,
				"connect_error": connectErr,
				"version":       clusterVersion,
				"created_by":    du.CreatedBy,
				"members":       du.Members,
				"create_time":   du.CreateTime,
				"update_time":   du.UpdateTime,
			})
		}(du)
	}
	wg.Wait()
	resp.Data = data
	return resp
}

func (clu *Cluster) create(c *views.Context) *utils.Response {
	var ser serializers.ClusterCreateSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	if ser.Name == "" {
		return c.GenerateResponseError(errors.New(code.ParamsError, "cluster name is blank"))
	}
	clusterObj := &types.Cluster{
		Name1:      ser.Name,
		Token:      utils.ShortUUID(),
		Status:     types.ClusterPending,
		CreatedBy:  c.User.Name,
		Members:    ser.Members,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	err := clu.models.ClusterManager.Create(clusterObj)
	if err != nil {
		err = errors.New(code.CreateError, err)
	}
	resp := c.GenerateResponse(err, clusterObj)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationCreate,
		OperateDetail:        "创建集群：" + clusterObj.Name1,
		Scope:                types.ScopeCluster,
		ScopeId:              clusterObj.ID,
		ScopeName:            clusterObj.Name1,
		ResourceId:           clusterObj.ID,
		ResourceType:         types.AuditResourceCluster,
		ResourceName:         clusterObj.Name1,
		Code:                 code.Success,
		OperateDataInterface: clusterObj,
	})
	return resp
}

func (clu *Cluster) update(c *views.Context) *utils.Response {
	var ser serializers.ClusterUpdateSerializers
	resp := &utils.Response{Code: code.Success}

	if err := c.ShouldBind(&ser); err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	clusterId, err := strconv.Atoi(c.Param("cluster"))
	if err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	clusterObj, err := clu.models.ClusterManager.Get(uint(clusterId))
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DataNotExists, fmt.Sprintf("not found cluster id=%d", clusterId)))
	}
	if err = clu.models.ClusterManager.UpdateByObject(uint(clusterId), &types.Cluster{KubeConfig: ser.KubeConfig}); err != nil {
		return &utils.Response{Code: code.UpdateError, Msg: err.Error()}
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationUpdate,
		OperateDetail:        "更新集群：" + clusterObj.Name1,
		Scope:                types.ScopeCluster,
		ScopeId:              clusterObj.ID,
		ScopeName:            clusterObj.Name1,
		ResourceId:           clusterObj.ID,
		ResourceType:         types.AuditResourceCluster,
		ResourceName:         clusterObj.Name1,
		Code:                 code.Success,
		OperateDataInterface: ser,
	})
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
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	for _, delCluster := range ser {
		resp := &utils.Response{Code: code.Success}
		id, err := strconv.ParseUint(delCluster.Id, 10, 64)
		if err != nil {
			return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
		}
		clusterObj, err := clu.models.ClusterManager.Get(uint(id))
		if err != nil {
			return &utils.Response{Code: code.DataNotExists, Msg: fmt.Sprintf("not found cluster id=%d", id)}
		}
		err = clu.models.ClusterManager.Delete(uint(id))
		if err != nil {
			klog.Errorf("delete cluster %s error: %s", c, err.Error())
			resp = &utils.Response{Code: code.DeleteError, Msg: err.Error()}
		}
		clusterObj.KubeConfig = ""

		c.CreateAudit(&types.AuditOperate{
			Operation:            types.AuditOperationDelete,
			OperateDetail:        "删除集群：" + clusterObj.Name1,
			Scope:                types.ScopeCluster,
			ScopeId:              clusterObj.ID,
			ScopeName:            clusterObj.Name1,
			ResourceId:           clusterObj.ID,
			ResourceType:         types.AuditResourceCluster,
			ResourceName:         clusterObj.Name1,
			Code:                 resp.Code,
			Message:              resp.Msg,
			OperateDataInterface: clusterObj,
		})
		if !resp.IsSuccess() {
			return resp
		}
	}
	return &utils.Response{Code: code.Success}
}
