package kube_views

import (
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"net/http"
)

type Secret struct {
	Views []*views.View
	*kube_resource.KubeResources
}

func NewSecret(kr *kube_resource.KubeResources) *Secret {
	s := &Secret{
		KubeResources: kr,
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "/:cluster/:namespace/:name", s.get),
		views.NewView(http.MethodGet, "/:cluster", s.list),
		views.NewView(http.MethodPost, "/:cluster/delete", s.delete),
		views.NewView(http.MethodPost, "/:cluster/update/:namespace/:name", s.updateYaml),
	}
	s.Views = vs
	return s
}

func (s *Secret) list(c *views.Context) *utils.Response {
	var ser serializers.ListSerializers
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name":      ser.Name,
		"namespace": ser.Namespace,
		"labels":    ser.Labels,
	}
	return s.Secret.List(c.Param("cluster"), reqParams)
}

func (s *Secret) get(c *views.Context) *utils.Response {
	var ser serializers.GetSerializers
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name":      c.Param("name"),
		"namespace": c.Param("namespace"),
		"output":    ser.Output,
	}
	return s.Secret.Get(c.Param("cluster"), reqParams)
}

func (s *Secret) delete(c *views.Context) *utils.Response {
	var ser serializers.DeleteSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return s.Secret.Delete(c.Param("cluster"), ser)
}

func (s *Secret) updateYaml(c *views.Context) *utils.Response {
	var ser serializers.UpdateSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name":      c.Param("name"),
		"namespace": c.Param("namespace"),
		"yaml":      ser.Yaml,
	}
	return s.Secret.UpdateYaml(c.Param("cluster"), reqParams)
}
