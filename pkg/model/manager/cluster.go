package manager

import (
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog"
)

type ClusterManager struct {
	*CommonManager
}

func NewClusterManager(redisClient *redis.Client) *ClusterManager {
	return &ClusterManager{
		CommonManager: NewCommonManager(redisClient, nil, "osp:cluster", true),
	}
}

func (clu *ClusterManager) parseToStore(cluster *types.Cluster) (*types.ClusterStore, error) {
	members, err := json.Marshal(cluster.Members)
	if err != nil {
		klog.Error("parse member error", err)
		return nil, err
	}
	clusterStore := &types.ClusterStore{
		Name:      cluster.Name,
		Status:    cluster.Status,
		Token:     cluster.Token,
		Members:   string(members),
		Common:    cluster.Common,
		CreatedBy: cluster.CreatedBy,
	}
	return clusterStore, nil
}

func (clu *ClusterManager) parseToCluster(clusterStore *types.ClusterStore) (*types.Cluster, error) {
	var members []string
	if clusterStore.Members != "" {
		err := json.Unmarshal([]byte(clusterStore.Members), &members)
		if err != nil {
			klog.Error("parse member error: ", err)
			return nil, err
		}
	}
	cluster := &types.Cluster{
		Name:      clusterStore.Name,
		Status:    clusterStore.Status,
		Token:     clusterStore.Token,
		Members:   members,
		Common:    clusterStore.Common,
		CreatedBy: clusterStore.CreatedBy,
	}
	return cluster, nil
}

func (clu *ClusterManager) Create(cluster *types.Cluster) error {
	clusterStore, err := clu.parseToStore(cluster)
	if err != nil {
		return err
	}
	if err := clu.CommonManager.Save(cluster.Name, clusterStore, -1, true); err != nil {
		return err
	}

	return nil
}

func (clu *ClusterManager) Update(cluster *types.Cluster) error {
	clusterStore, err := clu.parseToStore(cluster)
	if err != nil {
		return err
	}
	if err := clu.CommonManager.Update(cluster.Name, clusterStore, -1, false); err != nil {
		return err
	}

	return nil
}

func (clu *ClusterManager) Get(name string) (*types.Cluster, error) {
	clusterStore := &types.ClusterStore{}
	if err := clu.CommonManager.Get(name, clusterStore); err != nil {
		return nil, err
	}
	return clu.parseToCluster(clusterStore)
}

func (clu *ClusterManager) List(filters map[string]interface{}) ([]*types.Cluster, error) {
	dList, err := clu.CommonManager.List(filters)
	if err != nil {
		return nil, err
	}
	jsonBody, err := json.Marshal(dList)
	if err != nil {
		return nil, err
	}
	var clus []*types.ClusterStore

	if err := json.Unmarshal(jsonBody, &clus); err != nil {
		return nil, err
	}

	var clusters []*types.Cluster
	for _, c := range clus {
		cluster, err := clu.parseToCluster(c)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, cluster)
	}
	return clusters, nil
}

func (clu *ClusterManager) GetByToken(token string) (*types.Cluster, error) {

	clusterList, err := clu.List(map[string]interface{}{
		"token": token,
	})
	if err != nil {
		return nil, err
	}
	for _, clu := range clusterList {
		if clu.Token == token {
			return clu, nil
		}
	}
	return nil, nil
}

func (clu *ClusterManager) HasMember(cluster *types.Cluster, user *types.User) bool {
	// 不是超级用户且当前用户不在集群邀请之内
	if user.IsSuper {
		return true
	}
	if cluster.CreatedBy == user.Name {
		return true
	}
	if utils.Contains(cluster.Members, user.Name) {
		return true
	}
	return false
}
