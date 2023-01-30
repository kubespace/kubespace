package resource

import (
	"github.com/kubespace/kubespace/pkg/kubernetes/config"
	"github.com/kubespace/kubespace/pkg/kubernetes/types"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	RoleBindingGVR = &schema.GroupVersionResource{
		Group:    "rbac.authorization.k8s.io",
		Version:  "v1",
		Resource: "rolebindings",
	}
	RoleBindingV1Beta1GVR = &schema.GroupVersionResource{
		Group:    "rbac.authorization.k8s.io",
		Version:  "v1beta1",
		Resource: "rolebindings",
	}

	ClusterRoleBindingGVR = &schema.GroupVersionResource{
		Group:    "rbac.authorization.k8s.io",
		Version:  "v1",
		Resource: "clusterrolebindings",
	}
	ClusterRoleBindingV1Beta1GVR = &schema.GroupVersionResource{
		Group:    "rbac.authorization.k8s.io",
		Version:  "v1beta1",
		Resource: "clusterrolebindings",
	}
)

type RoleBinding struct {
	*Resource
}

func NewRoleBinding(config *config.KubeConfig) *RoleBinding {
	p := &RoleBinding{}
	gvr := RoleBindingGVR
	if !config.Client.VersionGreaterThan(types.ServerVersion17) {
		gvr = RoleBindingV1Beta1GVR
	}
	p.Resource = NewResource(config, types.RoleBindingType, gvr, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.DeleteAction: p.Delete,
		types.UpdateAction: p.Update,
	}
	return p
}

type BuildRoleBinding struct {
	UID             string           `json:"uid"`
	Kind            string           `json:"kind"`
	Name            string           `json:"name"`
	Namespace       string           `json:"namespace"`
	Subjects        []rbacv1.Subject `json:"subjects"`
	Role            rbacv1.RoleRef   `json:"role"`
	ResourceVersion string           `json:"resource_version"`
	Created         metav1.Time      `json:"created"`
}

func (s *RoleBinding) ToBuildRoleBinding(roleBinding *rbacv1.RoleBinding) *BuildRoleBinding {
	if roleBinding == nil {
		return nil
	}
	data := &BuildRoleBinding{
		UID:             string(roleBinding.UID),
		Name:            roleBinding.Name,
		Kind:            "RoleBinding",
		Subjects:        roleBinding.Subjects,
		Role:            roleBinding.RoleRef,
		Namespace:       roleBinding.Namespace,
		Created:         roleBinding.CreationTimestamp,
		ResourceVersion: roleBinding.ResourceVersion,
	}

	return data
}

func (s *RoleBinding) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	rb := &rbacv1.RoleBinding{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, rb); err != nil {
		return nil, err
	}
	return s.ToBuildRoleBinding(rb), nil
}

type ClusterRoleBinding struct {
	*Resource
}

func NewClusterRoleBinding(config *config.KubeConfig) *ClusterRoleBinding {
	p := &ClusterRoleBinding{}
	gvr := ClusterRoleBindingGVR
	if !config.Client.VersionGreaterThan(types.ServerVersion17) {
		gvr = ClusterRoleBindingV1Beta1GVR
	}
	p.Resource = NewResource(config, types.ClusterRoleBindingType, gvr, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.DeleteAction: p.Delete,
		types.UpdateAction: p.Update,
	}
	return p
}

func (s *ClusterRoleBinding) ToBuildClusterRoleBinding(roleBinding *rbacv1.ClusterRoleBinding) *BuildRoleBinding {
	if roleBinding == nil {
		return nil
	}
	data := &BuildRoleBinding{
		UID:             string(roleBinding.UID),
		Kind:            "ClusterRoleBinding",
		Name:            roleBinding.Name,
		Namespace:       roleBinding.Namespace,
		Subjects:        roleBinding.Subjects,
		Role:            roleBinding.RoleRef,
		Created:         roleBinding.CreationTimestamp,
		ResourceVersion: roleBinding.ResourceVersion,
	}

	return data
}

func (s *ClusterRoleBinding) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	rb := &rbacv1.ClusterRoleBinding{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, rb); err != nil {
		return nil, err
	}
	return s.ToBuildClusterRoleBinding(rb), nil
}
