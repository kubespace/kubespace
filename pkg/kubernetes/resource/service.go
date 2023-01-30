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

var ServiceGVR = &schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "services",
}

var EndpointsGVR = &schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "endpoints",
}

type Service struct {
	*Resource
}

func NewService(config *config.KubeConfig) *Service {
	p := &Service{}
	p.Resource = NewResource(config, types.ServiceType, ServiceGVR, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.DeleteAction: p.Delete,
		types.UpdateAction: p.Update,
	}
	return p
}

type BuildService struct {
	UID             string               `json:"uid"`
	Name            string               `json:"name"`
	Namespace       string               `json:"namespace"`
	Type            string               `json:"type"`
	ClusterIP       string               `json:"cluster_ip"`
	Ports           []corev1.ServicePort `json:"ports"`
	ExternalIP      []string             `json:"external_ip"`
	Selector        map[string]string    `json:"selector"`
	ResourceVersion string               `json:"resource_version"`
	Created         metav1.Time          `json:"created"`
}

func (s *Service) ToBuildService(service *corev1.Service) *BuildService {
	if service == nil {
		return nil
	}
	data := &BuildService{
		UID:             string(service.UID),
		Name:            service.Name,
		Namespace:       service.Namespace,
		Type:            string(service.Spec.Type),
		ClusterIP:       service.Spec.ClusterIP,
		Ports:           service.Spec.Ports,
		ExternalIP:      service.Spec.ExternalIPs,
		Selector:        service.Spec.Selector,
		Created:         service.CreationTimestamp,
		ResourceVersion: service.ResourceVersion,
	}

	return data
}

func (s *Service) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	se := &corev1.Service{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, se); err != nil {
		return nil, err
	}
	return s.ToBuildService(se), nil
}

type Endpoints struct {
	*Resource
}

func NewEndpoints(config *config.KubeConfig) *Endpoints {
	p := &Endpoints{}
	p.Resource = NewResource(config, types.EndpointType, EndpointsGVR, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.DeleteAction: p.Delete,
		types.UpdateAction: p.Update,
	}
	return p
}

type BuildEndpoints struct {
	UID             string                  `json:"uid"`
	Name            string                  `json:"name"`
	Namespace       string                  `json:"namespace"`
	Subsets         []corev1.EndpointSubset `json:"subsets"`
	Created         metav1.Time             `json:"created"`
	ResourceVersion string                  `json:"resource_version"`
}

func (e *Endpoints) ToBuildEndpoints(endpoints *corev1.Endpoints) *BuildEndpoints {
	if endpoints == nil {
		return nil
	}
	data := &BuildEndpoints{
		UID:             string(endpoints.UID),
		Name:            endpoints.Name,
		Namespace:       endpoints.Namespace,
		Subsets:         endpoints.Subsets,
		Created:         endpoints.CreationTimestamp,
		ResourceVersion: endpoints.ResourceVersion,
	}

	return data
}

func (e *Endpoints) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	ep := &corev1.Endpoints{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, ep); err != nil {
		return nil, err
	}
	return e.ToBuildEndpoints(ep), nil
}
