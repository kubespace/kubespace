package kube_views

import (
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"net/http"
)

type Rolebinding struct {
	Views []*views.View
	*kube_resource.KubeResources
}

func NewRolebinding(kr *kube_resource.KubeResources) *Rolebinding {
	d := &Rolebinding{
		KubeResources: kr,
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "/:cluster/:namespace/:name", d.get),
		views.NewView(http.MethodGet, "/:cluster", d.list),
		views.NewView(http.MethodPost, "/:cluster/delete", d.delete),
		views.NewView(http.MethodPost, "/:cluster/update/:namespace/:name", d.updateYaml),
	}
	d.Views = vs
	return d
}

func (d *Rolebinding) list(c *views.Context) *utils.Response {
	var ser serializers.ListSerializers
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name":      ser.Name,
		"namespace": ser.Namespace,
	}
	return d.Rolebinding.List(c.Param("cluster"), reqParams)
}

func (d *Rolebinding) get(c *views.Context) *utils.Response {
	var ser serializers.GetSerializers
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name":      c.Param("name"),
		"namespace": c.Param("namespace"),
		"output":    ser.Output,
		"kind":      ser.Kind,
	}
	return d.Rolebinding.Get(c.Param("cluster"), reqParams)
}

func (d *Rolebinding) delete(c *views.Context) *utils.Response {
	var ser serializers.DeleteSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return d.Rolebinding.Delete(c.Param("cluster"), ser)
}

func (d *Rolebinding) updateYaml(c *views.Context) *utils.Response {
	var ser serializers.UpdateSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	namespace := c.Param("namespace")
	if ser.Kind == "ClusterRoleBinding" {
		namespace = ""
	}
	reqParams := map[string]interface{}{
		"name":      c.Param("name"),
		"namespace": namespace,
		"yaml":      ser.Yaml,
		"kind":      ser.Kind,
	}
	return d.Rolebinding.UpdateYaml(c.Param("cluster"), reqParams)
}
