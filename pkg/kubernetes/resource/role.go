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
	RoleGVR = &schema.GroupVersionResource{
		Group:    "rbac.authorization.k8s.io",
		Version:  "v1",
		Resource: "roles",
	}
	RoleV1Beta1GVR = &schema.GroupVersionResource{
		Group:    "rbac.authorization.k8s.io",
		Version:  "v1beta1",
		Resource: "roles",
	}

	ClusterRoleGVR = &schema.GroupVersionResource{
		Group:    "rbac.authorization.k8s.io",
		Version:  "v1",
		Resource: "clusterroles",
	}
	ClusterRoleV1Beta1GVR = &schema.GroupVersionResource{
		Group:    "rbac.authorization.k8s.io",
		Version:  "v1beta1",
		Resource: "clusterroles",
	}
)

type Role struct {
	*Resource
}

func NewRole(config *config.KubeConfig) *Role {
	p := &Role{}
	gvr := RoleGVR
	if !config.Client.VersionGreaterThan(types.ServerVersion17) {
		gvr = RoleV1Beta1GVR
	}
	p.Resource = NewResource(config, types.RoleType, gvr, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.DeleteAction: p.Delete,
		types.UpdateAction: p.Update,
	}
	return p
}

type BuildRole struct {
	UID             string      `json:"uid"`
	Kind            string      `json:"kind"`
	Name            string      `json:"name"`
	Namespace       string      `json:"namespace"`
	ResourceVersion string      `json:"resource_version"`
	Created         metav1.Time `json:"created"`
}

func (s *Role) ToBuildRole(role *rbacv1.Role) *BuildRole {
	if role == nil {
		return nil
	}
	data := &BuildRole{
		UID:             string(role.UID),
		Name:            role.Name,
		Kind:            "Role",
		Namespace:       role.Namespace,
		Created:         role.CreationTimestamp,
		ResourceVersion: role.ResourceVersion,
	}

	return data
}

func (s *Role) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	role := &rbacv1.Role{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, role); err != nil {
		return nil, err
	}
	return s.ToBuildRole(role), nil
}

type ClusterRole struct {
	*Resource
}

func NewClusterRole(config *config.KubeConfig) *ClusterRole {
	p := &ClusterRole{}
	gvr := ClusterRoleGVR
	if !config.Client.VersionGreaterThan(types.ServerVersion17) {
		gvr = ClusterRoleV1Beta1GVR
	}
	p.Resource = NewResource(config, types.ClusterRoleType, gvr, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.DeleteAction: p.Delete,
		types.UpdateAction: p.Update,
	}
	return p
}

func (s *ClusterRole) ToBuildClusterRole(role *rbacv1.ClusterRole) *BuildRole {
	if role == nil {
		return nil
	}
	data := &BuildRole{
		UID:             string(role.UID),
		Kind:            "ClusterRole",
		Name:            role.Name,
		Namespace:       role.Namespace,
		Created:         role.CreationTimestamp,
		ResourceVersion: role.ResourceVersion,
	}

	return data
}

func (s *ClusterRole) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	role := &rbacv1.ClusterRole{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, role); err != nil {
		return nil, err
	}
	return s.ToBuildClusterRole(role), nil
}
