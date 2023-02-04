package cluster

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/config"
	"github.com/kubespace/kubespace/pkg/model/manager/project"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
	"strconv"
	"time"
)

type ClusterManager struct {
	//*CommonManager
	*gorm.DB
	listWatcherConfig *config.ListWatcherConfig
	appManager        *project.AppManager
}

func NewClusterManager(db *gorm.DB, listWatcherConfig *config.ListWatcherConfig, appMgr *project.AppManager) *ClusterManager {
	c := &ClusterManager{
		DB:                db,
		appManager:        appMgr,
		listWatcherConfig: listWatcherConfig,
	}
	var cnt int64
	err := db.Model(&types.Cluster{}).Where("").Count(&cnt).Error
	if cnt == 0 {
		localCluster := &types.Cluster{
			Name1:      "local",
			Token:      utils.ShortUUID(),
			CreatedBy:  "admin",
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		}
		if err = c.Create(localCluster); err != nil {
			klog.Errorf("init create local cluster error: %s", err.Error())
		}
	}
	return c
}

func (clu *ClusterManager) Create(cluster *types.Cluster) error {
	if err := clu.DB.Create(cluster).Error; err != nil {
		return err
	}
	cluster.Name = fmt.Sprintf("%d", cluster.ID)
	return nil
}

func (clu *ClusterManager) Update(cluster *types.Cluster) error {
	if err := clu.DB.Save(cluster).Error; err != nil {
		return err
	}
	cluster.Name = fmt.Sprintf("%d", cluster.ID)
	return nil
}

func (clu *ClusterManager) UpdateByObject(id uint, cluster *types.Cluster) error {
	return clu.DB.Where("id=?", id).Updates(cluster).Error
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
	if err = clu.DB.First(cluster, "id = ?", id).Error; err != nil {
		return nil, err
	}
	cluster.Name = fmt.Sprintf("%d", cluster.ID)
	return cluster, nil
}

func (clu *ClusterManager) List(filters map[string]interface{}) ([]types.Cluster, error) {

	var clusters []types.Cluster
	if err := clu.DB.Find(&clusters, filters).Error; err != nil {
		return nil, err
	}
	for i, c := range clusters {
		clusters[i].Name = fmt.Sprintf("%d", c.ID)
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
	if err := clu.DB.Delete(&types.UserRole{}, "scope = ? and scope_id = ?", types.RoleScopeCluster, id).Error; err != nil {
		return err
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
