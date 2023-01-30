package resource

import (
	"github.com/kubespace/kubespace/pkg/kubernetes/config"
	"github.com/kubespace/kubespace/pkg/kubernetes/types"
	networkv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var NetworkPolicyGVR = &schema.GroupVersionResource{
	Group:    "networking",
	Version:  "v1",
	Resource: "networkpolicies",
}

type NetworkPolicy struct {
	*Resource
}

func NewNetworkPolicy(config *config.KubeConfig) *NetworkPolicy {
	p := &NetworkPolicy{}
	p.Resource = NewResource(config, types.NetworkPolicyType, NetworkPolicyGVR, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.DeleteAction: p.Delete,
		types.UpdateAction: p.Update,
	}
	return p
}

type BuildNetworkPolicy struct {
	UID             string                 `json:"uid"`
	Name            string                 `json:"name"`
	Namespace       string                 `json:"namespace"`
	PolicyTypes     []networkv1.PolicyType `json:"policy_types"`
	Created         metav1.Time            `json:"created"`
	ResourceVersion string                 `json:"resource_version"`
}

func (n *NetworkPolicy) ToBuildNetworkPolicy(networkpolicy *networkv1.NetworkPolicy) *BuildNetworkPolicy {
	if networkpolicy == nil {
		return nil
	}
	data := &BuildNetworkPolicy{
		UID:             string(networkpolicy.UID),
		Name:            networkpolicy.Name,
		Namespace:       networkpolicy.Namespace,
		PolicyTypes:     networkpolicy.Spec.PolicyTypes,
		Created:         networkpolicy.CreationTimestamp,
		ResourceVersion: networkpolicy.ResourceVersion,
	}

	return data
}

func (n *NetworkPolicy) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	np := &networkv1.NetworkPolicy{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, np); err != nil {
		return nil, err
	}
	return n.ToBuildNetworkPolicy(np), nil
}
