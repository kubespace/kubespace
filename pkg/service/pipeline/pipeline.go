package pipeline

import (
	"context"
	"errors"
	"fmt"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/third/git"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"gorm.io/gorm"
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
	if resp := p.CheckTrigger(workspace, pipelineSer); !resp.IsSuccess() {
		return resp
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
	pipeline, err = p.models.PipelineManager.CreatePipeline(pipeline, stages)
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

func (p *ServicePipeline) CheckTrigger(workspace *types.PipelineWorkspace, pipelineSer *serializers.PipelineSerializer) *utils.Response {
	triggerWorkspaceIdMap := make(map[uint]struct{})
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
			if _, ok := triggerWorkspaceIdMap[trigger.Workspace]; ok {
				return &utils.Response{Code: code.ParamsError, Msg: "触发流水线空间源不能相同"}
			}
			triggerWorkspaceIdMap[trigger.Workspace] = struct{}{}
			//if trigger.Stage == 0 {
			//	return &utils.Response{Code: code.ParamsError, Msg: "触发流水线阶段不能为空"}
			//}
		}
	}
	return &utils.Response{Code: code.Success}
}

func (p *ServicePipeline) Update(pipelineSer *serializers.PipelineSerializer, user *types.User) *utils.Response {
	workspace, err := p.models.PipelineWorkspaceManager.Get(pipelineSer.WorkspaceId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	pipeline, err := p.models.PipelineManager.Get(pipelineSer.ID)
	if err != nil {
		return &utils.Response{
			Code: code.DBError,
			Msg:  fmt.Sprintf("获取流水线失败:%s", err.Error()),
		}
	}
	pipeline.Name = pipelineSer.Name
	pipeline.Triggers = pipelineSer.Triggers
	pipeline.UpdateUser = user.Name
	if resp := p.CheckTrigger(workspace, pipelineSer); !resp.IsSuccess() {
		return resp
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
	pipeline, err = p.models.PipelineManager.UpdatePipeline(pipeline, stages)
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
	pipeline, err := p.models.PipelineManager.Get(pipelineId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	workspace, err := p.models.PipelineWorkspaceManager.Get(pipeline.WorkspaceId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	stages, err := p.models.PipelineManager.Stages(pipelineId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if workspace.Type == types.WorkspaceTypeCustom {
		var triggers []*types.PipelineTrigger
		for i, t := range pipeline.Triggers {
			if t.Type == types.PipelineTriggerTypePipeline {
				w, err := p.models.PipelineWorkspaceManager.Get(t.Workspace)
				if err != nil {
					if !errors.Is(err, gorm.ErrRecordNotFound) {
						return &utils.Response{Code: code.DBError, Msg: err.Error()}
					} else {
						continue
					}
				} else {
					pipeline.Triggers[i].WorkspaceName = w.Name
				}
				p, err := p.models.PipelineManager.Get(t.Pipeline)
				if err != nil {
					if !errors.Is(err, gorm.ErrRecordNotFound) {
						return &utils.Response{Code: code.DBError, Msg: err.Error()}
					} else {
						continue
					}
				} else {
					pipeline.Triggers[i].PipelineName = p.Name
				}
				triggers = append(triggers, pipeline.Triggers[i])
			}
		}
		pipeline.Triggers = triggers
	}
	data := map[string]interface{}{
		"pipeline":  pipeline,
		"stages":    stages,
		"workspace": workspace,
	}
	return &utils.Response{Code: code.Success, Data: data}
}

func (p *ServicePipeline) ListPipeline(workspaceId uint) *utils.Response {
	pipelines, err := p.models.PipelineManager.List(workspaceId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: fmt.Sprintf("获取流水线列表错误: %v", err)}
	}
	var retData []map[string]interface{}
	for _, pipeline := range pipelines {
		lastPipelineRun, err := p.models.PipelineRunManager.GetLastPipelineRun(pipeline.ID)
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

func (p *ServicePipeline) ListRepoBranches(pipelineId uint) *utils.Response {
	pipelineObj, err := p.models.PipelineManager.Get(pipelineId)
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: err.Error()}
	}
	workspace, err := p.models.PipelineWorkspaceManager.Get(pipelineObj.WorkspaceId)
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: err.Error()}
	}
	if workspace.Type != types.WorkspaceTypeCode {
		return &utils.Response{Code: code.ParamsError, Msg: "当前流水线空间不是代码空间"}
	}
	if workspace.Code == nil {
		return &utils.Response{Code: code.ParamsError, Msg: "当前流水线代码空间未获取到仓库"}
	}
	secret, err := p.models.SettingsSecretManager.Get(workspace.Code.SecretId)
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: err.Error()}
	}
	gitcli, err := git.NewClient(workspace.Code.Type, workspace.Code.ApiUrl, &git.Secret{
		Type:        secret.Type,
		User:        secret.User,
		Password:    secret.Password,
		PrivateKey:  secret.PrivateKey,
		AccessToken: secret.AccessToken,
	})
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	repoBranches, err := gitcli.ListRepoBranches(context.Background(), workspace.Code.CloneUrl)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	var branches []*git.Reference
	for i, b := range repoBranches {
		if MatchTriggerBranch(pipelineObj.Triggers, b.Name) {
			branches = append(branches, repoBranches[i])
		}
	}
	return &utils.Response{Code: code.Success, Data: branches}
}
