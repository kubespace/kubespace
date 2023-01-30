package resource

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/kubernetes/config"
	"github.com/kubespace/kubespace/pkg/kubernetes/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"math"
	"time"
)

var NodeGVR = &schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "nodes",
}

type Node struct {
	*Resource
}

func NewNode(config *config.KubeConfig) *Node {
	p := &Node{}
	p.Resource = NewResource(config, types.PodType, NodeGVR, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.UpdateAction: p.Update,
	}
	return p
}

type BuildNode struct {
	UID              string            `json:"uid"`
	Name             string            `json:"name"`
	Taints           int               `json:"taints"`
	Roles            string            `json:"roles"`
	Version          string            `json:"version"`
	Age              string            `json:"age"`
	Status           string            `json:"status"`
	OS               string            `json:"os"`
	OSImage          string            `json:"os_image"`
	KernelVersion    string            `json:"kernel_version"`
	ContainerRuntime string            `json:"container_runtime"`
	Labels           map[string]string `json:"labels"`
	TotalCPU         string            `json:"total_cpu"`
	AllocatableCpu   string            `json:"allocatable_cpu"`
	TotalMem         string            `json:"total_mem"`
	AllocatableMem   string            `json:"allocatable_mem"`
	InternalIP       string            `json:"internal_ip"`
	Created          metav1.Time       `json:"created"`
}

func (n *Node) ToBuildNode(node *corev1.Node) *BuildNode {
	if node == nil {
		return nil
	}
	nodeData := &BuildNode{
		UID:              string(node.UID),
		Name:             node.Name,
		Taints:           len(node.Spec.Taints),
		Version:          node.Status.NodeInfo.KubeletVersion,
		OS:               node.Status.NodeInfo.OperatingSystem,
		OSImage:          node.Status.NodeInfo.OSImage,
		KernelVersion:    node.Status.NodeInfo.KernelVersion,
		ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
		Labels:           node.Labels,
		AllocatableCpu:   node.Status.Allocatable.Cpu().String(),
		TotalCPU:         node.Status.Capacity.Cpu().String(),
		AllocatableMem:   node.Status.Allocatable.Memory().String(),
		TotalMem:         node.Status.Capacity.Memory().String(),
		Created:          node.CreationTimestamp,
	}
	dur := time.Now().Sub(node.CreationTimestamp.Time)
	nodeData.Age = fmt.Sprintf("%vd", math.Floor(dur.Hours()/24))

	for _, c := range node.Status.Conditions {
		if c.Type == "Ready" && c.Status == corev1.ConditionTrue {
			nodeData.Status = "Ready"
		} else {
			nodeData.Status = "NotReady"
		}
	}
	for _, i := range node.Status.Addresses {
		if i.Type == corev1.NodeInternalIP {
			nodeData.InternalIP = i.Address
		}
	}

	return nodeData
}

func (n *Node) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	object := &corev1.Node{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, object); err != nil {
		return nil, err
	}
	return n.ToBuildNode(object), nil
}
