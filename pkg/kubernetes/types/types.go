package types

import (
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ServerVersion16 = "v1.16.0"
const ServerVersion17 = "v1.17.0"
const ServerVersion19 = "v1.19.0"
const ServerVersion21 = "v1.21.0"
const ServerVersion22 = "v1.22.0"

var ProjectLabelSelector = &metav1.LabelSelector{
	MatchLabels: map[string]string{"kubespace.cn/belong-to": "project"},
}

type Request struct {
	TraceId  string      `json:"trace_id"`
	Resource string      `json:"resource"`
	Action   string      `json:"action"`
	Params   interface{} `json:"params"`
}

func (r *Request) Unmarshal(data []byte) (interface{}, error) {
	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}
	return req, nil
}

type Response struct {
	TraceId string      `json:"trace_id"`
	Data    interface{} `json:"data"`
}

const (
	PodType                      = "pod"
	ClusterType                  = "cluster"
	EventType                    = "event"
	NodeType                     = "node"
	DeploymentType               = "deployment"
	StatefulsetType              = "statefulset"
	DaemonsetType                = "daemonset"
	CronjobType                  = "cronjob"
	JobType                      = "job"
	NamespaceType                = "namespace"
	ServiceType                  = "service"
	IngressType                  = "ingress"
	NetworkPolicyType            = "networkpolicy"
	EndpointType                 = "endpoints"
	ServiceAccountType           = "serviceaccount"
	RoleBindingType              = "rolebinding"
	ClusterRoleBindingType       = "clusterrolebinding"
	RoleType                     = "role"
	ClusterRoleType              = "clusterrole"
	ConfigMapType                = "configmap"
	SecretType                   = "secret"
	HpaType                      = "horizontalPodAutoscaler"
	PersistentVolumeClaimType    = "persistentVolumeClaim"
	PersistentVolumeType         = "persistentVolume"
	StorageClassType             = "storageclass"
	HelmType                     = "helm"
	CustomResourceDefinitionType = "crd"
	CustomResourceType           = "cr"
)

const (
	WatchAction  = "watch"
	ListAction   = "list"
	GetAction    = "get"
	CreateAction = "create"
	DeleteAction = "delete"
	UpdateAction = "update"
	ApplyAction  = "apply"
	PatchAction  = "patch"

	ExecAction   = "exec"
	StdinAction  = "stdin"
	LogAction    = "log"
	CloseSession = "close_session"
)
