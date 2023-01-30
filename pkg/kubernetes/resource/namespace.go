package resource

import (
	"github.com/kubespace/kubespace/pkg/kubernetes/config"
	"github.com/kubespace/kubespace/pkg/kubernetes/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var NamespaceGVR = &schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "namespaces",
}

type Namespace struct {
	*Resource
}

func NewNamespace(config *config.KubeConfig) *Namespace {
	p := &Namespace{}
	p.Resource = NewResource(config, types.PodType, NamespaceGVR, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction: p.List,
		types.GetAction:  p.Get,
	}
	return p
}

type BuildNamespace struct {
	UID             string            `json:"uid"`
	Name            string            `json:"name"`
	Created         metav1.Time       `json:"created"`
	Status          string            `json:"status"`
	Labels          map[string]string `json:"labels"`
	ResourceVersion string            `json:"resource_version"`
}

func (n *Namespace) ToBuildNamespace(ns *corev1.Namespace) *BuildNamespace {
	if ns == nil {
		return nil
	}
	return &BuildNamespace{
		UID:             string(ns.UID),
		Name:            ns.Name,
		Created:         ns.CreationTimestamp,
		Status:          string(ns.Status.Phase),
		Labels:          ns.Labels,
		ResourceVersion: ns.ResourceVersion,
	}
}

func (n *Namespace) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	ns := &corev1.Namespace{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, ns); err != nil {
		return nil, err
	}
	return n.ToBuildNamespace(ns), nil
}
