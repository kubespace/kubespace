package pipeline

import (
	"context"
	"fmt"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/manager/pipeline"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/service/pipeline/schemas"
	utilgit "github.com/kubespace/kubespace/pkg/third/git"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"k8s.io/klog/v2"
	"regexp"
	"strings"
	"time"
)

type ServicePipelineRun struct {
	models *model.Models
}

func NewPipelineRunService(models *model.Models) *ServicePipelineRun {
	r := &ServicePipelineRun{
		models: models,
	}
	return r
}

func (r *ServicePipelineRun) ListPipelineRun(pipelineId uint, lastBuildNumber int, status string, limit int) *utils.Response {
	pipelineRuns, err := r.models.PipelineRunManager.ListPipelineRun(pipelineId, lastBuildNumber, status, limit)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	var retData []map[string]interface{}
	for _, pipelineRun := range pipelineRuns {
		stagesRun, err := r.models.PipelineRunManager.StagesRun(pipelineRun.ID)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
		data := map[string]interface{}{
			"pipeline_run": pipelineRun,
			"stages_run":   stagesRun,
		}
		retData = append(retData, data)
	}
	return &utils.Response{Code: code.Success, Data: retData}
}

func (r *ServicePipelineRun) GetPipelineRun(pipelineRunId uint) *utils.Response {
	pipelineRun, err := r.models.PipelineRunManager.Get(pipelineRunId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	stagesRun, err := r.models.PipelineRunManager.StagesRun(pipelineRunId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	pipelineObj, err := r.models.PipelineManager.Get(pipelineRun.PipelineId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	workspace, err := r.models.PipelineWorkspaceManager.Get(pipelineObj.WorkspaceId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	cloneUrl := ""
	if workspace.Code != nil {
		cloneUrl = workspace.Code.CloneUrl
	}
	data := map[string]interface{}{
		"pipeline":     pipelineObj,
		"pipeline_run": pipelineRun,
		"stages_run":   stagesRun,
		"workspace": map[string]interface{}{
			"id":       workspace.ID,
			"name":     workspace.Name,
			"type":     workspace.Type,
			"code_url": cloneUrl,
		},
	}
	return &utils.Response{Code: code.Success, Data: data}
}

// MatchBranchSource 判断是否匹配代码分支源
func MatchBranchSource(sources types.PipelineSources, branch string) bool {
	for _, source := range sources {
		if source.Branch == "" && source.Operator != types.PipelineTriggerOperatorExclude {
			return true
		}
		if source.Operator == types.PipelineTriggerOperatorEqual && source.Branch == branch {
			return true
		}
		if source.Operator == types.PipelineTriggerOperatorExclude && source.Branch == branch {
			return false
		}
		if source.Operator == types.PipelineTriggerOperatorInclude {
			matched, err := regexp.MatchString(source.Branch, branch)
			if err != nil {
				klog.Errorf("regex %s match branch %s error: %s", source.Branch, branch, err.Error())
				continue
			}
			if matched {
				return true
			}
		}
	}
	return false
}

type BuildForPipelineParamsBuilds struct {
	WorkspaceId         uint   `json:"workspace_id"`
	WorkspaceName       string `json:"workspace_name"`
	PipelineId          uint   `json:"pipeline_id"`
	PipelineName        string `json:"pipeline_name"`
	BuildReleaseVersion string `json:"build_release_version"`
	BuildId             uint   `json:"build_id"`
	BuildNumber         uint   `json:"build_number"`
	BuildOperator       string `json:"build_operator"`
	CodeAuthor          string `json:"code_author"`
	CodeBranch          string `json:"code_branch"`
	CodeComment         string `json:"code_comment"`
	CodeCommit          string `json:"code_commit"`
	CodeCommitTime      string `json:"code_commit_time"`
	IsBuild             bool   `json:"is_build" default:"true"`
}

type BuildForPipelineParams struct {
	BuildIds []*BuildForPipelineParamsBuilds `json:"build_ids"`
}

func (r *ServicePipelineRun) InitialEnvs(pipeline *types.Pipeline, workspace *types.PipelineWorkspace, params *types.PipelineBuildConfig) (map[string]interface{}, error) {
	envs := map[string]interface{}{}
	envs[types.PipelineEnvWorkspaceId] = workspace.ID
	envs[types.PipelineEnvWorkspaceName] = workspace.Name
	envs[types.PipelineEnvPipelineId] = pipeline.ID
	envs[types.PipelineEnvPipelineName] = pipeline.Name
	if workspace.Type == types.WorkspaceTypeCode {
		if err := r.InitialCodeEnvs(pipeline, workspace, params.CodeBranch, envs); err != nil {
			return nil, err
		}
	} else if workspace.Type == types.WorkspaceTypeCustom {
		// 合并构建源流水线时要删除的变量
		delPipelineEnvs := []string{
			types.PipelineEnvWorkspaceId,
			types.PipelineEnvWorkspaceName,
			types.PipelineEnvPipelineId,
			types.PipelineEnvPipelineName,
			types.PipelineEnvPipelineBuildNumber,
			types.PipelineEnvPipelineTriggerUser,
		}
		var pipelineBuildId []string
		for _, buildInfo := range params.CustomSources {
			if !buildInfo.IsBuild {
				continue
			}
			build, err := r.models.PipelineRunManager.Get(buildInfo.BuildId)
			if err != nil {
				return nil, fmt.Errorf("获取流水线构建源失败：%s", err.Error())
			}
			pipelineSrc, err := r.models.PipelineManager.Get(build.PipelineId)
			if err != nil {
				return nil, fmt.Errorf("获取流水线源失败：%s", err.Error())
			}
			if build.Status != types.PipelineStatusOK {
				return nil, fmt.Errorf("构建源流水线%s执行状态未完成", pipelineSrc.Name)
			}
			find := false
			for _, source := range pipeline.Sources {
				if pipelineSrc.ID != source.Pipeline {
					continue
				}
				find = true
				break
			}
			if !find {
				return nil, fmt.Errorf("构建参数错误，流水线id=%d不在触发源", build.PipelineId)
			}
			stageRuns, err := r.models.PipelineRunManager.StagesRun(buildInfo.BuildId)
			if err != nil {
				return nil, fmt.Errorf("获取流水线构建源阶段失败：%s", err.Error())
			}
			if len(stageRuns) > 0 {
				lastStage := stageRuns[len(stageRuns)-1]
				for k := range lastStage.Env {
					if utils.Contains(delPipelineEnvs, k) {
						delete(lastStage.Env, k)
					}
				}
				// 合并流水线源变量
				envs = utils.MergeMap(envs, lastStage.Env)
			}
			pipelineBuildId = append(pipelineBuildId, fmt.Sprintf("%d", buildInfo.BuildId))

		}
		envs[types.PipelineEnvPipelineBuildId] = strings.Join(pipelineBuildId, ",")
	}

	return envs, nil
}

func (r *ServicePipelineRun) InitialCodeEnvs(
	pipeline *types.Pipeline,
	workspace *types.PipelineWorkspace,
	codeBranch *types.PipelineBuildCodeBranch,
	envs map[string]interface{}) error {

	if workspace.Code == nil {
		return fmt.Errorf("未获取到流水线空间代码信息")
	}
	if codeBranch.Branch == "" {
		return fmt.Errorf("参数错误，代码分支为空")
	}
	if !MatchBranchSource(pipeline.Sources, codeBranch.Branch) {
		return fmt.Errorf("代码分支未匹配到该流水线")
	}
	envs["PIPELINE_CODE_URL"] = workspace.Code.CloneUrl
	envs["PIPELINE_CODE_API_URL"] = workspace.Code.ApiUrl
	envs["PIPELINE_CODE_TYPE"] = workspace.Code.Type
	envs["PIPELINE_CODE_BRANCH"] = codeBranch.Branch
	if codeBranch.CommitId != "" {
		// 指定分支提交id
		envs["PIPELINE_CODE_COMMIT_ID"] = codeBranch.CommitId
		envs["PIPELINE_CODE_COMMIT_AUTHOR"] = codeBranch.Author
		envs["PIPELINE_CODE_COMMIT_MESSAGE"] = codeBranch.Message
		envs["PIPELINE_CODE_COMMIT_TIME"] = codeBranch.CommitTime
	} else {
		// 获取分支最新提交id
		secret, err := r.models.SettingsSecretManager.Get(workspace.Code.SecretId)
		if err != nil {
			return fmt.Errorf("获取代码密钥失败：" + err.Error())
		}
		gitcli, err := utilgit.NewClient(workspace.Code.Type, workspace.Code.ApiUrl, secret.GetSecret())
		if err != nil {
			return err
		}
		commit, err := gitcli.GetBranchLatestCommit(context.Background(), workspace.Code.CloneUrl, codeBranch.Branch)
		if err != nil {
			return fmt.Errorf("获取远程分支%s失败：%s", codeBranch.Branch, err.Error())
		}
		envs["PIPELINE_CODE_COMMIT_ID"] = commit.CommitId
		envs["PIPELINE_CODE_COMMIT_AUTHOR"] = commit.Author
		envs["PIPELINE_CODE_COMMIT_MESSAGE"] = commit.Message
		envs["PIPELINE_CODE_COMMIT_TIME"] = commit.CommitTime
	}
	return nil
}

func (r *ServicePipelineRun) Build(pipelineId uint, buildConfig *types.PipelineBuildConfig, username string) *utils.Response {
	pipelineObj, err := r.models.PipelineManager.Get(pipelineId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	workspace, err := r.models.PipelineWorkspaceManager.Get(pipelineObj.WorkspaceId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	stages, err := r.models.PipelineManager.Stages(pipelineId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if len(stages) == 0 {
		return &utils.Response{Code: code.DataNotExists, Msg: "当前流水线未配置阶段"}
	}
	envs, err := r.InitialEnvs(pipelineObj, workspace, buildConfig)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	var stagesRun []*types.PipelineRunStage
	for _, stage := range stages {
		stageRun := types.PipelineRunStage{
			Name:         stage.Name,
			TriggerMode:  stage.TriggerMode,
			Status:       types.PipelineStatusWait,
			Env:          map[string]interface{}{},
			CustomParams: stage.CustomParams,
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}
		var stageRunJobs types.PipelineRunJobs
		for _, stageJob := range stage.Jobs {
			stageRunJob := &types.PipelineRunJob{
				Name:      stageJob.Name,
				PluginKey: stageJob.PluginKey,
				Status:    types.PipelineStatusWait,
				Params:    stageJob.Params,
				Env:       map[string]interface{}{},
			}
			stageRunJobs = append(stageRunJobs, stageRunJob)
		}
		stageRun.Jobs = stageRunJobs
		stagesRun = append(stagesRun, &stageRun)
	}
	var paramsMap = make(types.Map)
	if err = utils.ConvertTypeByJson(buildConfig, &paramsMap); err != nil {
		return &utils.Response{Code: code.UnMarshalError, Msg: err.Error()}
	}
	pipelineRun := &types.PipelineRun{
		PipelineId: pipelineId,
		Status:     types.PipelineStatusWait,
		Operator:   username,
		Params:     paramsMap,
		Env:        envs,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	pipelineRun, err = r.models.PipelineRunManager.CreatePipelineRun(pipelineRun, stagesRun)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success, Data: pipelineRun}
}

func (r *ServicePipelineRun) ManualExecuteStage(manualSer *serializers.PipelineStageManualSerializer) *utils.Response {
	stageRun, err := r.models.PipelineRunManager.GetStageRun(manualSer.StageRunId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	for i, job := range stageRun.Jobs {
		for pluginKey, params := range manualSer.JobParams {
			if job.PluginKey == pluginKey && len(params) > 0 {
				stageRun.Jobs[i].Params = params
				break
			}
		}
	}
	now := time.Now()
	if _, _, err = r.models.PipelineRunManager.UpdatePipelineStageRun(&pipeline.UpdateStageObj{
		StageRunId:     stageRun.ID,
		StageRunStatus: types.PipelineStatusDoing,
		StageExecTime:  &now,
		StageRunJobs:   stageRun.Jobs,
		CustomParams:   manualSer.CustomParams,
	}); err != nil {
		return &utils.Response{Code: code.DBError, Msg: "更新阶段任务参数失败:" + err.Error()}
	}
	return &utils.Response{Code: code.Success}
}

func (r *ServicePipelineRun) RetryStage(retrySer *serializers.PipelineStageRetrySerializer) *utils.Response {
	stageRun, err := r.models.PipelineRunManager.GetStageRun(retrySer.StageRunId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if stageRun.Status != types.PipelineStatusError {
		msg := fmt.Sprintf("current stage run id=%v status is %v, not error", stageRun.ID, stageRun.Status)
		return &utils.Response{Code: code.RequestError, Msg: msg}
	}
	now := time.Now()
	_, stageRun, err = r.models.PipelineRunManager.UpdatePipelineStageRun(&pipeline.UpdateStageObj{
		StageRunId:     stageRun.ID,
		StageRunStatus: types.PipelineStatusDoing,
		StageExecTime:  &now,
	})
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success}
}

func (r *ServicePipelineRun) CancelStage(cancelParams *schemas.PipelineStageCancelParams) *utils.Response {
	stageRun, err := r.models.PipelineRunManager.GetStageRun(cancelParams.StageRunId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if stageRun.Status != types.PipelineStatusDoing {
		klog.Infof("current stage run id=%v status is %v, not running", stageRun.ID, stageRun.Status)
		return &utils.Response{Code: code.RequestError, Msg: "当前阶段不在执行中，请刷新重试"}
	}
	// 更新当前阶段状态为取消中
	_, stageRun, err = r.models.PipelineRunManager.UpdatePipelineStageRun(&pipeline.UpdateStageObj{
		StageRunId:     stageRun.ID,
		StageRunStatus: types.PipelineStatusCancel,
	})
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success}
}

// ReExecStage 取消之后重新执行
func (r *ServicePipelineRun) ReExecStage(retrySer *schemas.PipelineStageReexecParams) *utils.Response {
	stageRun, err := r.models.PipelineRunManager.GetStageRun(retrySer.StageRunId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if stageRun.Status != types.PipelineStatusCanceled {
		msg := fmt.Sprintf("current stage run id=%v status is %v, not canceled", stageRun.ID, stageRun.Status)
		return &utils.Response{Code: code.RequestError, Msg: msg}
	}
	now := time.Now()
	_, stageRun, err = r.models.PipelineRunManager.UpdatePipelineStageRun(&pipeline.UpdateStageObj{
		StageRunId:     stageRun.ID,
		StageRunStatus: types.PipelineStatusDoing,
		StageExecTime:  &now,
	})
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success}
}

// JobCallback spacelet节点执行完成任务后进行回调，不写数据库，通知controller-manager
func (r *ServicePipelineRun) JobCallback(params *schemas.JobCallbackParams) *utils.Response {
	jobRun, err := r.models.PipelineRunManager.GetJobRun(params.JobId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	jobRun.Status = params.Status
	// 通知controller-manager
	if err = r.models.PipelineRunManager.NotifyJobRun(jobRun); err != nil {
		return &utils.Response{Code: code.RedisError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success}
}
