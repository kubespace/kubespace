package pipeline

import (
	"github.com/kubespace/kubespace/pkg/conf"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
	"k8s.io/klog"
	"time"
)

type ManagerPipelinePlugin struct {
	DB *gorm.DB
}

func NewPipelinePluginManager(db *gorm.DB) *ManagerPipelinePlugin {
	p := &ManagerPipelinePlugin{DB: db}
	p.Init()
	return p
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

var BuiltinPlugins = []types.PipelinePlugin{
	{
		Name: "构建代码镜像",
		Key:  types.BuiltinPluginBuildCodeToImage,
		Url:  conf.AppConfig.PipelinePluginUrl + "/" + types.BuiltinPluginBuildCodeToImage,
		Params: types.PipelinePluginParams{
			Params: []*types.PipelinePluginParamsSpec{
				{
					ParamName: "code_url",
					From:      types.PluginParamsFromEnv,
					FromName:  "PIPELINE_CODE_URL",
					Default:   "",
				},
				{
					ParamName: "code_branch",
					From:      types.PluginParamsFromEnv,
					FromName:  "PIPELINE_CODE_BRANCH",
					Default:   "",
				},
				{
					ParamName: "code_commit_id",
					From:      types.PluginParamsFromEnv,
					FromName:  "PIPELINE_CODE_COMMIT_ID",
					Default:   "",
				},
				{
					ParamName: "code_secret",
					From:      types.PluginParamsFromCodeSecret,
					FromName:  "",
					Default:   nil,
				},
				{
					ParamName: "code_build",
					From:      types.PluginParamsFromJob,
					FromName:  "code_build",
					Default:   true,
				},
				{
					ParamName: "code_build_type",
					From:      types.PluginParamsFromJob,
					FromName:  "code_build_type",
					Default:   "file",
				},
				{
					ParamName: "code_build_image",
					From:      types.PluginParamsFromPipelineResource,
					FromName:  "code_build_image",
					Default:   "",
				},
				{
					ParamName: "code_build_file",
					From:      types.PluginParamsFromJob,
					FromName:  "code_build_file",
					Default:   "build.sh",
				},
				{
					ParamName: "code_build_script",
					From:      types.PluginParamsFromJob,
					FromName:  "code_build_script",
					Default:   "",
				},
				{
					ParamName: "code_build_exec",
					From:      types.PluginParamsFromJob,
					FromName:  "code_build_exec",
					Default:   "",
				},
				{
					ParamName: "image_build_registry",
					From:      types.PluginParamsFromImageRegistry,
					FromName:  "image_build_registry",
					Default:   "",
				},
				{
					ParamName: "image_registry_id",
					From:      types.PluginParamsFromJob,
					FromName:  "image_build_registry",
					Default:   "",
				},
				{
					ParamName: "image_builds",
					From:      types.PluginParamsFromJob,
					FromName:  "image_builds",
					Default:   nil,
				},
			},
		},
		ResultEnv: types.PipelinePluginResultEnv{
			EnvPath: []*types.PipelinePluginResultEnvPath{
				{
					ResultName: "images",
					EnvName:    "CODE_BUILD_IMAGES",
				},
				{
					ResultName: "image_registry",
					EnvName:    "CODE_BUILD_REGISTRY",
				},
				{
					ResultName: "image_registry_id",
					EnvName:    "CODE_BUILD_REGISTRY_ID",
				},
			},
		},
	},
	{
		Name: "执行shell脚本",
		Key:  types.BuiltinPluginExecuteShell,
		Url:  conf.AppConfig.PipelinePluginUrl + "/" + types.BuiltinPluginExecuteShell,
		Params: types.PipelinePluginParams{
			Params: []*types.PipelinePluginParamsSpec{
				{
					ParamName: "resource",
					From:      types.PluginParamsFromPipelineResource,
					FromName:  "resource",
					Default:   nil,
				},
				{
					ParamName: "port",
					From:      types.PluginParamsFromJob,
					FromName:  "port",
					Default:   "22",
				},
				{
					ParamName: "script",
					From:      types.PluginParamsFromJob,
					FromName:  "script",
					Default:   "",
				},
				{
					ParamName: "shell",
					From:      types.PluginParamsFromJob,
					FromName:  "shell",
					Default:   "bash",
				},
				{
					ParamName: "env",
					From:      types.PluginParamsFromPipelineEnv,
					FromName:  "",
					Default:   nil,
				},
			},
		},
	},
	{
		Name: "升级空间应用",
		Key:  types.BuiltinPluginUpgradeApp,
		Url:  types.PipelinePluginBuiltinUrl,
		Params: types.PipelinePluginParams{
			Params: []*types.PipelinePluginParamsSpec{
				{
					ParamName: "project",
					From:      types.PluginParamsFromJob,
					FromName:  "project",
					Default:   "",
				},
				{
					ParamName: "images",
					From:      types.PluginParamsFromEnv,
					FromName:  "CODE_BUILD_IMAGES",
					Default:   "",
				},
				{
					ParamName: "apps",
					From:      types.PluginParamsFromJob,
					FromName:  "apps",
					Default:   "",
				},
				{
					ParamName: "with_install",
					From:      types.PluginParamsFromJob,
					FromName:  "with_install",
					Default:   "",
				},
			},
		},
	},
	{
		Name: "版本发布",
		Key:  types.BuiltinPluginRelease,
		Url:  conf.AppConfig.PipelinePluginUrl + "/" + types.BuiltinPluginRelease,
		Params: types.PipelinePluginParams{
			Params: []*types.PipelinePluginParamsSpec{
				{
					ParamName: "workspace_id",
					From:      types.PluginParamsFromEnv,
					FromName:  "PIPELINE_WORKSPACE_ID",
					Default:   "",
				},
				{
					ParamName: "code_url",
					From:      types.PluginParamsFromEnv,
					FromName:  "PIPELINE_CODE_URL",
					Default:   "",
				},
				{
					ParamName: "code_branch",
					From:      types.PluginParamsFromEnv,
					FromName:  "PIPELINE_CODE_BRANCH",
					Default:   "",
				},
				{
					ParamName: "code_commit_id",
					From:      types.PluginParamsFromEnv,
					FromName:  "PIPELINE_CODE_COMMIT_ID",
					Default:   "",
				},
				{
					ParamName: "code_secret",
					From:      types.PluginParamsFromCodeSecret,
					FromName:  "",
					Default:   nil,
				},
				{
					ParamName: "version",
					From:      types.PluginParamsFromJob,
					FromName:  "version",
					Default:   nil,
				},
				{
					ParamName: "images",
					From:      types.PluginParamsFromEnv,
					FromName:  "CODE_BUILD_IMAGES",
					Default:   "",
				},
				{
					ParamName: "image_registry",
					From:      types.PluginParamsFromImageRegistry,
					FromName:  "",
					Default:   nil,
				},
			},
		},
		ResultEnv: types.PipelinePluginResultEnv{
			EnvPath: []*types.PipelinePluginResultEnvPath{
				{
					ResultName: "version",
					EnvName:    "RELEASE_VERSION",
				},
				{
					ResultName: "images",
					EnvName:    "CODE_BUILD_IMAGES",
				},
			},
		},
	},
}

func (p *ManagerPipelinePlugin) Init() {
	now := time.Now()
	for _, plugin := range BuiltinPlugins {
		var cnt int64
		if err := p.DB.Model(&types.PipelinePlugin{}).Where("`key` = ?", plugin.Key).Count(&cnt).Error; err != nil {
			klog.Errorf("get plugin %s error: %s", plugin.Key, err.Error())
			return
		}
		if cnt == 0 {
			if plugin.Url != types.PipelinePluginBuiltinUrl {
				plugin.Url = conf.AppConfig.PipelinePluginUrl + "/" + plugin.Key
			}
			plugin.CreateTime = now
			plugin.UpdateTime = now
			if err := p.DB.Create(&plugin).Error; err != nil {
				klog.Info("create pipeline plugin %s=%s error: %s", plugin.Key, plugin.Name, err.Error())
			}
		}
	}

}
