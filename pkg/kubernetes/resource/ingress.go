package resource

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/kubernetes/config"
	"github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/utils"
	extv1betav1 "k8s.io/api/extensions/v1beta1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	NetworkIngressGVR = &schema.GroupVersionResource{
		Group:    "networking.k8s.io",
		Version:  "v1",
		Resource: "ingresses",
	}

	ExtensionIngressGVR = &schema.GroupVersionResource{
		Group:    "extensions",
		Version:  "v1beta1",
		Resource: "ingresses",
	}
)

type Ingress struct {
	*Resource
}

func NewIngress(config *config.KubeConfig) *Ingress {
	p := &Ingress{}
	gvr := ExtensionIngressGVR
	if config.Client.VersionGreaterThan(types.ServerVersion19) {
		gvr = NetworkIngressGVR
	}
	p.Resource = NewResource(config, types.IngressType, gvr, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.DeleteAction: p.Delete,
		types.UpdateAction: p.Update,
	}
	return p
}

type BuildIngress struct {
	UID             string                      `json:"uid"`
	Name            string                      `json:"name"`
	Namespace       string                      `json:"namespace"`
	Backend         *extv1betav1.IngressBackend `json:"backend"`
	TLS             []extv1betav1.IngressTLS    `json:"tls"`
	Rules           []extv1betav1.IngressRule   `json:"rules"`
	Created         metav1.Time                 `json:"created"`
	ResourceVersion string                      `json:"resource_version"`
}

func (i *Ingress) ToBuildIngress(ingress *extv1betav1.Ingress) *BuildIngress {
	if ingress == nil {
		return nil
	}
	data := &BuildIngress{
		UID:             string(ingress.UID),
		Name:            ingress.Name,
		Namespace:       ingress.Namespace,
		Backend:         ingress.Spec.Backend,
		TLS:             ingress.Spec.TLS,
		Rules:           ingress.Spec.Rules,
		Created:         ingress.CreationTimestamp,
		ResourceVersion: ingress.ResourceVersion,
	}

	return data
}

type BuildNewIngress struct {
	UID             string                       `json:"uid"`
	Name            string                       `json:"name"`
	Namespace       string                       `json:"namespace"`
	Backend         *networkingv1.IngressBackend `json:"backend"`
	TLS             []networkingv1.IngressTLS    `json:"tls"`
	Rules           []networkingv1.IngressRule   `json:"rules"`
	Created         metav1.Time                  `json:"created"`
	ResourceVersion string                       `json:"resource_version"`
}

func (i *Ingress) ToBuildNewIngress(ingress *networkingv1.Ingress) *BuildNewIngress {
	if ingress == nil {
		return nil
	}
	data := &BuildNewIngress{
		UID:             string(ingress.UID),
		Name:            ingress.Name,
		Namespace:       ingress.Namespace,
		Backend:         ingress.Spec.DefaultBackend,
		TLS:             ingress.Spec.TLS,
		Rules:           ingress.Spec.Rules,
		Created:         ingress.CreationTimestamp,
		ResourceVersion: ingress.ResourceVersion,
	}

	return data
}

func (i *Ingress) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	if i.gvr.Group == "networking" {
		ing := &networkingv1.Ingress{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, ing); err != nil {
			return nil, err
		}
		return i.ToBuildNewIngress(ing), nil
	} else {
		ing := &extv1betav1.Ingress{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, ing); err != nil {
			return nil, err
		}
		return i.ToBuildIngress(ing), nil
	}
}

func (i *Ingress) List(params interface{}) *utils.Response {
	resp := i.Resource.List(params)
	if !resp.IsSuccess() {
		return resp
	}
	data := map[string]interface{}{
		"ingresses": resp.Data,
		"group":     i.gvr.Group,
	}
	return &utils.Response{Code: code.Success, Msg: "Success", Data: data}
}
