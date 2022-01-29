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
	event := &Namespace{
		KubeResources: kr,
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "/:cluster", event.list),
	}
	event.Views = vs
	return event
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
