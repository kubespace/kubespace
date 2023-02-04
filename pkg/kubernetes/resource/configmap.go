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

var ConfigmapGVR = &schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "configmaps",
}

type Configmap struct {
	*Resource
}

func NewConfigmap(config *config.KubeConfig) *Configmap {
	p := &Configmap{}
	p.Resource = NewResource(config, types.ConfigMapType, ConfigmapGVR, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.DeleteAction: p.Delete,
		types.UpdateAction: p.Update,
	}
	return p
}

type BuildConfigMap struct {
	Name       string            `json:"name"`
	NameSpace  string            `json:"namespace"`
	Keys       []string          `json:"keys"`
	Labels     map[string]string `json:"labels"`
	CreateTime metav1.Time       `json:"create_time"`
	Data       map[string]string `json:"data"`
}

func (c *Configmap) ToBuildConfigMap(cm *corev1.ConfigMap) *BuildConfigMap {
	if cm == nil {
		return nil
	}

	cmData := &BuildConfigMap{
		Name:       cm.Name,
		NameSpace:  cm.Namespace,
		Labels:     cm.Labels,
		CreateTime: cm.CreationTimestamp,
		Data:       cm.Data,
	}

	keys := make([]string, 0, len(cm.Data))
	for k, _ := range cm.Data {
		keys = append(keys, k)
	}
	cmData.Keys = keys
	return cmData
}

func (c *Configmap) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	cm := &corev1.ConfigMap{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, cm); err != nil {
		return nil, err
	}
	return c.ToBuildConfigMap(cm), nil
}
