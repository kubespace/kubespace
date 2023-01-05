package kube_views

import (
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"k8s.io/klog"
	"net/http"
)

type Deployment struct {
	Views []*views.View
	*kube_resource.KubeResources
}

func NewDeployment(kr *kube_resource.KubeResources) *Deployment {
	d := &Deployment{
		KubeResources: kr,
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "/:cluster/:namespace/:name", d.get),
		views.NewView(http.MethodGet, "/:cluster", d.list),
		views.NewView(http.MethodPost, "/:cluster/delete", d.delete),
		views.NewView(http.MethodPost, "/:cluster/update/:namespace/:name", d.updateYaml),
		views.NewView(http.MethodPost, "/:cluster/update_obj/:namespace/:name", d.updateObj),
	}
	d.Views = vs
	return d
}

func (d *Deployment) list(c *views.Context) *utils.Response {
	var ser serializers.ListSerializers
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name":      ser.Name,
		"namespace": ser.Namespace,
	}
	return d.Deployment.List(c.Param("cluster"), reqParams)
}

func (d *Deployment) get(c *views.Context) *utils.Response {
	var ser serializers.GetSerializers
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name":      c.Param("name"),
		"namespace": c.Param("namespace"),
		"output":    ser.Output,
	}
	return d.Deployment.Get(c.Param("cluster"), reqParams)
}

func (d *Deployment) delete(c *views.Context) *utils.Response {
	var ser serializers.DeleteSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return d.Deployment.Delete(c.Param("cluster"), ser)
}

func (d *Deployment) updateYaml(c *views.Context) *utils.Response {
	var ser serializers.UpdateSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name":      c.Param("name"),
		"namespace": c.Param("namespace"),
		"yaml":      ser.Yaml,
	}
	return d.Deployment.UpdateYaml(c.Param("cluster"), reqParams)
}

func (d *Deployment) updateObj(c *views.Context) *utils.Response {
	var ser serializers.UpdateWorkloadSerializers
	if err := c.ShouldBindJSON(&ser); err != nil {
		klog.Errorf("bind update date error: %s", err.Error())
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name":      c.Param("name"),
		"namespace": c.Param("namespace"),
		"replicas":  ser.Replicas,
	}
	return d.Deployment.UpdateObj(c.Param("cluster"), reqParams)
}
