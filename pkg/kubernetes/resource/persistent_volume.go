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

var PersistentVolumeGVR = &schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "persistentvolumes",
}

type PersistentVolume struct {
	*Resource
}

func NewPersistentVolume(config *config.KubeConfig) *PersistentVolume {
	p := &PersistentVolume{}
	p.Resource = NewResource(config, types.PersistentVolumeType, PersistentVolumeGVR, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.DeleteAction: p.Delete,
		types.UpdateAction: p.Update,
	}
	return p
}

type BuildPersistentVolume struct {
	UID            string                           `json:"uid"`
	Name           string                           `json:"name"`
	Status         string                           `json:"status"`
	Claim          string                           `json:"claim"`
	ClaimNamespace string                           `json:"claim_namespace"`
	StorageClass   string                           `json:"storage_class"`
	Capacity       string                           `json:"capacity"`
	AccessModes    []corev1.PersistentVolumeAccessMode  `json:"access_modes"`
	CreateTime     metav1.Time                      `json:"create_time"`
	ReclaimPolicy  corev1.PersistentVolumeReclaimPolicy `json:"reclaim_policy"`
}

func (p *PersistentVolume) ToBuildPersistentVolume(pv *corev1.PersistentVolume) *BuildPersistentVolume {
	if pv == nil {
		return nil
	}
	var volumeSize, claimName, claimNamespace string
	if size, ok := pv.Spec.Capacity["storage"]; !ok {
		volumeSize = ""
	} else {
		volumeSize = size.String()
	}
	if (pv.Spec.ClaimRef) != nil {
		claimName = pv.Spec.ClaimRef.Name
		claimNamespace = pv.Spec.ClaimRef.Namespace
	}
	pvData := &BuildPersistentVolume{
		UID:            string(pv.UID),
		Name:           pv.Name,
		Status:         string(pv.Status.Phase),
		StorageClass:   pv.Spec.StorageClassName,
		Capacity:       volumeSize,
		Claim:          claimName,
		ClaimNamespace: claimNamespace,
		AccessModes:    pv.Spec.AccessModes,
		ReclaimPolicy:  pv.Spec.PersistentVolumeReclaimPolicy,
		CreateTime:     pv.CreationTimestamp,
	}

	return pvData
}

func (p *PersistentVolume) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	pv := &corev1.PersistentVolume{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, pv); err != nil {
		return nil, err
	}
	return p.ToBuildPersistentVolume(pv), nil
}
