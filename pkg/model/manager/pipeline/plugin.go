package pipeline

import (
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

type ManagerPipelinePlugin struct {
	DB *gorm.DB
}

func NewPipelinePluginManager(db *gorm.DB) *ManagerPipelinePlugin {
	return &ManagerPipelinePlugin{DB: db}
}

func (p *ManagerPipelinePlugin) Get(pluginId uint) (*types.PipelinePlugin, error) {
	var plugin types.PipelinePlugin
	if err := p.DB.First(plugin, pluginId).Error; err != nil {
		return nil, err
	}
	return &plugin, nil
}

func (p *ManagerPipelinePlugin) GetByKey(pluginKey string) (*types.PipelinePlugin, error) {
	var plugin types.PipelinePlugin
	if err := p.DB.First(&plugin, "`key` = ?", pluginKey).Error; err != nil {
		return nil, err
	}
	return &plugin, nil
}
