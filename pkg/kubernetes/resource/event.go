package resource

import (
	"context"
	"github.com/kubespace/kubespace/pkg/kubernetes/config"
	"github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

var EventGVR = &schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "namespaces",
}

type Event struct {
	*Resource
}

func NewEvent(config *config.KubeConfig) *Event {
	p := &Event{
		Resource: NewResource(config, types.PodType, EventGVR, nil),
	}
	p.actions = map[string]ActionHandle{
		types.ListAction: p.List,
		types.GetAction:  p.Get,
	}
	return p
}

type BuildEvent struct {
	UID             string                  `json:"uid"`
	Namespace       string                  `json:"namespace"`
	Reason          string                  `json:"reason"`
	Message         string                  `json:"message"`
	Type            string                  `json:"type"`
	Object          *corev1.ObjectReference `json:"object"`
	Source          *corev1.EventSource     `json:"source"`
	EventTime       metav1.Time             `json:"event_time"`
	Count           int32                   `json:"count"`
	ResourceVersion string                  `json:"resource_version"`
}

func (e *Event) ToBuildEvent(event *corev1.Event) *BuildEvent {
	if e == nil {
		return nil
	}
	eventTime := event.LastTimestamp
	if eventTime.IsZero() {
		eventTime = event.FirstTimestamp
	}
	if eventTime.IsZero() {
		eventTime = event.CreationTimestamp
	}
	eventData := &BuildEvent{
		UID:             string(event.UID),
		Namespace:       event.Namespace,
		Reason:          event.Reason,
		Message:         event.Message,
		Type:            event.Type,
		Object:          &event.InvolvedObject,
		Source:          &event.Source,
		EventTime:       eventTime,
		Count:           event.Count,
		ResourceVersion: event.ResourceVersion,
	}

	return eventData
}

func (e *Event) List(params interface{}) *utils.Response {
	query := &QueryParams{}
	if err := utils.ConvertTypeByJson(params, query); err != nil {
		return &utils.Response{Code: code.UnMarshalError, Msg: err.Error()}
	}
	listOptions, err := e.listOptionsFromQuery(query)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	fieldSet := fields.Set{}
	if query.Name != "" {
		fieldSet["involvedObject.name"] = query.Name
	}
	if query.Kind != "" {
		fieldSet["involvedObject.kind"] = query.Kind
	}
	if query.Namespace != "" {
		fieldSet["involvedObject.namespace"] = query.Namespace
	}
	listOptions.FieldSelector = fieldSet.String()
	objs, err := e.client.CoreV1().Events(query.Namespace).List(context.Background(), *listOptions)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	var nsRes []*BuildEvent
	for _, obj := range objs.Items {
		if query.Name == "" || strings.Contains(obj.ObjectMeta.Name, query.Name) {
			nsRes = append(nsRes, e.ToBuildEvent(&obj))
		}
	}
	return &utils.Response{Code: code.Success, Msg: "Success", Data: nsRes}
}
