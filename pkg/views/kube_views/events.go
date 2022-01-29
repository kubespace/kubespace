package kube_views

import (
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"net/http"
)

type Event struct {
	Views []*views.View
	*kube_resource.KubeResources
}

func NewEvent(kr *kube_resource.KubeResources) *Event {
	event := &Event{
		KubeResources: kr,
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "/:cluster", event.list),
	}
	event.Views = vs
	return event
}

func (e *Event) list(c *views.Context) *utils.Response {
	var ser serializers.EventListSerializers
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	reqParams := map[string]interface{}{}
	if ser.Namespace != "" {
		reqParams["namespace"] = ser.Namespace
	}
	if ser.Name != "" {
		reqParams["name"] = ser.Name
	}
	if ser.UID != "" {
		reqParams["uid"] = ser.UID
	}
	if ser.Kind != "" {
		reqParams["kind"] = ser.Kind
	}
	return e.Event.List(c.Param("cluster"), reqParams)
}
