package kube_views

import (
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"net/http"
)

type Crd struct {
	Views []*views.View
	*kube_resource.KubeResources
}

func NewCrd(kr *kube_resource.KubeResources) *Crd {
	crd := &Crd{
		KubeResources: kr,
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "/:cluster/cr", crd.listCR),
		views.NewView(http.MethodGet, "/:cluster/cr/:name", crd.getCR),
		views.NewView(http.MethodDelete, "/:cluster/cr/:name", crd.deleteCR),
		views.NewView(http.MethodGet, "/:cluster", crd.list),
		views.NewView(http.MethodGet, "/:cluster/:name", crd.get),
	}
	crd.Views = vs
	return crd
}

func (d *Crd) list(c *views.Context) *utils.Response {
	return d.Crd.List(c.Param("cluster"), struct{}{})
}

func (d *Crd) get(c *views.Context) *utils.Response {
	var ser serializers.GetSerializers
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name":   c.Param("name"),
		"output": ser.Output,
	}
	return d.Crd.Get(c.Param("cluster"), reqParams)
}

func (d *Crd) listCR(c *views.Context) *utils.Response {
	var ser serializers.CRSerializers
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"group":    ser.Group,
		"version":  ser.Version,
		"resource": ser.Resource,
	}
	return d.Cr.List(c.Param("cluster"), reqParams)
}

func (d *Crd) getCR(c *views.Context) *utils.Response {
	var ser serializers.CRSerializers
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"group":     ser.Group,
		"version":   ser.Version,
		"resource":  ser.Resource,
		"namespace": ser.Namespace,
		"name":      c.Param("name"),
		"output":    ser.Output,
	}
	return d.Cr.Get(c.Param("cluster"), reqParams)
}

func (d *Crd) deleteCR(c *views.Context) *utils.Response {
	var ser serializers.CRSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"group":     ser.Group,
		"version":   ser.Version,
		"resource":  ser.Resource,
		"namespace": ser.Namespace,
		"name":      c.Param("name"),
	}
	return d.Cr.Delete(c.Param("cluster"), reqParams)
}
