package kube_views

import (
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"net/http"
)

type ConfigMap struct {
	Views []*views.View
	*kube_resource.KubeResources
}

func NewConfigMap(kr *kube_resource.KubeResources) *ConfigMap {
	cm := &ConfigMap{
		KubeResources: kr,
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "/:cluster/:namespace/:name", cm.get),
		views.NewView(http.MethodGet, "/:cluster", cm.list),
		views.NewView(http.MethodPost, "/:cluster/delete", cm.delete),
		views.NewView(http.MethodPost, "/:cluster/update/:namespace/:name", cm.updateYaml),
		views.NewView(http.MethodPost, "/:cluster/update_obj/:namespace/:name", cm.updateObj),
	}
	cm.Views = vs
	return cm
}

func (cm *ConfigMap) list(c *views.Context) *utils.Response {
	var ser serializers.ListSerializers
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name":      ser.Name,
		"namespace": ser.Namespace,
	}
	return cm.ConfigMap.List(c.Param("cluster"), reqParams)
}

func (cm *ConfigMap) get(c *views.Context) *utils.Response {
	var ser serializers.GetSerializers
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name":      c.Param("name"),
		"namespace": c.Param("namespace"),
		"output":    ser.Output,
	}
	return cm.ConfigMap.Get(c.Param("cluster"), reqParams)
}

func (cm *ConfigMap) delete(c *views.Context) *utils.Response {
	var ser serializers.DeleteSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return cm.ConfigMap.Delete(c.Param("cluster"), ser)
}

func (cm *ConfigMap) updateYaml(c *views.Context) *utils.Response {
	var ser serializers.UpdateSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name":      c.Param("name"),
		"namespace": c.Param("namespace"),
		"yaml":      ser.Yaml,
	}
	return cm.ConfigMap.UpdateYaml(c.Param("cluster"), reqParams)
}

func (cm *ConfigMap) updateObj(c *views.Context) *utils.Response {
	var ser serializers.UpdateMapSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"data": ser.Data,
	}

	return cm.ConfigMap.UpdateObj(c.Param("cluster"), reqParams)
}
