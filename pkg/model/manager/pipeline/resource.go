package pipeline

import (
	"github.com/kubespace/kubespace/pkg/model/manager"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
	"k8s.io/klog"
	"time"
)

var InitGlobalResources = []*types.PipelineResource{
	{
		Name:        "kubespace/golang:1.16",
		Type:        "image",
		Value:       "kubespace/golang:1.16",
		Global:      true,
		Description: "内置golang编译镜像",
		CreateUser:  "admin",
		UpdateUser:  "admin",
	},
	{
		Name:        "kubespace/node:17.9.0",
		Type:        "image",
		Value:       "kubespace/node:17.9.0",
		Global:      true,
		Description: "内置node编译镜像",
		CreateUser:  "admin",
		UpdateUser:  "admin",
	},
	{
		Name:        "kubespace/python:3.8",
		Type:        "image",
		Value:       "kubespace/python:3.8",
		Global:      true,
		Description: "内置python编译镜像",
		CreateUser:  "admin",
		UpdateUser:  "admin",
	},
}

type ResourceManager struct {
	*manager.CommonManager
}

func NewResourceManager(db *gorm.DB) *ResourceManager {
	res := &ResourceManager{
		CommonManager: manager.NewCommonManager(nil, db, "", false),
	}
	res.Init()
	return res
}

func (r *ResourceManager) Create(resource *types.PipelineResource) (*types.PipelineResource, error) {
	result := r.DB.Create(resource)
	if result.Error != nil {
		return nil, result.Error
	}
	return resource, nil
}

func (r *ResourceManager) Update(resource *types.PipelineResource) (*types.PipelineResource, error) {
	result := r.DB.Save(resource)
	if result.Error != nil {
		return nil, result.Error
	}
	return resource, nil
}

func (r *ResourceManager) Get(resourceId uint) (*types.PipelineResource, error) {
	var ws types.PipelineResource
	if err := r.DB.First(&ws, resourceId).Error; err != nil {
		return nil, err
	}
	var secret types.SettingsSecret
	if err := r.DB.Where("id = ?", ws.SecretId).First(&secret).Error; err == nil {
		ws.Secret = &secret
	}
	return &ws, nil
}

func (r *ResourceManager) List(workspaceId uint) ([]types.PipelineResource, error) {
	var ws []types.PipelineResource
	result := r.DB.Where("workspace_id = ? or global = 1", workspaceId).Find(&ws)
	if result.Error != nil {
		return nil, result.Error
	}
	var secret types.SettingsSecret
	for _, res := range ws {
		if err := r.DB.Where("id = ?", res.SecretId).First(&secret).Error; err == nil {
			res.Secret = &types.SettingsSecret{
				ID:   secret.ID,
				Name: secret.Name,
				Type: secret.Type,
				User: secret.User,
			}
		}
	}
	return ws, nil
}

func (r *ResourceManager) Delete(resource *types.PipelineResource) error {
	if err := r.DB.Delete(resource).Error; err != nil {
		return err
	}
	return nil
}

func (r *ResourceManager) Init() {
	var cnt int64
	if err := r.DB.Model(&types.PipelineResource{}).Count(&cnt).Error; err != nil {
		return
	}
	if cnt == 0 {
		now := time.Now()
		for _, res := range InitGlobalResources {
			res.CreateTime = now
			res.UpdateTime = now
			if _, err := r.Create(res); err != nil {
				klog.Info("create resource %s error: %s", res.Name, err.Error())
			}
		}
	}
}
