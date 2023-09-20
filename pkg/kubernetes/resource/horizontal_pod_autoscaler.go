package resource

import (
	"github.com/kubespace/kubespace/pkg/kubernetes/config"
	"github.com/kubespace/kubespace/pkg/kubernetes/types"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var HorizontalPodAutoscalerGVR = &schema.GroupVersionResource{
	Group:    "autoscaling",
	Version:  "v2beta1",
	Resource: "horizontalpodautoscalers",
}

var HorizontalPodAutoscalerV1GVR = &schema.GroupVersionResource{
	Group:    "autoscaling",
	Version:  "v1",
	Resource: "horizontalpodautoscalers",
}

type HorizontalPodAutoscaler struct {
	*Resource
}

func NewHorizontalPodAutoscaler(config *config.KubeConfig) *HorizontalPodAutoscaler {
	p := &HorizontalPodAutoscaler{}
	gvr := HorizontalPodAutoscalerGVR
	if config.Client.VersionGreaterThan(types.ServerVersion22) {
		gvr = HorizontalPodAutoscalerV1GVR
	}
	p.Resource = NewResource(config, types.HpaType, gvr, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.DeleteAction: p.Delete,
		types.UpdateAction: p.Update,
	}
	return p
}

type BuildHorizontalPodAutoscaler struct {
	Name      string `json:"name"`
	NameSpace string `json:"namespace"`
	MinPods   *int32 `json:"min_pods"`
	MaxPods   int32  `json:"max_pods"`
	Replicas  int32  `json:"replicas"`
	//Status        autoscalingv2beta1.HorizontalPodAutoscalerStatus `json:"status"`
	CreateTime    metav1.Time `json:"create_time"`
	TargetCpuPer  *int32      `json:"target_cpu_per"`
	CurrentCpuPer *int32      `json:"current_cpu_per"`
}

func (h *HorizontalPodAutoscaler) ToBuildHorizontalPodAutoscaler(hpa *autoscalingv2beta1.HorizontalPodAutoscaler) *BuildHorizontalPodAutoscaler {
	if hpa == nil {
		return nil
	}

	hpaData := &BuildHorizontalPodAutoscaler{
		Name:      hpa.Name,
		NameSpace: hpa.Namespace,
		MinPods:   hpa.Spec.MinReplicas,
		MaxPods:   hpa.Spec.MaxReplicas,
		Replicas:  hpa.Status.CurrentReplicas,
		//Status:     hpa.Status,
		CreateTime: hpa.CreationTimestamp,
		//TargetCpuPer: 	hpa.Spec.TargetCPUUtilizationPercentage,
		//CurrentCpuPer:	hpa.Status.CurrentCPUUtilizationPercentage,
	}

	return hpaData
}

func (h *HorizontalPodAutoscaler) ToBuildV1HorizontalPodAutoscaler(hpa *autoscalingv1.HorizontalPodAutoscaler) *BuildHorizontalPodAutoscaler {
	if hpa == nil {
		return nil
	}

	hpaData := &BuildHorizontalPodAutoscaler{
		Name:      hpa.Name,
		NameSpace: hpa.Namespace,
		MinPods:   hpa.Spec.MinReplicas,
		MaxPods:   hpa.Spec.MaxReplicas,
		Replicas:  hpa.Status.CurrentReplicas,
		//Status:     hpa.Status,
		CreateTime: hpa.CreationTimestamp,
		//TargetCpuPer: 	hpa.Spec.TargetCPUUtilizationPercentage,
		//CurrentCpuPer:	hpa.Status.CurrentCPUUtilizationPercentage,
	}

	return hpaData
}

func (h *HorizontalPodAutoscaler) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	if h.config.Client.VersionGreaterThan(types.ServerVersion22) {
		hpa := &autoscalingv1.HorizontalPodAutoscaler{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, hpa); err != nil {
			return nil, err
		}
		return h.ToBuildV1HorizontalPodAutoscaler(hpa), nil
	}
	hpa := &autoscalingv2beta1.HorizontalPodAutoscaler{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, hpa); err != nil {
		return nil, err
	}
	return h.ToBuildHorizontalPodAutoscaler(hpa), nil
}
