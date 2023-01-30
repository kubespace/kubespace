package resource

import (
	"github.com/kubespace/kubespace/pkg/kubernetes/config"
	"github.com/kubespace/kubespace/pkg/kubernetes/types"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var StorageClassGVR = &schema.GroupVersionResource{
	Group:    "storage.k8s.io",
	Version:  "v1",
	Resource: "storageclasses",
}
var StorageClassV1Beta1GVR = &schema.GroupVersionResource{
	Group:    "storage.k8s.io",
	Version:  "v1beta1",
	Resource: "storageclasses",
}

type StorageClass struct {
	*Resource
}

func NewStorageClass(config *config.KubeConfig) *StorageClass {
	p := &StorageClass{}
	gvr := StorageClassV1Beta1GVR
	if config.Client.VersionGreaterThan(types.ServerVersion19) {
		gvr = StorageClassGVR
	}
	p.Resource = NewResource(config, types.StorageClassType, gvr, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.DeleteAction: p.Delete,
		types.UpdateAction: p.Update,
	}
	return p
}

type BuildStorageClass struct {
	UID               string                                `json:"uid"`
	Name              string                                `json:"name"`
	CreateTime        metav1.Time                           `json:"create_time"`
	Provisioner       string                                `json:"provisioner"`
	ReclaimPolicy     *corev1.PersistentVolumeReclaimPolicy `json:"reclaim_policy"`
	VolumeBindingMode string                                `json:"binding_mode"`
	ResourceVersion   string                                `json:"resource_version"`
}

func (s *StorageClass) ToBuildStorageClass(sc *storagev1.StorageClass) *BuildStorageClass {
	if sc == nil {
		return nil
	}
	bindMode := ""
	if sc.VolumeBindingMode != nil {
		bindMode = string(*sc.VolumeBindingMode)
	}

	data := &BuildStorageClass{
		UID:               string(sc.UID),
		Name:              sc.Name,
		CreateTime:        sc.CreationTimestamp,
		Provisioner:       sc.Provisioner,
		ReclaimPolicy:     sc.ReclaimPolicy,
		VolumeBindingMode: bindMode,
		ResourceVersion:   sc.ResourceVersion,
	}

	return data
}

func (s *StorageClass) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	sc := &storagev1.StorageClass{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, sc); err != nil {
		return nil, err
	}
	return s.ToBuildStorageClass(sc), nil
}
