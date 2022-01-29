package kube_views

import (
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"net/http"
)

type Statefulset struct {
	Views []*views.View
	*kube_resource.KubeResources
}

func NewStatefulset(kr *kube_resource.KubeResources) *Statefulset {
	d := &Statefulset{
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

func (d *Statefulset) list(c *views.Context) *utils.Response {
	var ser serializers.ListSerializers
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name":      ser.Name,
		"namespace": ser.Namespace,
	}
	return d.Statefulset.List(c.Param("cluster"), reqParams)
}

func (d *Statefulset) get(c *views.Context) *utils.Response {
	var ser serializers.GetSerializers
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name":      c.Param("name"),
		"namespace": c.Param("namespace"),
		"output":    ser.Output,
	}
	return d.Statefulset.Get(c.Param("cluster"), reqParams)
}

func (d *Statefulset) delete(c *views.Context) *utils.Response {
	var ser serializers.DeleteSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return d.Statefulset.Delete(c.Param("cluster"), ser)
}

func (d *Statefulset) updateYaml(c *views.Context) *utils.Response {
	var ser serializers.UpdateSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name":      c.Param("name"),
		"namespace": c.Param("namespace"),
		"yaml":      ser.Yaml,
	}
	return d.Statefulset.UpdateYaml(c.Param("cluster"), reqParams)
}

func (d *Statefulset) updateObj(c *views.Context) *utils.Response {
	var ser serializers.UpdateWorkloadSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name":      c.Param("name"),
		"namespace": c.Param("namespace"),
		"replicas":  ser.Replicas,
	}
	return d.Statefulset.UpdateObj(c.Param("cluster"), reqParams)
}
