package kube_views

import (
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"net/http"
)

type Node struct {
	Views []*views.View
	*kube_resource.KubeResources
}

func NewNode(kr *kube_resource.KubeResources) *Node {
	node := &Node{
		KubeResources: kr,
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "/:cluster", node.list),
		views.NewView(http.MethodGet, "/:cluster/:name", node.get),
		views.NewView(http.MethodPost, "/:cluster/delete", node.delete),
		views.NewView(http.MethodPost, "/:cluster/update/:name", node.updateYaml),
	}
	node.Views = vs
	return node
}

func (n *Node) list(c *views.Context) *utils.Response {
	return n.Node.List(c.Param("cluster"), struct{}{})
}

func (n *Node) get(c *views.Context) *utils.Response {
	var ser serializers.GetSerializers
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name":   c.Param("name"),
		"output": ser.Output,
	}
	return n.Node.Get(c.Param("cluster"), reqParams)
}

func (n *Node) delete(c *views.Context) *utils.Response {
	var ser serializers.DeleteSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return n.Node.Delete(c.Param("cluster"), ser)
}

func (n *Node) updateYaml(c *views.Context) *utils.Response {
	var ser serializers.UpdateSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{
		"name": c.Param("name"),
		"yaml": ser.Yaml,
	}
	return n.Node.UpdateYaml(c.Param("cluster"), reqParams)
}
