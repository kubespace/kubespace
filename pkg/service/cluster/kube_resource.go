package cluster

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/listwatcher/cluster"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
)

type KubeResource struct {
	models      *model.Models
	kubeClient  *kubeClient
	agentClient *agentClient
}

func NewKubeResource(models *model.Models) *KubeResource {
	return &KubeResource{
		models:      models,
		kubeClient:  &kubeClient{},
		agentClient: &agentClient{models: models},
	}
}

func (k *KubeResource) List(token, resType string, params interface{}) {

}

func (k *KubeResource) Request(token, resType, action string, params interface{}) *utils.Response {
	var kr kubeRequest
	clusterObj, err := k.models.ClusterManager.GetByToken(token)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: fmt.Sprintf("获取集群%s失败：%s", token, err.Error())}
	}
	if clusterObj == nil {
		return &utils.Response{Code: code.DBError, Msg: fmt.Sprintf("未找到集群%s", token)}
	}
	if clusterObj.KubeConfig != "" {
		kr = k.kubeClient
	} else {
		kr = k.agentClient
	}
	return kr.Request(clusterObj, resType, action, params)
}

type kubeRequest interface {
	Request(cluster *types.Cluster, resType, action string, params interface{}) *utils.Response
}

type kubeClient struct {
}

func (k *kubeClient) Request(cluster *types.Cluster, resType, action string, params interface{}) *utils.Response {
	return nil
}

type agentClient struct {
	models *model.Models
}

func (a *agentClient) Request(clusterObj *types.Cluster, resType, action string, params interface{}) *utils.Response {
	agentListWatcher := cluster.NewAgentListWatcher(clusterObj.Token, a.models.ListWatcherConfig)
	agentListWatcher.Notify(nil)
	return nil
}
