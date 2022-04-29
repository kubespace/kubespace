package pipeline

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"time"
)

type ServicePipeline struct {
	models *model.Models
}

func NewPipelineService(models *model.Models) *ServicePipeline {
	return &ServicePipeline{
		models: models,
	}
}

func (p *ServicePipeline) Create(pipelineSer *serializers.PipelineSerializer, user *types.User) *utils.Response {
	workspace, err := p.models.PipelineWorkspaceManager.Get(pipelineSer.WorkspaceId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	pipeline := &types.Pipeline{
		Name:        pipelineSer.Name,
		WorkspaceId: pipelineSer.WorkspaceId,
		Triggers:    pipelineSer.Triggers,
		CreateUser:  user.Name,
		UpdateUser:  user.Name,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
	if len(pipelineSer.Triggers) == 0 {
		return &utils.Response{Code: code.ParamsError, Msg: "流水线触发源不能为空"}
	}
	for _, trigger := range pipelineSer.Triggers {
		if workspace.Type == types.WorkspaceTypeCode && trigger.Type != types.PipelineTriggerTypeCode {
			return &utils.Response{
				Code: code.ParamsError,
				Msg:  fmt.Sprintf("pipeline trigger type %s is wrong", trigger.Type),
			}
		}
		if workspace.Type == types.WorkspaceTypeCustom {
			if trigger.Type != types.PipelineTriggerTypePipeline {
				return &utils.Response{
					Code: code.ParamsError,
					Msg:  fmt.Sprintf("pipeline trigger type %s is wrong", trigger.Type),
				}
			}
			if trigger.Workspace == 0 {
				return &utils.Response{Code: code.ParamsError, Msg: "流水线触发空间不能为空"}
			}
			if trigger.Pipeline == 0 {
				return &utils.Response{Code: code.ParamsError, Msg: "触发空间流水线不能为空"}
			}
			//if trigger.Stage == 0 {
			//	return &utils.Response{Code: code.ParamsError, Msg: "触发流水线阶段不能为空"}
			//}
		}
	}
	var stages []*types.PipelineStage
	for _, stageSer := range pipelineSer.Stages {
		if stageSer.TriggerMode != types.StageTriggerModeAuto && stageSer.TriggerMode != types.StageTriggerModeManual {
			return &utils.Response{
				Code: code.ParamsError,
				Msg:  fmt.Sprintf("trigger mode %s is unknown", stageSer.TriggerMode),
			}
		}
		stage := &types.PipelineStage{
			Name:         stageSer.Name,
			TriggerMode:  stageSer.TriggerMode,
			CustomParams: stageSer.CustomParams,
			Jobs:         stageSer.Jobs,
		}
		stages = append(stages, stage)
	}
	pipeline, err = p.models.ManagerPipeline.CreatePipeline(pipeline, stages)
	if err != nil {
		return &utils.Response{
			Code: code.DBError,
			Msg:  err.Error(),
		}
	}
	return &utils.Response{
		Code: code.Success,
		Data: pipeline,
	}
}

func (p *ServicePipeline) Update(pipelineSer *serializers.PipelineSerializer, user *types.User) *utils.Response {
	workspace, err := p.models.PipelineWorkspaceManager.Get(pipelineSer.WorkspaceId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	pipeline, err := p.models.ManagerPipeline.Get(pipelineSer.ID)
	if err != nil {
		return &utils.Response{
			Code: code.DBError,
			Msg:  fmt.Sprintf("获取流水线失败:%s", err.Error()),
		}
	}
	pipeline.Name = pipelineSer.Name
	pipeline.Triggers = pipelineSer.Triggers
	pipeline.UpdateUser = user.Name
	for _, trigger := range pipelineSer.Triggers {
		if workspace.Type == types.WorkspaceTypeCode && trigger.Type != types.PipelineTriggerTypeCode {
			return &utils.Response{
				Code: code.ParamsError,
				Msg:  fmt.Sprintf("pipeline trigger type %s is wrong", trigger.Type),
			}
		}
		if workspace.Type == types.WorkspaceTypeCustom {
			if trigger.Type != types.PipelineTriggerTypePipeline {
				return &utils.Response{
					Code: code.ParamsError,
					Msg:  fmt.Sprintf("pipeline trigger type %s is wrong", trigger.Type),
				}
			}
			if trigger.Workspace == 0 {
				return &utils.Response{Code: code.ParamsError, Msg: "流水线触发空间不能为空"}
			}
			if trigger.Pipeline == 0 {
				return &utils.Response{Code: code.ParamsError, Msg: "触发空间流水线不能为空"}
			}
			//if trigger.Stage == 0 {
			//	return &utils.Response{Code: code.ParamsError, Msg: "触发流水线阶段不能为空"}
			//}
		}
	}
	var stages []*types.PipelineStage
	for _, stageSer := range pipelineSer.Stages {
		if stageSer.TriggerMode != types.StageTriggerModeAuto && stageSer.TriggerMode != types.StageTriggerModeManual {
			return &utils.Response{
				Code: code.ParamsError,
				Msg:  fmt.Sprintf("trigger mode %s is unknown", stageSer.TriggerMode),
			}
		}
		stage := &types.PipelineStage{
			ID:           stageSer.ID,
			Name:         stageSer.Name,
			TriggerMode:  stageSer.TriggerMode,
			CustomParams: stageSer.CustomParams,
			Jobs:         stageSer.Jobs,
		}
		stages = append(stages, stage)
	}
	pipeline, err = p.models.ManagerPipeline.UpdatePipeline(pipeline, stages)
	if err != nil {
		return &utils.Response{
			Code: code.DBError,
			Msg:  err.Error(),
		}
	}
	return &utils.Response{
		Code: code.Success,
		Data: pipeline,
	}
}

func (p *ServicePipeline) GetPipeline(pipelineId uint) *utils.Response {
	pipeline, err := p.models.ManagerPipeline.Get(pipelineId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	workspace, err := p.models.PipelineWorkspaceManager.Get(pipeline.WorkspaceId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	stages, err := p.models.ManagerPipeline.Stages(pipelineId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	data := map[string]interface{}{
		"pipeline": pipeline,
		"stages":   stages,
		"workspace": map[string]interface{}{
			"id":       workspace.ID,
			"name":     workspace.Name,
			"type":     workspace.Type,
			"code_url": workspace.CodeUrl,
		},
	}
	return &utils.Response{Code: code.Success, Data: data}
}

func (p *ServicePipeline) ListPipeline(workspaceId uint) *utils.Response {
	pipelines, err := p.models.ManagerPipeline.List(workspaceId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: fmt.Sprintf("获取流水线列表错误: %v", err)}
	}
	var retData []map[string]interface{}
	for _, pipeline := range pipelines {
		lastPipelineRun, err := p.models.ManagerPipelineRun.GetLastPipelineRun(pipeline.ID)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: fmt.Sprintf("获取流水线构建列表错误: %v", err)}
		}
		data := map[string]interface{}{
			"pipeline":   pipeline,
			"last_build": lastPipelineRun,
		}
		retData = append(retData, data)
	}
	return &utils.Response{Code: code.Success, Data: retData}
}
