package kube_resource

import (
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
)

const (
	ListAction       = "list"
	ExecAction       = "exec"
	CloseExecConn    = "closeExecConn"
	GetAction        = "get"
	DeleteAction     = "delete"
	UpdateYamlAction = "update_yaml"
	UpdateObjAction  = "update_obj"
	StdinAction      = "stdin"
	OpenLogAction    = "openLog"
	CloseLogAction   = "closeLog"
	APPLY            = "apply"
	CREATE           = "create"
	STATUS           = "status"
)

type KubeResource struct {
	ResType     string
	KubeMessage *MiddleMessage
}

func (k *KubeResource) Get(cluster string, params interface{}) *utils.Response {
	return k.request(cluster, GetAction, params)
}

func (k *KubeResource) List(cluster string, params interface{}) *utils.Response {
	return k.request(cluster, ListAction, params)
}

func (k *KubeResource) Delete(cluster string, params interface{}) *utils.Response {
	return k.request(cluster, DeleteAction, params)
}

func (k *KubeResource) UpdateYaml(cluster string, params interface{}) *utils.Response {
	return k.request(cluster, UpdateYamlAction, params)
}

func (k *KubeResource) UpdateObj(cluster string, params interface{}) *utils.Response {
	return k.request(cluster, UpdateObjAction, params)
}

func (k *KubeResource) Create(cluster string, params interface{}) *utils.Response {
	return k.request(cluster, CREATE, params)
}

func (k *KubeResource) Apply(cluster string, params interface{}) *utils.Response {
	return k.request(cluster, APPLY, params)
}

func (k *KubeResource) Exec(cluster string, params interface{}) *utils.Response {
	return k.request(cluster, ExecAction, params)
}

func (k *KubeResource) CloseExecConn(cluster string, params interface{}) *utils.Response {
	return k.request(cluster, CloseExecConn, params)
}

func (k *KubeResource) Stdin(cluster string, params interface{}) *utils.Response {
	return k.request(cluster, StdinAction, params)
}

func (k *KubeResource) OpenLog(cluster string, params interface{}) *utils.Response {
	return k.request(cluster, OpenLogAction, params)
}

func (k *KubeResource) CloseLog(cluster string, params interface{}) *utils.Response {
	return k.request(cluster, CloseLogAction, params)
}

func (k *KubeResource) Status(cluster string, params interface{}) *utils.Response {
	return k.request(cluster, STATUS, params)
}

func (k *KubeResource) request(cluster, action string, params interface{}) *utils.Response {
	middleRequest := NewMiddleRequest(cluster, k.ResType, action, params, 120)
	res := k.KubeMessage.SendRequest(middleRequest)
	return res
}

type WatchResource struct {
	*KubeResource
}

func (w *WatchResource) OpenWatch(cluster string, params ) *utils.Response {
	return w.Get(cluster, map[string]interface{}{"action": "open"})
}

func (w *WatchResource) CloseWatch(cluster string) *utils.Response {
	hasReceive := w.KubeMessage.HasWatchReceive(cluster)
	if !hasReceive {
		return w.Get(cluster, map[string]interface{}{"action": "close"})
	}
	return &utils.Response{Code: code.Success}
}

const (
	WatchResType       = "watch"
	PodType            = "pod"
	ClusterType        = "cluster"
	EventType          = "event"
	NodeType           = "node"
	DeploymentType     = "deployment"
	StatefulsetType    = "statefulset"
	DaemonsetType      = "daemonset"
	CronjobType        = "cronjob"
	JobType            = "job"
	NamespaceType      = "namespace"
	ServiceType        = "service"
	IngressType        = "ingress"
	NetworkPolicyType  = "networkpolicy"
	EndpointType       = "endpoints"
	ServiceAccountType = "serviceaccount"
	RolebindingType    = "rolebinding"
	RoleType           = "role"
	ConfigMapType      = "configMap"
	SecretType         = "secret"
	HpaType            = "horizontalPodAutoscaler"
	PvcType            = "persistentVolumeClaim"
	PVType             = "persistentVolume"
	StorageClassType   = "storageClass"
	Helm               = "helm"
	Crd                = "crd"
)

type KubeResources struct {
	Watch          *WatchResource
	Pod            *KubeResource
	Cluster        *KubeResource
	Event          *KubeResource
	Node           *KubeResource
	Deployment     *KubeResource
	Statefulset    *KubeResource
	Daemonset      *KubeResource
	Cronjob        *KubeResource
	Job            *KubeResource
	Namespace      *KubeResource
	Service        *KubeResource
	Ingress        *KubeResource
	NetworkPolicy  *KubeResource
	Endpoint       *KubeResource
	ServiceAccount *KubeResource
	Rolebinding    *KubeResource
	Role           *KubeResource
	ConfigMap      *KubeResource
	Secret         *KubeResource
	Hpa            *KubeResource
	Pvc            *KubeResource
	PV             *KubeResource
	StorageClass   *KubeResource
	Helm           *KubeResource
	Crd            *KubeResource
}

func NewKubeResources(message *MiddleMessage) *KubeResources {
	return &KubeResources{
		Watch:          &WatchResource{&KubeResource{ResType: WatchResType, KubeMessage: message}},
		Pod:            &KubeResource{ResType: PodType, KubeMessage: message},
		Cluster:        &KubeResource{ResType: ClusterType, KubeMessage: message},
		Event:          &KubeResource{ResType: EventType, KubeMessage: message},
		Node:           &KubeResource{ResType: NodeType, KubeMessage: message},
		Deployment:     &KubeResource{ResType: DeploymentType, KubeMessage: message},
		Statefulset:    &KubeResource{ResType: StatefulsetType, KubeMessage: message},
		Daemonset:      &KubeResource{ResType: DaemonsetType, KubeMessage: message},
		Cronjob:        &KubeResource{ResType: CronjobType, KubeMessage: message},
		Job:            &KubeResource{ResType: JobType, KubeMessage: message},
		Namespace:      &KubeResource{ResType: NamespaceType, KubeMessage: message},
		Service:        &KubeResource{ResType: ServiceType, KubeMessage: message},
		Ingress:        &KubeResource{ResType: IngressType, KubeMessage: message},
		NetworkPolicy:  &KubeResource{ResType: NetworkPolicyType, KubeMessage: message},
		Endpoint:       &KubeResource{ResType: EndpointType, KubeMessage: message},
		ServiceAccount: &KubeResource{ResType: ServiceAccountType, KubeMessage: message},
		Rolebinding:    &KubeResource{ResType: RolebindingType, KubeMessage: message},
		Role:           &KubeResource{ResType: RoleType, KubeMessage: message},
		ConfigMap:      &KubeResource{ResType: ConfigMapType, KubeMessage: message},
		Secret:         &KubeResource{ResType: SecretType, KubeMessage: message},
		Hpa:            &KubeResource{ResType: HpaType, KubeMessage: message},
		Pvc:            &KubeResource{ResType: PvcType, KubeMessage: message},
		PV:             &KubeResource{ResType: PVType, KubeMessage: message},
		StorageClass:   &KubeResource{ResType: StorageClassType, KubeMessage: message},
		Helm:           &KubeResource{ResType: Helm, KubeMessage: message},
		Crd:            &KubeResource{ResType: Crd, KubeMessage: message},
	}
}
