package resource

import (
	"github.com/kubespace/kubespace/pkg/kubernetes/config"
	"github.com/kubespace/kubespace/pkg/kubernetes/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var DaemonSetGVR = &schema.GroupVersionResource{
	Group:    "apps",
	Version:  "v1",
	Resource: "daemonsets",
}

type DaemonSet struct {
	*Resource
}

func NewDaemonSet(config *config.KubeConfig) *DaemonSet {
	p := &DaemonSet{}
	p.Resource = NewResource(config, types.DaemonsetType, DaemonSetGVR, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.PatchAction:  p.Patch,
		types.UpdateAction: p.Update,
		types.DeleteAction: p.Delete,
	}
	return p
}

type BuildDaemonSet struct {
	UID                    string            `json:"uid"`
	Name                   string            `json:"name"`
	Namespace              string            `json:"namespace"`
	DesiredNumberScheduled int32             `json:"desired_number_scheduled"`
	NumberReady            int32             `json:"number_ready"`
	ResourceVersion        string            `json:"resource_version"`
	Strategy               string            `json:"strategy"`
	Conditions             []string          `json:"conditions"`
	NodeSelector           map[string]string `json:"node_selector"`
	Created                metav1.Time       `json:"created"`
}

func (d *DaemonSet) ToBuildDaemonSet(ds *appsv1.DaemonSet) *BuildDaemonSet {
	if ds == nil {
		return nil
	}
	var conditions []string
	for _, c := range ds.Status.Conditions {
		if c.Status == corev1.ConditionTrue {
			conditions = append(conditions, string(c.Type))
		}
	}
	data := &BuildDaemonSet{
		UID:                    string(ds.UID),
		Name:                   ds.Name,
		Namespace:              ds.Namespace,
		DesiredNumberScheduled: ds.Status.DesiredNumberScheduled,
		NumberReady:            ds.Status.NumberReady,
		ResourceVersion:        ds.ResourceVersion,
		Conditions:             conditions,
		Strategy:               string(ds.Spec.UpdateStrategy.Type),
		NodeSelector:           ds.Spec.Template.Spec.NodeSelector,
		Created:                ds.CreationTimestamp,
	}

	return data
}

func (d *DaemonSet) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	ds := &appsv1.DaemonSet{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, ds); err != nil {
		return nil, err
	}
	return d.ToBuildDaemonSet(ds), nil
}
