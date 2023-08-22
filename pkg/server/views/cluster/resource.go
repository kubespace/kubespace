package cluster

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/kubernetes/resource"
	kubetypes "github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/service/cluster"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
	"net/http"
	"strconv"
	"strings"
)

type KubeResource struct {
	Views  []*views.View
	client *cluster.KubeClient
	models *model.Models
}

func NewKubeResource(config *config.ServerConfig) *KubeResource {
	res := &KubeResource{
		models: config.Models,
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
	c.SSEvent("message", "{}")
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
	var scope, scopeName string
	var scopeId uint
	if c.Query("project_id") != "" {
		projectId, _ := strconv.Atoi(c.Query("project_id"))
		if projectId == 0 {
			return &utils.Response{Code: code.ParamsError, Msg: "project_id参数错误"}
		}
		projectObj, err := p.models.ProjectManager.Get(uint(projectId))
		if err != nil {
			return &utils.Response{Code: code.ParamsError, Msg: fmt.Sprintf("获取工作空间id=%d错误：%s", projectId, err.Error())}
		}
		scope = types.ScopeProject
		scopeId = projectObj.ID
		scopeName = projectObj.Name
	} else {
		clusterId, _ := strconv.Atoi(c.Param("cluster"))
		clusterObj, err := p.models.ClusterManager.Get(uint(clusterId))
		if err != nil {
			return &utils.Response{Code: code.ParamsError, Msg: fmt.Sprintf("获取集群id=%d错误：%s", clusterId, err.Error())}
		}
		if clusterObj == nil {
			return &utils.Response{Code: code.GetError, Msg: fmt.Sprintf("未找到集群id=%d", clusterId)}
		}
		scope = types.ScopeCluster
		scopeId = clusterObj.ID
		scopeName = clusterObj.Name1
	}
	resp := p.client.Delete(c.Param("cluster"), c.Param("resType"), &ser)
	namespace := ""
	namespaceMap := make(map[string]struct{})
	var delNames []string
	var delNamesWithNamespace []string
	var opDetail = ""
	var delNameStr string
	if len(ser.Resources) > 0 {
		for _, r := range ser.Resources {
			if r.Namespace != "" {
				namespaceMap[r.Namespace] = struct{}{}
			}
			delNames = append(delNames, r.Name)
			delNamesWithNamespace = append(delNamesWithNamespace, r.Namespace+"/"+r.Name)
		}
		var namespaces []string
		for n := range namespaceMap {
			namespaces = append(namespaces, n)
		}
		namespace = strings.Join(namespaces, ",")
		if len(namespaces) > 1 {
			delNameStr = strings.Join(delNamesWithNamespace, ",")
		} else {
			delNameStr = strings.Join(delNames, ",")
		}
		opDetail = fmt.Sprintf("删除%s:%s", c.Param("resType"), delNameStr)
	} else {
		namespace = ser.Namespace
		opDetail = fmt.Sprintf("删除%s:%s", c.Param("resType"), ser.LabelSelector.String())
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationDelete,
		OperateDetail:        opDetail,
		Scope:                scope,
		ScopeId:              scopeId,
		ScopeName:            scopeName,
		Namespace:            namespace,
		ResourceType:         c.Param("resType"),
		ResourceName:         delNameStr,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: &ser,
	})
	return resp
}

func (p *KubeResource) update(c *views.Context) *utils.Response {
	var ser resource.UpdateParams
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	ser.Namespace = c.Param("namespace")
	ser.Name = c.Param("name")

	var scope, scopeName string
	var scopeId uint
	if c.Query("project_id") != "" {
		projectId, _ := strconv.Atoi(c.Query("project_id"))
		if projectId == 0 {
			return &utils.Response{Code: code.ParamsError, Msg: "project_id参数错误"}
		}
		projectObj, err := p.models.ProjectManager.Get(uint(projectId))
		if err != nil {
			return &utils.Response{Code: code.ParamsError, Msg: fmt.Sprintf("获取工作空间id=%d错误：%s", projectId, err.Error())}
		}
		scope = types.ScopeProject
		scopeId = projectObj.ID
		scopeName = projectObj.Name
	} else {
		clusterId, _ := strconv.Atoi(c.Param("cluster"))
		clusterObj, err := p.models.ClusterManager.Get(uint(clusterId))
		if err != nil {
			return &utils.Response{Code: code.ParamsError, Msg: fmt.Sprintf("获取集群id=%d错误：%s", clusterId, err.Error())}
		}
		if clusterObj == nil {
			return &utils.Response{Code: code.GetError, Msg: fmt.Sprintf("未找到集群id=%d", clusterId)}
		}
		scope = types.ScopeCluster
		scopeId = clusterObj.ID
		scopeName = clusterObj.Name1
	}

	resp := p.client.Update(c.Param("cluster"), c.Param("resType"), &ser)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationUpdate,
		OperateDetail:        "更新" + c.Param("resType") + ":" + ser.Name,
		Scope:                scope,
		ScopeId:              scopeId,
		ScopeName:            scopeName,
		Namespace:            ser.Namespace,
		ResourceType:         c.Param("resType"),
		ResourceName:         ser.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: &ser,
	})
	return resp
}

func (p *KubeResource) patch(c *views.Context) *utils.Response {
	var ser resource.PatchParams
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	var scope, scopeName string
	var scopeId uint
	if c.Query("project_id") != "" {
		projectId, _ := strconv.Atoi(c.Query("project_id"))
		if projectId == 0 {
			return &utils.Response{Code: code.ParamsError, Msg: "project_id参数错误"}
		}
		projectObj, err := p.models.ProjectManager.Get(uint(projectId))
		if err != nil {
			return &utils.Response{Code: code.ParamsError, Msg: fmt.Sprintf("获取工作空间id=%d错误：%s", projectId, err.Error())}
		}
		scope = types.ScopeProject
		scopeId = projectObj.ID
		scopeName = projectObj.Name
	} else {
		clusterId, _ := strconv.Atoi(c.Param("cluster"))
		clusterObj, err := p.models.ClusterManager.Get(uint(clusterId))
		if err != nil {
			return &utils.Response{Code: code.ParamsError, Msg: fmt.Sprintf("获取集群id=%d错误：%s", clusterId, err.Error())}
		}
		if clusterObj == nil {
			return &utils.Response{Code: code.GetError, Msg: fmt.Sprintf("未找到集群id=%d", clusterId)}
		}
		scope = types.ScopeCluster
		scopeId = clusterObj.ID
		scopeName = clusterObj.Name1
	}
	resp := p.client.Patch(c.Param("cluster"), c.Param("resType"), &ser)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationPatch,
		OperateDetail:        fmt.Sprintf("Patch %s %s/%s", c.Param("resType"), ser.Namespace, ser.Name),
		Scope:                scope,
		ScopeId:              scopeId,
		ScopeName:            scopeName,
		Namespace:            ser.Namespace,
		ResourceType:         c.Param("resType"),
		ResourceName:         ser.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: &ser,
	})
	return resp
}

func (p *KubeResource) apply(c *views.Context) *utils.Response {
	var ser resource.ApplyParams
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	var scope, scopeName string
	var scopeId uint
	if c.Query("project_id") != "" {
		projectId, _ := strconv.Atoi(c.Query("project_id"))
		if projectId == 0 {
			return &utils.Response{Code: code.ParamsError, Msg: "project_id参数错误"}
		}
		projectObj, err := p.models.ProjectManager.Get(uint(projectId))
		if err != nil {
			return &utils.Response{Code: code.ParamsError, Msg: fmt.Sprintf("获取工作空间id=%d错误：%s", projectId, err.Error())}
		}
		scope = types.ScopeProject
		scopeId = projectObj.ID
		scopeName = projectObj.Name
	} else {
		clusterId, _ := strconv.Atoi(c.Param("cluster"))
		clusterObj, err := p.models.ClusterManager.Get(uint(clusterId))
		if err != nil {
			return &utils.Response{Code: code.ParamsError, Msg: fmt.Sprintf("获取集群id=%d错误：%s", clusterId, err.Error())}
		}
		if clusterObj == nil {
			return &utils.Response{Code: code.GetError, Msg: fmt.Sprintf("未找到集群id=%d", clusterId)}
		}
		scope = types.ScopeCluster
		scopeId = clusterObj.ID
		scopeName = clusterObj.Name1
	}
	resp := p.client.Apply(c.Param("cluster"), &ser)
	var applyResources []*resource.ApplyResource
	if err := utils.ConvertTypeByJson(resp.Data, &applyResources); err != nil {
		klog.Errorf("convert apply resource error: %s", err.Error())
	} else if len(applyResources) > 0 {
		var resNames []string
		var resNamesWithKind []string
		namespaceMap := make(map[string]struct{})
		resTypeMap := make(map[string]struct{})
		for _, r := range applyResources {
			if r.Namespace != "" {
				namespaceMap[r.Namespace] = struct{}{}
			}
			if r.Kind != "" {
				resTypeMap[strings.ToLower(r.Kind)] = struct{}{}
			}
			resNames = append(resNames, r.Name)
			resNamesWithKind = append(resNamesWithKind, r.Kind+"/"+r.Name)
		}
		var namespaces []string
		var resTypes []string
		for n := range namespaceMap {
			namespaces = append(namespaces, n)
		}
		for t := range resTypeMap {
			resTypes = append(resTypes, t)
		}
		operation := types.AuditOperationApply
		opDetail := "Apply resources: " + strings.Join(resNamesWithKind, ",")
		if ser.Create {
			operation = types.AuditOperationCreate
			opDetail = "创建资源：" + strings.Join(resNamesWithKind, ",")
		}
		c.CreateAudit(&types.AuditOperate{
			Operation:            operation,
			OperateDetail:        opDetail,
			Scope:                scope,
			ScopeId:              scopeId,
			ScopeName:            scopeName,
			Namespace:            strings.Join(namespaces, ","),
			ResourceType:         strings.Join(resTypes, ","),
			ResourceName:         strings.Join(resNames, ","),
			Code:                 resp.Code,
			Message:              resp.Msg,
			OperateDataInterface: &ser,
		})
	}
	return resp
}
