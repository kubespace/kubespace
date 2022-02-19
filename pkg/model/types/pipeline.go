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
	WorkspaceTypeCode   = "code"
	WorkspaceTypeCustom = "custom"
)

const (
	WorkspaceCodeTypeHttps = "https"
	WorkspaceCodeTypeGit = "git"
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
	PipelineTriggerTypeCode = "code"

	PipelineTriggerOperatorEqual   = "equal"
	PipelineTriggerOperatorExclude = "exclude"
	PipelineTriggerOperatorInclude = "include"
)

type PipelineWorkspace struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Description string `gorm:"type:text;" json:"description"`
	Type         string    `gorm:"size:20;not null" json:"type"`
	CodeType string `gorm:"size:20;" json:"code_type"`
	CodeUrl      string    `gorm:"size:512" json:"code_url"`
	CodeSecretId uint  `gorm:"" json:"code_secret_id"`
	CreateUser   string    `gorm:"size:50;not null" json:"create_user"`
	UpdateUser   string    `gorm:"size:50;not null" json:"update_user"`
	CreateTime   time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime   time.Time `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}

type Pipeline struct {
	ID          uint             `gorm:"primaryKey" json:"id"`
	Name        string           `gorm:"size:50;not null;uniqueIndex:idx_workspace_name" json:"name"`
	WorkspaceId uint             `gorm:"not null;uniqueIndex:idx_workspace_name" json:"workspace_id"`
	Triggers    PipelineTriggers `gorm:"type:json" json:"triggers"`
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

type PipelineTrigger struct {
	Type        string                      `json:"type"`
	Expressions []PipelineTriggerExpression `json:"expressions"`
}

type PipelineTriggerExpression struct {
	Key      string `json:"key"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type PipelineStage struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"size:50;not null;uniqueIndex:idx_pipeline_stage_name" json:"name"`
	PipelineId  uint         `gorm:"not null;uniqueIndex:idx_pipeline_stage_name" json:"pipeline_id"`
	TriggerMode string       `gorm:"size:20;not null;" json:"trigger_mode"`
	PrevStageId uint         `gorm:"not null" json:"prev_stage_id"`
	Jobs        PipelineJobs `gorm:"type:json;not null" json:"jobs"`
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

type PipelinePlugin struct {
	ID         uint                    `gorm:"primaryKey"`
	Name       string                  `gorm:"size:255;not null;uniqueIndex:idx_plugin_name"`
	Key        string                  `gorm:"size:50;not null;uniqueIndex:idx_plugin_key"`
	Url        string                  `gorm:"size:255;not null"`
	Params     PipelinePluginParams    `gorm:"type:json;not null"`
	ResultEnv  PipelinePluginResultEnv `gorm:"type:json;"`
	CreateTime time.Time               `gorm:"not null;autoCreateTime"`
	UpdateTime time.Time               `gorm:"not null;autoUpdateTime"`
}

type PipelinePluginParams struct {
	Params []*PipelinePluginParamsSpec `json:"params"`
}

const (
	PluginParamsFromEnv = "env"
	PluginParamsFromJob = "job"
)

type PipelinePluginParamsSpec struct {
	ParamName string `json:"param_name"`
	From      string `json:"from"`
	FromName  string `json:"from_name"`
	Default interface{} `json:"default"`
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
	Jobs           PipelineRunJobs `gorm:"-" json:"jobs"`
	ExecTime     time.Time       `gorm:"not null;autoCreateTime" json:"exec_time"`
	CreateTime     time.Time       `gorm:"not null;autoCreateTime" json:"create_time"`
	UpdateTime     time.Time       `gorm:"not null;autoUpdateTime" json:"update_time"`
}

type PipelineRunJobs []*PipelineRunJob

type PipelineRunJob struct {
	ID         uint            `gorm:"primaryKey" json:"id"`
	StageRunId uint            `gorm:"not null" json:"stage_run_id"`
	Name       string          `gorm:"size:50;not null" json:"name"`
	PluginKey  string          `gorm:"size:255;not null" json:"plugin_key"`
	Status     string          `gorm:"size:50;not null" json:"status"`
	Params     Map             `gorm:"type:json;not null" json:"params"`
	Result     *utils.Response `gorm:"type:json;" json:"result"`
	CreateTime time.Time       `gorm:"not null;autoCreateTime" json:"create_time"`
	UpdateTime time.Time       `gorm:"not null;autoUpdateTime" json:"update_time"`
}

type PipelineRunJobLog struct {
	ID         uint      `gorm:"primaryKey"`
	JobRunId   uint      `gorm:"column:job_run_id;not null"`
	JobName    string    `gorm:"not null"`
	Logs       string    `gorm:"type:longtext"`
	CreateTime time.Time `gorm:"not null;autoCreateTime"`
	UpdateTime time.Time `gorm:"not null;autoUpdateTime"`
}