package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kubespace/kubespace/pkg/utils"
	"time"
)

const (
	WorkspaceTypeCode     = "code"
	WorkspaceTypePipeline = "pipeline"
)

const (
	WorkspaceCodeTypeHttps = "https"
	WorkspaceCodeTypeGit   = "git"
)

const (
	StageTriggerModeAuto   = "auto"
	StageTriggerModeManual = "manual"
)

const (
	PipelineStatusWait   = "wait"
	PipelineStatusDoing  = "doing"
	PipelineStatusCancel = "cancel"
	PipelineStatusOK     = "ok"
	PipelineStatusError  = "error"
	PipelineStatusPause  = "pause"
)

const (
	PipelineTriggerTypeCode     = "code"
	PipelineTriggerTypePipeline = "pipeline"

	PipelineTriggerOperatorEqual   = "equal"
	PipelineTriggerOperatorExclude = "exclude"
	PipelineTriggerOperatorInclude = "regex"
)

type PipelineWorkspace struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Description  string    `gorm:"type:text;" json:"description"`
	Type         string    `gorm:"size:20;not null" json:"type"`
	CodeType     string    `gorm:"size:20;" json:"code_type"`
	CodeUrl      string    `gorm:"size:512" json:"code_url"`
	CodeSecretId uint      `gorm:"" json:"code_secret_id"`
	CreateUser   string    `gorm:"size:50;not null" json:"create_user"`
	UpdateUser   string    `gorm:"size:50;not null" json:"update_user"`
	CreateTime   time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime   time.Time `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}

type PipelineWorkspaceRelease struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	WorkspaceId    uint      `gorm:"not null;uniqueIndex:idx_workspace_version" json:"workspace_id"`
	ReleaseVersion string    `gorm:"size:500;not null;uniqueIndex:idx_workspace_version" json:"release_version"`
	JobRunId       uint      `gorm:"not null;" json:"job_run_id"`
	CreateTime     time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime     time.Time `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}

type Pipeline struct {
	ID          uint             `gorm:"primaryKey" json:"id"`
	Name        string           `gorm:"size:50;not null;uniqueIndex:idx_workspace_name" json:"name"`
	WorkspaceId uint             `gorm:"not null;uniqueIndex:idx_workspace_name" json:"workspace_id"`
	Triggers    PipelineTriggers `gorm:"type:json" json:"triggers"`
	Stages      []*PipelineStage `gorm:"-" json:"stages"`
	CreateUser  string           `gorm:"size:50;not null" json:"create_user"`
	UpdateUser  string           `gorm:"size:50;not null" json:"update_user"`
	CreateTime  time.Time        `gorm:"not null;autoCreateTime" json:"create_time"`
	UpdateTime  time.Time        `gorm:"not null;autoUpdateTime" json:"update_time"`
}

type PipelineTriggers []*PipelineTrigger

func (pt *PipelineTriggers) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to convert to bytes:", value))
	}
	err := json.Unmarshal(bytes, pt)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bytes: %s", string(bytes))
	}
	return nil
}

// Value return json value, implement driver.Valuer interface
func (pt PipelineTriggers) Value() (driver.Value, error) {
	bytes, err := json.Marshal(pt)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

const (
	PipelineBranchTypeBranch  = "branch"
	PipelineBranchTypeRequest = "request"
)

type PipelineTrigger struct {
	Type       string `json:"type"`
	Workspace  uint   `json:"workspace"`
	Pipeline   uint   `json:"pipeline"`
	Stage      uint   `json:"stage"`
	BranchType string `json:"branch_type"`
	Operator   string `json:"operator"`
	Branch     string `json:"branch"`
}

type PipelineStage struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Name       string `gorm:"size:50;not null;uniqueIndex:idx_pipeline_stage_name" json:"name"`
	PipelineId uint   `gorm:"not null;uniqueIndex:idx_pipeline_stage_name" json:"pipeline_id"`
	// 在流水线中对阶段的自定义参数，执行时自动放到阶段的env中
	CustomParams Map          `gorm:"type:json" json:"custom_params"`
	TriggerMode  string       `gorm:"size:20;not null;" json:"trigger_mode"`
	PrevStageId  uint         `gorm:"not null" json:"prev_stage_id"`
	Jobs         PipelineJobs `gorm:"type:json;not null" json:"jobs"`
}

func (pj *PipelineJobs) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to convert to bytes:", value))
	}
	err := json.Unmarshal(bytes, pj)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bytes: %s", string(bytes))
	}
	return nil
}

// Value return json value, implement driver.Valuer interface
func (pj PipelineJobs) Value() (driver.Value, error) {
	bytes, err := json.Marshal(pj)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

type PipelineJobs []*PipelineJob

type PipelineJob struct {
	Name      string                 `json:"name"`
	PluginKey string                 `json:"plugin_key"`
	Params    map[string]interface{} `json:"params"`
}

const (
	BuiltinPluginBuildCodeToImage = "build_code_to_image"
	BuiltinPluginExecuteShell     = "execute_shell"

	// BuiltinPluginUpgradeApp 根据构建出来的代码镜像，升级项目应用
	BuiltinPluginUpgradeApp = "upgrade_app"
	// BuiltinPluginRelease 发布版本，给代码及镜像打发布tag号
	BuiltinPluginRelease = "release"
	// BuiltinPluginDeployK8s 替换镜像，并部署k8s资源
	BuiltinPluginDeployK8s = "deploy_k8s"
)

const PipelinePluginBuiltinUrl = "builtin"

type PipelinePlugin struct {
	ID         uint                    `gorm:"primaryKey"`
	Name       string                  `gorm:"size:255;not null;uniqueIndex:idx_plugin_name"`
	Key        string                  `gorm:"size:50;not null;uniqueIndex:idx_plugin_key"`
	Url        string                  `gorm:"size:255;not null"`
	Params     PipelinePluginParams    `gorm:"type:json;not null"`
	Version    string                  `gorm:"version"`
	ResultEnv  PipelinePluginResultEnv `gorm:"type:json;"`
	CreateTime time.Time               `gorm:"not null;autoCreateTime"`
	UpdateTime time.Time               `gorm:"not null;autoUpdateTime"`
}

type PipelinePluginParams struct {
	Params []*PipelinePluginParamsSpec `json:"params"`
}

const (
	PluginParamsFromEnv              = "env"
	PluginParamsFromJob              = "job"
	PluginParamsFromCodeSecret       = "code_secret"
	PluginParamsFromImageRegistry    = "image_registry"
	PluginParamsFromPipelineResource = "pipeline_resource"
	PluginParamsFromPipelineEnv      = "pipeline_env"
)

type PipelinePluginParamsSpec struct {
	ParamName string      `json:"param_name"`
	From      string      `json:"from"`
	FromName  string      `json:"from_name"`
	Default   interface{} `json:"default"`
}

func (p *PipelinePluginParams) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to convert to bytes:", value))
	}
	err := json.Unmarshal(bytes, p)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bytes: %s", string(bytes))
	}
	return nil
}

// Value return json value, implement driver.Valuer interface
func (p PipelinePluginParams) Value() (driver.Value, error) {
	bytes, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

type PipelinePluginResultEnv struct {
	EnvPath []*PipelinePluginResultEnvPath `json:"env_path"`
}

type PipelinePluginResultEnvPath struct {
	EnvName    string `json:"env_name"`
	ResultName string `json:"result_name"`
}

func (p *PipelinePluginResultEnv) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to convert to bytes:", value))
	}
	err := json.Unmarshal(bytes, p)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bytes: %s", string(bytes))
	}
	return nil
}

// Value return json value, implement driver.Valuer interface
func (p PipelinePluginResultEnv) Value() (driver.Value, error) {
	bytes, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

type Map map[string]interface{}

func (m *Map) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to convert to bytes:", value))
	}
	err := json.Unmarshal(bytes, m)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bytes: %s", string(bytes))
	}
	return nil
}

// Value return json value, implement driver.Valuer interface
func (m Map) Value() (driver.Value, error) {
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

type PipelineRun struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	PipelineId  uint      `gorm:"not null;uniqueIndex:idx_pipeline_build_number" json:"pipeline_id"`
	BuildNumber uint      `gorm:"not null;uniqueIndex:idx_pipeline_build_number" json:"build_number"`
	Status      string    `gorm:"size:50;not null" json:"status"`
	Env         Map       `gorm:"type:json" json:"env"`
	Operator    string    `gorm:"size:50;not null" json:"operator"`
	CreateTime  time.Time `gorm:"not null;autoCreateTime" json:"create_time"`
	UpdateTime  time.Time `gorm:"not null;autoUpdateTime" json:"update_time"`
}

type PipelineRunStage struct {
	ID             uint            `gorm:"primaryKey" json:"id"`
	Name           string          `gorm:"name;size:255;not null" json:"name"`
	TriggerMode    string          `gorm:"size:20;not null" json:"trigger_mode"`
	PrevStageRunId uint            `gorm:"not null" json:"prev_stage_run_id"`
	PipelineRunId  uint            `gorm:"not null" json:"pipeline_run_id"`
	Status         string          `gorm:"size:50;not null" json:"status"`
	Env            Map             `gorm:"type:json" json:"env"`
	CustomParams   Map             `gorm:"json" json:"custom_params"`
	Jobs           PipelineRunJobs `gorm:"-" json:"jobs"`
	ExecTime       time.Time       `gorm:"not null;autoCreateTime" json:"exec_time"`
	CreateTime     time.Time       `gorm:"not null;autoCreateTime" json:"create_time"`
	UpdateTime     time.Time       `gorm:"not null;autoUpdateTime" json:"update_time"`
}

type PipelineRunJobs []*PipelineRunJob

type PipelineRunJob struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	PipelineRunId uint   `gorm:"not null" json:"pipeline_run_id"`
	StageRunId    uint   `gorm:"not null" json:"stage_run_id"`
	Name          string `gorm:"size:50;not null" json:"name"`
	PluginKey     string `gorm:"size:255;not null" json:"plugin_key"`
	Status        string `gorm:"size:50;not null" json:"status"`
	// 每个Job执行完之后的环境变量
	Env        Map             `gorm:"type:json" json:"env"`
	Params     Map             `gorm:"type:json;not null" json:"params"`
	Result     *utils.Response `gorm:"type:json;" json:"result"`
	CreateTime time.Time       `gorm:"not null;autoCreateTime" json:"create_time"`
	UpdateTime time.Time       `gorm:"not null;autoUpdateTime" json:"update_time"`
}

type PipelineRunJobLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	JobRunId   uint      `gorm:"column:job_run_id;not null" json:"job_run_id"`
	Logs       string    `gorm:"type:longtext" json:"logs"`
	CreateTime time.Time `gorm:"not null;autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"not null;autoUpdateTime" json:"update_time"`
}

type PipelineResource struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	WorkspaceId uint            `gorm:"not null;uniqueIndex:idx_workspace_resource" json:"workspace_id"`
	Name        string          `gorm:"size:255;not null;uniqueIndex:idx_workspace_resource" json:"name"`
	Global      bool            `gorm:"default:false" json:"global"`
	Type        string          `gorm:"size:50;not null" json:"type"`
	Value       string          `gorm:"size:500; not null;" json:"value"`
	SecretId    uint            `gorm:"" json:"secret_id"`
	Secret      *SettingsSecret `gorm:"-" json:"secret"`
	Description string          `gorm:"size:2000" json:"description"`
	CreateUser  string          `gorm:"size:50;not null" json:"create_user"`
	UpdateUser  string          `gorm:"size:50;not null" json:"update_user"`
	CreateTime  time.Time       `gorm:"not null;autoCreateTime" json:"create_time"`
	UpdateTime  time.Time       `gorm:"not null;autoUpdateTime" json:"update_time"`
}
