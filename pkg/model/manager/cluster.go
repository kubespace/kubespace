package manager

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/model/manager/project"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"gorm.io/gorm"
	"k8s.io/klog"
	"strconv"
)

type ClusterManager struct {
	//*CommonManager
	*gorm.DB
	appManager *project.AppManager
}

func NewClusterManager(db *gorm.DB, appMgr *project.AppManager) *ClusterManager {
	return &ClusterManager{
		//CommonManager: NewCommonManager(redisClient, nil, "osp:cluster", true),
		DB:         db,
		appManager: appMgr,
	}
}

//func (clu *ClusterManager) parseToStore(cluster *types.Cluster) (*types.ClusterStore, error) {
//	members, err := json.Marshal(cluster.Members)
//	if err != nil {
//		klog.Error("parse member error", err)
//		return nil, err
//	}
//	clusterStore := &types.ClusterStore{
//		Name:      cluster.Name,
//		Status:    cluster.Status,
//		Token:     cluster.Token,
//		Members:   string(members),
//		Common:    cluster.Common,
//		CreatedBy: cluster.CreatedBy,
//	}
//	return clusterStore, nil
//}
//
//func (clu *ClusterManager) parseToCluster(clusterStore *types.ClusterStore) (*types.Cluster, error) {
//	var members []string
//	if clusterStore.Members != "" {
//		err := json.Unmarshal([]byte(clusterStore.Members), &members)
//		if err != nil {
//			klog.Error("parse member error: ", err)
//			return nil, err
//		}
//	}
//	cluster := &types.Cluster{
//		Name:      clusterStore.Name,
//		Status:    clusterStore.Status,
//		Token:     clusterStore.Token,
//		Members:   members,
//		Common:    clusterStore.Common,
//		CreatedBy: clusterStore.CreatedBy,
//	}
//	return cluster, nil
//}

func (clu *ClusterManager) Create(cluster *types.Cluster) error {
	//clusterStore, err := clu.parseToStore(cluster)
	//if err != nil {
	//	return err
	//}
	if err := clu.DB.Create(cluster).Error; err != nil {
		return err
	}
	cluster.Name = fmt.Sprintf("%d", cluster.ID)
	return nil
}

func (clu *ClusterManager) Update(cluster *types.Cluster) error {
	//clusterStore, err := clu.parseToStore(cluster)
	//if err != nil {
	//	return err
	//}
	if err := clu.DB.Save(cluster).Error; err != nil {
		return err
	}
	cluster.Name = fmt.Sprintf("%d", cluster.ID)
	return nil
}

func (clu *ClusterManager) Get(id uint) (*types.Cluster, error) {
	cluster := &types.Cluster{}
	if err := clu.DB.First(cluster, "id = ?", id).Error; err != nil {
		return nil, err
	}
	cluster.Name = fmt.Sprintf("%d", cluster.ID)
	return cluster, nil
}

func (clu *ClusterManager) GetByName(name string) (*types.Cluster, error) {
	cluster := &types.Cluster{}
	id, err := strconv.ParseUint(name, 10, 64)
	if err != nil {
		return nil, err
	}
	if err := clu.DB.First(cluster, "id = ?", id).Error; err != nil {
		return nil, err
	}
	cluster.Name = fmt.Sprintf("%d", cluster.ID)
	return cluster, nil
}

func (clu *ClusterManager) List(filters map[string]interface{}) ([]types.Cluster, error) {
	//dList, err := clu.CommonManager.List(filters)
	//if err != nil {
	//	return nil, err
	//}
	//jsonBody, err := json.Marshal(dList)
	//if err != nil {
	//	return nil, err
	//}
	//var clus []*types.ClusterStore
	//
	//if err := json.Unmarshal(jsonBody, &clus); err != nil {
	//	return nil, err
	//}

	var clusters []types.Cluster
	if err := clu.DB.Find(&clusters, filters).Error; err != nil {
		return nil, err
	}
	for i, c := range clusters {
		clusters[i].Name = fmt.Sprintf("%d", c.ID)
		klog.Info(c.Name)
	}

	//for _, c := range clusters {
	//	//cluster, err := clu.parseToCluster(c)
	//	if err != nil {
	//		return nil, err
	//	}
	//	clusters = append(clusters, cluster)
	//}
	return clusters, nil
}

func (clu *ClusterManager) GetByToken(token string) (*types.Cluster, error) {

	clusterList, err := clu.List(map[string]interface{}{
		"token": token,
	})
	if err != nil {
		return nil, err
	}
	for _, cluster := range clusterList {
		if cluster.Token == token {
			return &cluster, nil
		}
	}
	return nil, nil
}

func (clu *ClusterManager) Delete(id uint) error {
	var cnt int64
	if err := clu.DB.Model(&types.Project{}).Where("cluster_id=?", id).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return fmt.Errorf("当前集群存在工作空间绑定")
	}
	var apps []types.ProjectApp
	if err := clu.DB.Find(&apps, "scope=? and scope_id=?", types.AppVersionScopeComponent, id).Error; err != nil {
		return err
	}
	for _, app := range apps {
		if err := clu.appManager.DeleteProjectApp(app.ID); err != nil {
			return err
		}
	}
	if err := clu.DB.Delete(types.Cluster{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
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
