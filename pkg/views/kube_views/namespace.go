package kube_views

import (
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"net/http"
)

type Namespace struct {
	Views []*views.View
	*kube_resource.KubeResources
}

func NewNamespace(kr *kube_resource.KubeResources) *Namespace {
	ns := &Namespace{
		KubeResources: kr,
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "/:cluster", ns.list),
		views.NewView(http.MethodGet, "/:cluster/:name", ns.get),
		views.NewView(http.MethodPost, "/:cluster/delete", ns.delete),
		views.NewView(http.MethodPost, "/:cluster/update/:name", ns.updateYaml),
	}
	ns.Views = vs
	return ns
}

func (n *Namespace) list(c *views.Context) *utils.Response {
	var ser serializers.ListSerializers
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name": ser.Name,
	}
	return n.Namespace.List(c.Param("cluster"), reqParams)
}

func (n *Namespace) get(c *views.Context) *utils.Response {
	var ser serializers.GetSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name":   c.Param("name"),
		"output": ser.Output,
	}
	return n.Namespace.Get(c.Param("cluster"), reqParams)
}

func (n *Namespace) delete(c *views.Context) *utils.Response {
	var ser serializers.DeleteSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return n.Namespace.Delete(c.Param("cluster"), ser)
}

func (n *Namespace) updateYaml(c *views.Context) *utils.Response {
	var ser serializers.UpdateSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name": c.Param("name"),
		"yaml": ser.Yaml,
	}
	return n.Namespace.UpdateYaml(c.Param("cluster"), reqParams)
}
