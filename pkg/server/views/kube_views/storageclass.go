package kube_views

import (
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"net/http"
)

type StorageClass struct {
	Views []*views.View
	*kube_resource.KubeResources
}

func NewStorageClass(kr *kube_resource.KubeResources) *StorageClass {
	s := &StorageClass{
		KubeResources: kr,
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "/:cluster/:name", s.get),
		views.NewView(http.MethodGet, "/:cluster", s.list),
		views.NewView(http.MethodPost, "/:cluster/delete", s.delete),
		views.NewView(http.MethodPost, "/:cluster/update/:name", s.updateYaml),
	}
	s.Views = vs
	return s
}

func (s *StorageClass) list(c *views.Context) *utils.Response {
	var ser serializers.ListSerializers
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name":      ser.Name,
		"namespace": ser.Namespace,
	}
	return s.StorageClass.List(c.Param("cluster"), reqParams)
}

func (s *StorageClass) get(c *views.Context) *utils.Response {
	var ser serializers.GetSerializers
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name":      c.Param("name"),
		"namespace": c.Param("namespace"),
		"output":    ser.Output,
	}
	return s.StorageClass.Get(c.Param("cluster"), reqParams)
}

func (s *StorageClass) delete(c *views.Context) *utils.Response {
	var ser serializers.DeleteSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return s.StorageClass.Delete(c.Param("cluster"), ser)
}

func (s *StorageClass) updateYaml(c *views.Context) *utils.Response {
	var ser serializers.UpdateSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name": c.Param("name"),
		"yaml": ser.Yaml,
	}
	return s.StorageClass.UpdateYaml(c.Param("cluster"), reqParams)
}
