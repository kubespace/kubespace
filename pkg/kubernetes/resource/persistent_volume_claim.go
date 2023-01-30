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

var PersistentVolumeClaimGVR = &schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "persistentvolumeclaims",
}

type PersistentVolumeClaim struct {
	*Resource
}

func NewPersistentVolumeClaim(config *config.KubeConfig) *PersistentVolumeClaim {
	p := &PersistentVolumeClaim{}
	p.Resource = NewResource(config, types.PersistentVolumeClaimType, PersistentVolumeClaimGVR, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.DeleteAction: p.Delete,
		types.UpdateAction: p.Update,
	}
	return p
}

type BuildPersistentVolumeClaim struct {
	UID           string                               `json:"uid"`
	Name          string                               `json:"name"`
	Namespace     string                               `json:"namespace"`
	Status        string                               `json:"status"`
	StorageClass  *string                              `json:"storage_class"`
	Capacity      string                               `json:"capacity"`
	AccessModes   []corev1.PersistentVolumeAccessMode  `json:"access_modes"`
	CreateTime    metav1.Time                          `json:"create_time"`
	ReclaimPolicy corev1.PersistentVolumeReclaimPolicy `json:"reclaim_policy"`
}

func (p *PersistentVolumeClaim) ToBuildPersistentVolumeClaim(pvc *corev1.PersistentVolumeClaim) *BuildPersistentVolumeClaim {
	if pvc == nil {
		return nil
	}
	var volumeSize string
	if size, ok := pvc.Spec.Resources.Requests["storage"]; !ok {
		volumeSize = ""
	} else {
		volumeSize = size.String()
	}

	storageClass := pvc.Spec.StorageClassName

	pvcData := &BuildPersistentVolumeClaim{
		UID:          string(pvc.UID),
		Name:         pvc.Name,
		Status:       string(pvc.Status.Phase),
		AccessModes:  pvc.Spec.AccessModes,
		CreateTime:   pvc.CreationTimestamp,
		Namespace:    pvc.Namespace,
		StorageClass: storageClass,
		Capacity:     volumeSize,
	}

	return pvcData
}

func (p *PersistentVolumeClaim) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	pvc := &corev1.PersistentVolumeClaim{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, pvc); err != nil {
		return nil, err
	}
	return p.ToBuildPersistentVolumeClaim(pvc), nil
}
