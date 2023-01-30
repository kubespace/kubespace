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

var ServiceAccountGVR = &schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "serviceaccounts",
}

type ServiceAccount struct {
	*Resource
}

func NewServiceAccount(config *config.KubeConfig) *ServiceAccount {
	p := &ServiceAccount{}
	p.Resource = NewResource(config, types.ServiceAccountType, ServiceAccountGVR, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.DeleteAction: p.Delete,
		types.UpdateAction: p.Update,
	}
	return p
}

type BuildServiceAccount struct {
	UID             string                   `json:"uid"`
	Name            string                   `json:"name"`
	Namespace       string                   `json:"namespace"`
	ResourceVersion string                   `json:"resource_version"`
	Secrets         []corev1.ObjectReference `json:"secrets"`
	Created         metav1.Time              `json:"created"`
}

func (s *ServiceAccount) ToBuildServiceAccount(serviceAccount *corev1.ServiceAccount) *BuildServiceAccount {
	if serviceAccount == nil {
		return nil
	}
	data := &BuildServiceAccount{
		UID:             string(serviceAccount.UID),
		Name:            serviceAccount.Name,
		Namespace:       serviceAccount.Namespace,
		Secrets:         serviceAccount.Secrets,
		Created:         serviceAccount.CreationTimestamp,
		ResourceVersion: serviceAccount.ResourceVersion,
	}

	return data
}

func (s *ServiceAccount) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	se := &corev1.ServiceAccount{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, se); err != nil {
		return nil, err
	}
	return s.ToBuildServiceAccount(se), nil
}
