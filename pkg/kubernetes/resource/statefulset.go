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

var StatefulSetGVR = &schema.GroupVersionResource{
	Group:    "apps",
	Version:  "v1",
	Resource: "statefulsets",
}

type StatefulSet struct {
	*Resource
}

func NewStatefulSet(config *config.KubeConfig) *StatefulSet {
	p := &StatefulSet{}
	p.Resource = NewResource(config, types.StatefulsetType, StatefulSetGVR, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.DeleteAction: p.Delete,
		types.PatchAction:  p.Patch,
		types.UpdateAction: p.Update,
	}
	return p
}

type BuildStatefulSet struct {
	UID             string      `json:"uid"`
	Name            string      `json:"name"`
	Namespace       string      `json:"namespace"`
	Replicas        int32       `json:"replicas"`
	StatusReplicas  int32       `json:"status_replicas"`
	ReadyReplicas   int32       `json:"ready_replicas"`
	UpdatedReplicas int32       `json:"updated_replicas"`
	ResourceVersion string      `json:"resource_version"`
	Strategy        string      `json:"strategy"`
	Conditions      []string    `json:"conditions"`
	Created         metav1.Time `json:"created"`
}

func (s *StatefulSet) ToBuildStatefulSet(ss *appsv1.StatefulSet) *BuildStatefulSet {
	if ss == nil {
		return nil
	}
	var conditions []string
	for _, c := range ss.Status.Conditions {
		if c.Status == corev1.ConditionTrue {
			conditions = append(conditions, string(c.Type))
		}
	}
	data := &BuildStatefulSet{
		UID:             string(ss.UID),
		Name:            ss.Name,
		Namespace:       ss.Namespace,
		Replicas:        *ss.Spec.Replicas,
		StatusReplicas:  ss.Status.Replicas,
		ReadyReplicas:   ss.Status.ReadyReplicas,
		UpdatedReplicas: ss.Status.UpdatedReplicas,
		ResourceVersion: ss.ResourceVersion,
		Strategy:        string(ss.Spec.UpdateStrategy.Type),
		Conditions:      conditions,
		Created:         ss.CreationTimestamp,
	}

	return data
}

func (s *StatefulSet) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	ds := &appsv1.StatefulSet{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, ds); err != nil {
		return nil, err
	}
	return s.ToBuildStatefulSet(ds), nil
}
