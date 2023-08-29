package cluster

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	kubetypes "github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"strconv"
)

type kubeclient interface {
	request(cluster *types.Cluster, resType, action string, params interface{}) *utils.Response
	watch(cluster *types.Cluster, resType string, params interface{}) (Outer, error)
	pods(cluster *types.Cluster) (PodClient, error)
}

type PodClient interface {
	Exec(interface{}) (PodExec, error)
	Log(interface{}) (Outer, error)
}

type PodExec interface {
	Outer
	Stdin(interface{}) error
}

type KubeClient struct {
	models       *model.Models
	directClient kubeclient
	agentClient  kubeclient
}

func NewKubeClient(models *model.Models) *KubeClient {
	return &KubeClient{
		models:       models,
		directClient: &directClient{},
		agentClient:  NewAgentClient(models),
	}
}

func (k *KubeClient) List(clusterId, resType string, params interface{}) *utils.Response {
	return k.Request(clusterId, resType, kubetypes.ListAction, params)
}

func (k *KubeClient) Get(clusterId, resType string, params interface{}) *utils.Response {
	return k.Request(clusterId, resType, kubetypes.GetAction, params)
}

func (k *KubeClient) Delete(clusterId, resType string, params interface{}) *utils.Response {
	return k.Request(clusterId, resType, kubetypes.DeleteAction, params)
}

func (k *KubeClient) Update(clusterId, resType string, params interface{}) *utils.Response {
	return k.Request(clusterId, resType, kubetypes.UpdateAction, params)
}

func (k *KubeClient) Create(clusterId, resType string, params interface{}) *utils.Response {
	return k.Request(clusterId, resType, kubetypes.CreateAction, params)
}

func (k *KubeClient) Patch(clusterId, resType string, params interface{}) *utils.Response {
	return k.Request(clusterId, resType, kubetypes.PatchAction, params)
}

func (k *KubeClient) Apply(clusterId string, params interface{}) *utils.Response {
	return k.Request(clusterId, kubetypes.ClusterType, kubetypes.ApplyAction, params)
}

func (k *KubeClient) Watch(clusterId, resType string, params interface{}) (Outer, error) {
	cli, clusterObj, err := k.getClient(clusterId)
	if err != nil {
		return nil, err
	}
	return cli.watch(clusterObj, resType, params)
}

func (k *KubeClient) Pods(clusterId string) (PodClient, error) {
	cli, clusterObj, err := k.getClient(clusterId)
	if err != nil {
		return nil, err
	}
	return cli.pods(clusterObj)
}

func (k *KubeClient) getClient(clusterId string) (c kubeclient, clusterObj *types.Cluster, err error) {
	id, _ := strconv.Atoi(clusterId)
	clusterObj, err = k.models.ClusterManager.GetById(uint(id))
	if err != nil {
		return nil, nil, fmt.Errorf("获取集群%s失败：%s", clusterId, err.Error())
	}
	if clusterObj == nil {
		return nil, nil, fmt.Errorf("未找到集群%s", clusterId)
	}
	if clusterObj.KubeConfig != "" {
		c = k.directClient
	} else {
		c = k.agentClient
	}
	return
}

func (k *KubeClient) Request(clusterId, resType, action string, params interface{}) *utils.Response {
	cli, clusterObj, err := k.getClient(clusterId)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return cli.request(clusterObj, resType, action, params)
}

type Outer interface {
	OutCh() <-chan interface{}
	StopCh() <-chan struct{}
	Close()
}

type outer struct {
	outCh   chan interface{}
	stopCh  chan struct{}
	stopped bool
}

func newOuter() *outer {
	return &outer{
		outCh:  make(chan interface{}),
		stopCh: make(chan struct{}),
	}
}

func (o *outer) OutCh() <-chan interface{} {
	return o.outCh
}

func (o *outer) StopCh() <-chan struct{} {
	return o.stopCh
}

func (o *outer) Close() {
	if o.stopped {
		return
	}
	o.stopped = true
	close(o.stopCh)
	close(o.outCh)
}
