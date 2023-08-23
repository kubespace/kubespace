package pipeline_run

import (
	"context"
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/model/types"
	utilgit "github.com/kubespace/kubespace/pkg/third/git"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
	"regexp"
	"strings"
	"time"
)

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

func (r *PipelineRunService) Build(pipelineId uint, buildConfig *types.PipelineBuildConfig, username string) *utils.Response {
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

func (r *PipelineRunService) InitialEnvs(pipeline *types.Pipeline, workspace *types.PipelineWorkspace, params *types.PipelineBuildConfig) (map[string]interface{}, error) {
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

func (r *PipelineRunService) InitialCodeEnvs(
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
