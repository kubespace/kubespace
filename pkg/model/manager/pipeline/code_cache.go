package pipeline

import (
	"errors"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
	"time"
)

type PipelineCodeCacheManager struct {
	db *gorm.DB
}

func NewPipelineCodeCacheManager(db *gorm.DB) *PipelineCodeCacheManager {
	return &PipelineCodeCacheManager{
		db: db,
	}
}

func (r *PipelineCodeCacheManager) GetById(id uint) (*types.PipelineCodeCache, error) {
	var cache types.PipelineCodeCache
	if err := r.db.First(&cache, "id=?", id).Error; err != nil {
		return nil, err
	}
	return &cache, nil
}

func (r *PipelineCodeCacheManager) GetByWorkspaceId(workspaceId uint) (*types.PipelineCodeCache, error) {
	var cache types.PipelineCodeCache
	if err := r.db.First(&cache, "workspace_id=?", workspaceId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cache, nil
}

func (r *PipelineCodeCacheManager) CreateOrUpdate(workspaceId uint) error {
	var codeTriggersCnt int64
	if err := r.db.Debug().Model(&types.PipelineTrigger{}).Where("pipeline_id in (?)",
		r.db.Table("pipelines").Select("id").Where("workspace_id=?", workspaceId)).Count(
		&codeTriggersCnt).Error; err != nil {
		return err
	}
	codeCache, err := r.GetByWorkspaceId(workspaceId)
	if err != nil {
		return err
	}
	status := types.PipelineCodeCacheStatusClose
	if codeTriggersCnt > 0 {
		status = types.PipelineCodeCacheStatusOpen
	}
	// 有流水线代码更新触发配置，打开代码分支缓存
	if codeCache == nil && status == types.PipelineCodeCacheStatusOpen {
		codeCache = &types.PipelineCodeCache{
			WorkspaceId: workspaceId,
			Status:      types.PipelineCodeCacheStatusOpen,
			CreateTime:  time.Now(),
			UpdateTime:  time.Now(),
		}
		if err = r.db.Omit("id").Create(codeCache).Error; err != nil {
			return err
		}
		return nil
	}
	if codeCache != nil {
		if err = r.db.Model(&types.PipelineCodeCache{}).Where("id=?", codeCache.ID).Updates(&types.PipelineCodeCache{
			Status:     status,
			UpdateTime: time.Now(),
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *PipelineCodeCacheManager) Update(id uint, cache *types.PipelineCodeCache) error {
	return r.db.Model(types.PipelineCodeCache{}).Where("id=?", id).Updates(cache).Error
}
