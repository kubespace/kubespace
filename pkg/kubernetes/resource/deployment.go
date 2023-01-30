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

var DeploymentGVR = &schema.GroupVersionResource{
	Group:    "apps",
	Version:  "v1",
	Resource: "deployments",
}

type Deployment struct {
	*Resource
}

func NewDeployment(config *config.KubeConfig) *Deployment {
	p := &Deployment{}
	p.Resource = NewResource(config, types.DeploymentType, DeploymentGVR, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.UpdateAction: p.Update,
		types.DeleteAction: p.Delete,
		types.PatchAction:  p.Patch,
	}
	return p
}

type BuildDeployment struct {
	UID                 string      `json:"uid"`
	Name                string      `json:"name"`
	Namespace           string      `json:"namespace"`
	Replicas            int32       `json:"replicas"`
	StatusReplicas      int32       `json:"status_replicas"`
	ReadyReplicas       int32       `json:"ready_replicas"`
	UpdatedReplicas     int32       `json:"updated_replicas"`
	UnavailableReplicas int32       `json:"unavailable_replicas"`
	AvailableReplicas   int32       `json:"available_replicas"`
	ResourceVersion     string      `json:"resource_version"`
	Strategy            string      `json:"strategy"`
	Conditions          []string    `json:"conditions"`
	Created             metav1.Time `json:"created"`
}

func (d *Deployment) ToBuildDeployment(dp *appsv1.Deployment) *BuildDeployment {
	if dp == nil {
		return nil
	}
	var conditions []string
	for _, c := range dp.Status.Conditions {
		if c.Status == corev1.ConditionTrue {
			conditions = append(conditions, string(c.Type))
		}
	}
	dpData := &BuildDeployment{
		UID:                 string(dp.UID),
		Name:                dp.Name,
		Namespace:           dp.Namespace,
		Replicas:            *dp.Spec.Replicas,
		StatusReplicas:      dp.Status.Replicas,
		ReadyReplicas:       dp.Status.ReadyReplicas,
		UpdatedReplicas:     dp.Status.UpdatedReplicas,
		UnavailableReplicas: dp.Status.UnavailableReplicas,
		AvailableReplicas:   dp.Status.AvailableReplicas,
		ResourceVersion:     dp.ResourceVersion,
		Strategy:            string(dp.Spec.Strategy.Type),
		Conditions:          conditions,
		Created:             dp.CreationTimestamp,
	}

	return dpData
}

func (d *Deployment) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	deploy := &appsv1.Deployment{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, deploy); err != nil {
		return nil, err
	}
	return d.ToBuildDeployment(deploy), nil
}
