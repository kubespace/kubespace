package resource

import (
	"context"
	"github.com/kubespace/kubespace/pkg/kubernetes/config"
	"github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type Cluster struct {
	*Resource
}

func NewCluster(config *config.KubeConfig) *Cluster {
	p := &Cluster{}
	p.Resource = NewResource(config, "", nil, nil)
	p.actions = map[string]ActionHandle{
		types.GetAction:   p.Get,
		types.ApplyAction: p.Apply,
	}
	return p
}

type BuildCluster struct {
	ClusterVersion  string `json:"cluster_version"`
	ClusterCpu      string `json:"cluster_cpu"`
	ClusterMemory   string `json:"cluster_memory"`
	NodeNum         int    `json:"node_num"`
	NamespaceNum    int    `json:"namespace_num"`
	PodNum          int    `json:"pod_num"`
	PodRunningNum   int    `json:"pod_running_num"`
	PodSucceededNum int    `json:"pod_succeeded_num"`
	PodPendingNum   int    `json:"pod_pending_num"`
	PodFailedNum    int    `json:"pod_failed_num"`
	DeploymentNum   int    `json:"deployment_num"`
	StatefulSetNum  int    `json:"statefulset_num"`
	DaemonSetNum    int    `json:"daemonset_num"`
	ServiceNum      int    `json:"service_num"`
	IngressNum      int    `json:"ingress_num"`
	StorageClassNum int    `json:"storageclass_num"`
	PVNum           int    `json:"pv_num"`
	PVAvailableNum  int    `json:"pv_available_num"`
	PVReleasedNum   int    `json:"pv_released_num"`
	PVBoundNum      int    `json:"pv_bound_num"`
	PVFailedNum     int    `json:"pv_failed_num"`
	PVCNum          int    `json:"pvc_num"`
	ConfigMapNum    int    `json:"config_map_num"`
	SecretNum       int    `json:"secret_num"`
}

type ClusterQueryParams struct {
	Workspace   uint   `json:"workspace"`
	Namespace   string `json:"namespace"`
	OnlyVersion bool   `json:"only_version"`
}

func (c *Cluster) Get(params interface{}) *utils.Response {
	query := &ClusterQueryParams{}
	if err := utils.ConvertTypeByJson(params, query); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	if query.OnlyVersion {
		version, err := c.client.Discovery().ServerVersion()
		if err != nil {
			return &utils.Response{Code: code.RequestError, Msg: err.Error()}
		}
		return &utils.Response{Code: code.Success, Data: version.GitVersion}
	}
	if query.Workspace != 0 {
		return c.WorkspaceOverview(query)
	}
	var bc = BuildCluster{}
	version, err := c.client.Discovery().ServerVersion()
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	bc.ClusterVersion = version.GitVersion
	ctx := context.Background()
	listOptions := metav1.ListOptions{}

	nodes, err := c.client.CoreV1().Nodes().List(ctx, listOptions)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	bc.NodeNum = len(nodes.Items)
	var cpu resource.Quantity
	var memory resource.Quantity
	for _, n := range nodes.Items {
		cpu.Add(*n.Status.Capacity.Cpu())
		memory.Add(*n.Status.Capacity.Memory())
	}
	bc.ClusterCpu = cpu.String()
	bc.ClusterMemory = memory.String()
	namespaces, err := c.client.CoreV1().Namespaces().List(ctx, listOptions)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	bc.NamespaceNum = len(namespaces.Items)
	pods, err := c.client.CoreV1().Pods("").List(ctx, listOptions)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	bc.PodNum = len(pods.Items)
	for _, p := range pods.Items {
		if p.Status.Phase == corev1.PodRunning {
			bc.PodRunningNum += 1
		} else if p.Status.Phase == corev1.PodPending {
			bc.PodPendingNum += 1
		} else if p.Status.Phase == corev1.PodFailed {
			bc.PodFailedNum += 1
		} else if p.Status.Phase == corev1.PodSucceeded {
			bc.PodSucceededNum += 1
		}
	}
	deployments, err := c.client.AppsV1().Deployments("").List(ctx, listOptions)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	bc.DeploymentNum = len(deployments.Items)
	statefulsets, err := c.client.AppsV1().StatefulSets("").List(ctx, listOptions)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	bc.StatefulSetNum = len(statefulsets.Items)
	daemonsets, err := c.client.AppsV1().DaemonSets("").List(ctx, listOptions)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	bc.DaemonSetNum = len(daemonsets.Items)
	services, err := c.client.CoreV1().Services("").List(ctx, listOptions)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	bc.ServiceNum = len(services.Items)
	if c.client.VersionGreaterThan(types.ServerVersion19) {
		ingresses, err := c.client.NetworkingV1().Ingresses("").List(ctx, listOptions)
		if err != nil {
			return &utils.Response{Code: code.RequestError, Msg: err.Error()}
		}
		bc.IngressNum = len(ingresses.Items)
	} else {
		ingresses, err := c.client.ExtensionsV1beta1().Ingresses("").List(ctx, listOptions)
		if err != nil {
			return &utils.Response{Code: code.RequestError, Msg: err.Error()}
		}
		bc.IngressNum = len(ingresses.Items)
	}
	sc, err := c.client.StorageV1().StorageClasses().List(ctx, listOptions)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	bc.StorageClassNum = len(sc.Items)
	pv, err := c.client.CoreV1().PersistentVolumes().List(ctx, listOptions)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	bc.PVNum = len(pv.Items)
	for _, p := range pv.Items {
		if p.Status.Phase == corev1.VolumeAvailable {
			bc.PVAvailableNum += 1
		} else if p.Status.Phase == corev1.VolumeBound {
			bc.PVBoundNum += 1
		} else if p.Status.Phase == corev1.VolumeReleased {
			bc.PVReleasedNum += 1
		} else if p.Status.Phase == corev1.VolumeFailed {
			bc.PVFailedNum += 1
		}
	}
	pvc, err := c.client.CoreV1().PersistentVolumeClaims("").List(ctx, listOptions)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	bc.PVCNum = len(pvc.Items)
	return &utils.Response{Code: code.Success, Msg: "Success", Data: bc}
}

func (c *Cluster) WorkspaceOverview(queryParams *ClusterQueryParams) *utils.Response {
	if queryParams.Namespace == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "参数namespace为空"}
	}
	namespace := queryParams.Namespace
	ctx := context.Background()
	labelSelector := labels.Set(map[string]string{"kubespace.cn/belong-to": "project"}).AsSelector()
	listOptions := metav1.ListOptions{LabelSelector: labelSelector.String()}

	var bc = BuildCluster{}
	bc.ClusterVersion = c.client.ServerVersion().GitVersion

	services, err := c.client.CoreV1().Services(namespace).List(ctx, listOptions)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	bc.ServiceNum = len(services.Items)
	if c.client.VersionGreaterThan(types.ServerVersion19) {
		ingresses, err := c.client.NetworkingV1().Ingresses(namespace).List(ctx, listOptions)
		if err != nil {
			return &utils.Response{Code: code.RequestError, Msg: err.Error()}
		}
		bc.IngressNum = len(ingresses.Items)
	} else {
		ingresses, err := c.client.ExtensionsV1beta1().Ingresses(namespace).List(ctx, listOptions)
		if err != nil {
			return &utils.Response{Code: code.RequestError, Msg: err.Error()}
		}
		bc.IngressNum = len(ingresses.Items)
	}

	pvc, err := c.client.CoreV1().PersistentVolumeClaims(namespace).List(ctx, listOptions)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	bc.PVCNum = len(pvc.Items)

	cm, err := c.client.CoreV1().ConfigMaps(namespace).List(ctx, listOptions)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	bc.ConfigMapNum = len(cm.Items)
	secret, err := c.client.CoreV1().Secrets(namespace).List(ctx, listOptions)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	bc.SecretNum = len(secret.Items)
	return &utils.Response{Code: code.Success, Msg: "Success", Data: bc}
}
