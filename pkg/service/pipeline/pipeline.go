package pipeline

import (
	"context"
	"database/sql"
	oerrors "errors"
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	pipelinemgr "github.com/kubespace/kubespace/pkg/model/manager/pipeline"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/service/pipeline/pipeline_run"
	"github.com/kubespace/kubespace/pkg/service/pipeline/schemas"
	"github.com/kubespace/kubespace/pkg/third/git"
	"github.com/kubespace/kubespace/pkg/utils"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
	"time"
)

type PipelineService struct {
	models *model.Models
}

func NewPipelineService(models *model.Models) *PipelineService {
	return &PipelineService{
		models: models,
	}
}

func (p *PipelineService) Create(params *schemas.PipelineParams, user *types.User) (*types.Pipeline, error) {
	workspace, err := p.models.PipelineWorkspaceManager.Get(params.WorkspaceId)
	if err != nil {
		return nil, errors.New(code.DataNotExists, err)
	}
	pipeline := &types.Pipeline{
		Name:        params.Name,
		WorkspaceId: params.WorkspaceId,
		Sources:     params.Sources,
		CreateUser:  user.Name,
		UpdateUser:  user.Name,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
	if len(params.Sources) == 0 {
		return nil, errors.New(code.ParamsError, "流水线触发源不能为空")
	}
	if err = p.CheckSource(workspace, params.Sources); err != nil {
		return nil, err
	}
	var stages []*types.PipelineStage
	for _, stageSer := range params.Stages {
		if stageSer.TriggerMode != types.StageTriggerModeAuto && stageSer.TriggerMode != types.StageTriggerModeManual {
			return nil, errors.New(code.ParamsError, fmt.Sprintf("trigger mode %s is unknown", stageSer.TriggerMode))
		}
		stage := &types.PipelineStage{
			Name:         stageSer.Name,
			TriggerMode:  stageSer.TriggerMode,
			CustomParams: stageSer.CustomParams,
			Jobs:         stageSer.Jobs,
		}
		stages = append(stages, stage)
	}
	triggers, err := p.GetPipelineTrigger(0, params.Triggers, user.Name)
	if err != nil {
		return nil, errors.New(code.ParamsError, "获取触发配置失败: "+err.Error())
	}
	pipeline, err = p.models.PipelineManager.CreatePipeline(pipeline, stages, triggers)
	if err != nil {
		return nil, errors.New(code.CreateError, err)
	}
	if err = p.models.PipelineCodeCacheManager.CreateOrUpdate(workspace); err != nil {
		if delErr := p.models.PipelineManager.Delete(pipeline.ID); delErr != nil {
			klog.Errorf("delete pipeline id=%d error: %s", pipeline.ID, err.Error())
		}
		return nil, errors.New(code.DBError, "更新代码分支缓存失败："+err.Error())
	}
	return pipeline, nil
}

func (p *PipelineService) CheckSource(workspace *types.PipelineWorkspace, sources types.PipelineSources) error {
	triggerWorkspaceIdMap := make(map[uint]struct{})
	for _, source := range sources {
		if workspace.Type == types.WorkspaceTypeCode && source.Type != types.PipelineSourceTypeCode {
			return errors.New(code.ParamsError, fmt.Sprintf("pipeline trigger type %s is wrong", source.Type))
		}
		if workspace.Type == types.WorkspaceTypeCustom {
			if source.Type != types.PipelineSourceTypePipeline {
				return errors.New(code.ParamsError, fmt.Sprintf("pipeline trigger type %s is wrong", source.Type))
			}
			if source.Workspace == 0 {
				return errors.New(code.ParamsError, "流水线触发空间不能为空")
			}
			if source.Pipeline == 0 {
				return errors.New(code.ParamsError, "触发空间流水线不能为空")
			}
			if _, ok := triggerWorkspaceIdMap[source.Workspace]; ok {
				return errors.New(code.ParamsError, "触发流水线空间源不能相同")
			}
			triggerWorkspaceIdMap[source.Workspace] = struct{}{}
		}
	}
	return nil
}

func (p *PipelineService) GetPipelineTrigger(pipelineId uint, triggers []*schemas.PipelineTrigger, username string) ([]*types.PipelineTrigger, error) {
	var triggerObjs []*types.PipelineTrigger
	var err error
	for _, trig := range triggers {
		var triggerObj *types.PipelineTrigger
		if trig.Id != 0 {
			triggerObj, err = p.models.PipelineTriggerManager.Get(trig.Id)
			if err != nil {
				return nil, err
			}
		} else {
			triggerObj = &types.PipelineTrigger{
				ID:         0,
				PipelineId: pipelineId,
				Type:       trig.Type,
				Config:     types.PipelineTriggerConfig{},
				UpdateUser: username,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
			}

			if trig.Type == types.PipelineTriggerTypeCode {
				// 第一次立即触发并初始化分支配置
				triggerObj.NextTriggerTime = &sql.NullTime{Time: time.Now(), Valid: true}
			}
		}
		if trig.Type == types.PipelineTriggerTypeCron {
			triggerObj.Config.Cron = &types.PipelineTriggerConfigCron{Cron: trig.Cron}
			nextTriggerTime, err := utils.NextTriggerTime(trig.Cron)
			if err != nil {
				return nil, err
			}
			triggerObj.NextTriggerTime = &sql.NullTime{Time: nextTriggerTime, Valid: true}
		}

		triggerObjs = append(triggerObjs, triggerObj)
	}
	return triggerObjs, nil
}

func (p *PipelineService) Update(params *schemas.PipelineParams, user *types.User) (*types.Pipeline, error) {
	workspace, err := p.models.PipelineWorkspaceManager.Get(params.WorkspaceId)
	if err != nil {
		return nil, errors.New(code.DBError, err)
	}
	pipeline, err := p.models.PipelineManager.GetById(params.ID)
	if err != nil {
		return nil, errors.New(code.DataNotExists, fmt.Sprintf("获取流水线失败:%s", err.Error()))
	}
	pipeline.Name = params.Name
	pipeline.Sources = params.Sources
	pipeline.UpdateUser = user.Name
	if err = p.CheckSource(workspace, params.Sources); err != nil {
		return pipeline, err
	}
	var stages []*types.PipelineStage
	for _, stageSer := range params.Stages {
		if stageSer.TriggerMode != types.StageTriggerModeAuto && stageSer.TriggerMode != types.StageTriggerModeManual {
			return pipeline, errors.New(code.ParamsError, fmt.Sprintf("trigger mode %s is unknown", stageSer.TriggerMode))
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
	triggers, err := p.GetPipelineTrigger(pipeline.ID, params.Triggers, user.Name)
	if err != nil {
		return pipeline, errors.New(code.ParamsError, "获取触发配置失败: "+err.Error())
	}
	pipeline, err = p.models.PipelineManager.UpdatePipeline(pipeline, stages, triggers)
	if err != nil {
		return pipeline, errors.New(code.DBError, err)
	}
	if err = p.models.PipelineCodeCacheManager.CreateOrUpdate(workspace); err != nil {
		return pipeline, errors.New(code.DBError, "更新代码分支缓存失败："+err.Error())
	}
	return pipeline, nil
}

func (p *PipelineService) GetPipeline(pipelineId uint) *utils.Response {
	pipeline, err := p.models.PipelineManager.GetById(pipelineId)
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
		var sources []*types.PipelineSource
		for i, t := range pipeline.Sources {
			if t.Type == types.PipelineSourceTypePipeline {
				w, err := p.models.PipelineWorkspaceManager.Get(t.Workspace)
				if err != nil {
					if !oerrors.Is(err, gorm.ErrRecordNotFound) {
						return &utils.Response{Code: code.DBError, Msg: err.Error()}
					} else {
						continue
					}
				} else {
					pipeline.Sources[i].WorkspaceName = w.Name
				}
				p, err := p.models.PipelineManager.GetById(t.Pipeline)
				if err != nil {
					if !oerrors.Is(err, gorm.ErrRecordNotFound) {
						return &utils.Response{Code: code.DBError, Msg: err.Error()}
					} else {
						continue
					}
				} else {
					pipeline.Sources[i].PipelineName = p.Name
				}
				sources = append(sources, pipeline.Sources[i])
			}
		}
		pipeline.Sources = sources
	}
	triggerObjs, err := p.models.PipelineTriggerManager.List(&pipelinemgr.PipelineTriggerCondition{PipelineId: pipeline.ID})
	var triggers []*schemas.PipelineTrigger
	for _, obj := range triggerObjs {
		t := &schemas.PipelineTrigger{
			Id:   obj.ID,
			Type: obj.Type,
		}
		if obj.Type == types.PipelineTriggerTypeCron {
			t.Cron = obj.Config.Cron.Cron
		}
		triggers = append(triggers, t)
	}
	data := map[string]interface{}{
		"pipeline":  pipeline,
		"stages":    stages,
		"workspace": workspace,
		"triggers":  triggers,
	}
	return &utils.Response{Code: code.Success, Data: data}
}

func (p *PipelineService) ListPipeline(workspaceId uint) *utils.Response {
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

func (p *PipelineService) ListRepoBranches(pipelineId uint) *utils.Response {
	pipelineObj, err := p.models.PipelineManager.GetById(pipelineId)
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
	gitcli, err := git.NewClient(workspace.Code.Type, workspace.Code.ApiUrl, &types.Secret{
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
		if pipeline_run.MatchBranchSource(pipelineObj.Sources, b.Name) {
			branches = append(branches, repoBranches[i])
		}
	}
	return &utils.Response{Code: code.Success, Data: branches}
}
