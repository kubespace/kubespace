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

var SecretGVR = &schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "secrets",
}

type Secret struct {
	*Resource
}

func NewSecret(config *config.KubeConfig) *Secret {
	p := &Secret{}
	p.Resource = NewResource(config, types.SecretType, SecretGVR, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.DeleteAction: p.Delete,
		types.UpdateAction: p.Update,
	}
	return p
}

type BuildSecret struct {
	Name       string            `json:"name"`
	NameSpace  string            `json:"namespace"`
	Keys       []string          `json:"keys"`
	Labels     map[string]string `json:"labels"`
	CreateTime metav1.Time       `json:"create_time"`
	Type       corev1.SecretType `json:"type"`
	Data       map[string][]byte `json:"data"`
}

func (s *Secret) ToBuildSecret(se *corev1.Secret) *BuildSecret {
	if se == nil {
		return nil
	}

	sData := &BuildSecret{
		Name:       se.Name,
		NameSpace:  se.Namespace,
		Labels:     se.Labels,
		Type:       se.Type,
		CreateTime: se.CreationTimestamp,
		Data:       se.Data,
	}
	return sData
}

func (s *Secret) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	se := &corev1.Secret{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, se); err != nil {
		return nil, err
	}
	// don't list helm release secret
	if se.Type == "helm.sh/release.v1" {
		return nil, nil
	}
	return s.ToBuildSecret(se), nil
}
