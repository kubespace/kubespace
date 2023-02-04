package kubernetes

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/kubernetes/config"
	"github.com/kubespace/kubespace/pkg/kubernetes/resource"
	"github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"sync"
)

type KubeFactory interface {
	GetResource(resType string) (ResourceHandler, error)
	GetPod() (PodHandler, error)
}

type ResourceHandler interface {
	Handle(action string, params interface{}) *utils.Response
	Watch(params interface{}, writer resource.OutWriter) *utils.Response
}

var _ PodHandler = &resource.Pod{}

type PodHandler interface {
	ResourceHandler
	Exec(params interface{}, writer resource.OutWriter) *utils.Response
	Log(params interface{}, writer resource.OutWriter) *utils.Response
}

type kubeFactory struct {
	config      *config.KubeConfig
	resourceMap map[string]ResourceHandler
	mu          sync.Mutex
}

func NewKubeFactory(config *config.KubeConfig) KubeFactory {
	return &kubeFactory{
		config:      config,
		resourceMap: make(map[string]ResourceHandler),
		mu:          sync.Mutex{},
	}
}

func (k *kubeFactory) GetResource(resType string) (ResourceHandler, error) {
	if ins, ok := k.resourceMap[resType]; ok {
		return ins, nil
	}
	k.mu.Lock()
	ins, err := k.getResource(resType)
	if err != nil {
		return nil, err
	}
	k.resourceMap[resType] = ins
	k.mu.Unlock()
	return ins, nil
}

func (k *kubeFactory) getResource(resType string) (ResourceHandler, error) {
	switch resType {
	case types.ClusterType:
		return resource.NewCluster(k.config), nil
	case types.PodType:
		return resource.NewPod(k.config), nil
	case types.NamespaceType:
		return resource.NewNamespace(k.config), nil
	case types.EventType:
		return resource.NewEvent(k.config), nil
	case types.DeploymentType:
		return resource.NewDeployment(k.config), nil
	case types.StatefulsetType:
		return resource.NewStatefulSet(k.config), nil
	case types.DaemonsetType:
		return resource.NewDaemonSet(k.config), nil
	case types.JobType:
		return resource.NewJob(k.config), nil
	case types.CronjobType:
		return resource.NewCronJob(k.config), nil
	case types.ConfigMapType:
		return resource.NewConfigmap(k.config), nil
	case types.SecretType:
		return resource.NewSecret(k.config), nil
	case types.ServiceType:
		return resource.NewService(k.config), nil
	case types.IngressType:
		return resource.NewIngress(k.config), nil
	case types.NetworkPolicyType:
		return resource.NewNetworkPolicy(k.config), nil
	case types.RoleType:
		return resource.NewRole(k.config), nil
	case types.ClusterRoleType:
		return resource.NewClusterRole(k.config), nil
	case types.ServiceAccountType:
		return resource.NewServiceAccount(k.config), nil
	case types.RoleBindingType:
		return resource.NewRoleBinding(k.config), nil
	case types.ClusterRoleBindingType:
		return resource.NewClusterRoleBinding(k.config), nil
	case types.NodeType:
		return resource.NewNode(k.config), nil
	case types.PersistentVolumeType:
		return resource.NewPersistentVolume(k.config), nil
	case types.PersistentVolumeClaimType:
		return resource.NewPersistentVolumeClaim(k.config), nil
	case types.StorageClassType:
		return resource.NewStorageClass(k.config), nil
	case types.HpaType:
		return resource.NewHorizontalPodAutoscaler(k.config), nil
	case types.EndpointType:
		return resource.NewEndpoints(k.config), nil
	case types.CustomResourceDefinitionType:
		return resource.NewCustomResourceDefinition(k.config), nil
	case types.CustomResourceType:
		return resource.NewCustomResource(k.config), nil
	case types.HelmType:
		return resource.NewHelm(k.config), nil
	}
	return nil, fmt.Errorf("not found kubernetes %s resource handler", resType)
}

func (k *kubeFactory) GetPod() (PodHandler, error) {
	ins, err := k.GetResource(types.PodType)
	if err != nil {
		return nil, err
	}
	podIns, ok := ins.(PodHandler)
	if !ok {
		return nil, fmt.Errorf("not found kubernetes pod handler")
	}
	return podIns, nil
}
