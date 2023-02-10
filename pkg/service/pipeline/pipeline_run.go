package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	sshgit "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/manager/pipeline"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	utilgit "github.com/kubespace/kubespace/pkg/utils/git"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog/v2"
	"os"
	"regexp"
	"strings"
	"time"
)

type codeCommit struct {
	CommitId   string
	Author     string
	Message    string
	CommitTime time.Time
}

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

func (r *ServicePipelineRun) getCodeAuth(secretId uint) (transport.AuthMethod, error) {
	secret, err := r.models.SettingsSecretManager.Get(secretId)
	if err != nil {
		return nil, fmt.Errorf("获取代码密钥失败：" + err.Error())
	}
	var auth transport.AuthMethod
	if secret.Type == types.SettingsSecretTypeKey {
		privateKey, err := sshgit.NewPublicKeys("git", []byte(secret.PrivateKey), "")
		if err != nil {
			return nil, fmt.Errorf("生成代码密钥失败：" + err.Error())
		}
		privateKey.HostKeyCallbackHelper = sshgit.HostKeyCallbackHelper{
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		auth = privateKey
	} else if secret.Type == types.SettingsSecretTypePassword {
		auth = &http.BasicAuth{
			Username: secret.User,
			Password: secret.Password,
		}
	}
	return auth, nil
}

func (r *ServicePipelineRun) getCodeBranchCommit(codeUrl, branch string, secretId uint) (*codeCommit, error) {
	auth, err := r.getCodeAuth(secretId)
	if err != nil {
		return nil, err
	}
	uuid := utils.CreateUUID()
	refName := "refs/heads/" + branch
	ref, err := git.PlainClone("/tmp/"+uuid, true, &git.CloneOptions{
		Auth:            auth,
		URL:             codeUrl,
		Progress:        os.Stdout,
		ReferenceName:   plumbing.ReferenceName(refName),
		SingleBranch:    true,
		Depth:           1,
		NoCheckout:      true,
		InsecureSkipTLS: true,
	})
	if err != nil {
		klog.Errorf("git clone %s error: %v", codeUrl, err)
		return nil, err
	}
	defer os.RemoveAll("/tmp/" + uuid)
	commits, err := ref.Log(&git.LogOptions{})
	if err != nil {
		klog.Errorf("git log %s error: %v", codeUrl, err)
		return nil, err
	}
	commit, err := commits.Next()
	if err != nil {
		klog.Errorf("git log %s error: %v", codeUrl, err)
		return nil, err
	}
	return &codeCommit{
		CommitId:   commit.Hash.String(),
		Author:     commit.Author.Name,
		Message:    commit.Message,
		CommitTime: commit.Author.When,
	}, nil
}

func (r *ServicePipelineRun) getCodeBranchCommitId(codeUrl, branch string, secretId uint) (string, error) {
	auth, err := r.getCodeAuth(secretId)
	if err != nil {
		return "", err
	}
	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{codeUrl},
	})
	refs, err := rem.List(&git.ListOptions{Auth: auth, InsecureSkipTLS: true})
	if err != nil {
		return "", fmt.Errorf("获取代码远程分支" + branch + "失败：" + err.Error())
	}
	for _, ref := range refs {
		if ref.Name().IsBranch() && ref.Name().Short() == branch {
			return ref.Hash().String(), nil
		}
	}
	return "", fmt.Errorf("获取代码远程分支失败：未找到%s分支", branch)
}

func (r *ServicePipelineRun) MatchTriggerBranch(triggers types.PipelineTriggers, branch string) bool {
	for _, trigger := range triggers {
		if trigger.Branch == "" && trigger.Operator != types.PipelineTriggerOperatorExclude {
			return true
		}
		if trigger.Operator == types.PipelineTriggerOperatorEqual && trigger.Branch == branch {
			return true
		}
		if trigger.Operator == types.PipelineTriggerOperatorExclude && trigger.Branch == branch {
			return false
		}
		if trigger.Operator == types.PipelineTriggerOperatorInclude {
			matched, err := regexp.MatchString(trigger.Branch, branch)
			if err != nil {
				klog.Errorf("regex %s match branch %s error: %s", trigger.Branch, branch, err.Error())
				continue
			}
			if matched {
				return true
			}
		}
	}
	return true
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

func (r *ServicePipelineRun) InitialEnvs(pipeline *types.Pipeline, workspace *types.PipelineWorkspace, params map[string]interface{}) (map[string]interface{}, error) {
	envs := map[string]interface{}{}
	envs[types.PipelineEnvWorkspaceId] = workspace.ID
	envs[types.PipelineEnvWorkspaceName] = workspace.Name
	envs[types.PipelineEnvPipelineId] = pipeline.ID
	envs[types.PipelineEnvPipelineName] = pipeline.Name
	if workspace.Type == types.WorkspaceTypeCode {
		if err := r.InitialCodeEnvs(pipeline, workspace, params, envs); err != nil {
			return nil, err
		}
	} else if workspace.Type == types.WorkspaceTypeCustom {
		paramsBytes, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		var buildPipelineParams BuildForPipelineParams
		if err = json.Unmarshal(paramsBytes, &buildPipelineParams); err != nil {
			return nil, err
		}
		delPipelineEnvs := []string{
			types.PipelineEnvWorkspaceId,
			types.PipelineEnvWorkspaceName,
			types.PipelineEnvPipelineId,
			types.PipelineEnvPipelineName,
			types.PipelineEnvPipelineBuildNumber,
			types.PipelineEnvPipelineTriggerUser,
		}
		var pipelineBuildId []string
		for _, buildInfo := range buildPipelineParams.BuildIds {
			if !buildInfo.IsBuild {
				continue
			}
			find := false
			for _, trigger := range pipeline.Triggers {
				if buildInfo.PipelineId == trigger.Pipeline {
					find = true
					build, err := r.models.PipelineRunManager.Get(buildInfo.BuildId)
					if err != nil {
						return nil, fmt.Errorf("获取流水线构建源失败：%s", err.Error())
					}
					pipelineSrc, err := r.models.PipelineManager.Get(buildInfo.PipelineId)
					if err != nil {
						return nil, fmt.Errorf("获取流水线源失败：%s", err.Error())
					}
					if build.Status != types.PipelineStatusOK {
						return nil, fmt.Errorf("构建源流水线%s执行状态未完成", pipelineSrc.Name)
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
						envs = utils.MergeMap(envs, lastStage.Env)
					}
					pipelineBuildId = append(pipelineBuildId, fmt.Sprintf("%d", buildInfo.BuildId))
				}
			}
			if !find {
				return nil, fmt.Errorf("构建参数错误，流水线id=%d不在触发源", buildInfo.PipelineId)
			}
		}
		envs[types.PipelineEnvPipelineBuildId] = strings.Join(pipelineBuildId, ",")
	}

	return envs, nil
}

func (r *ServicePipelineRun) InitialCodeEnvs(pipeline *types.Pipeline, workspace *types.PipelineWorkspace, params, envs map[string]interface{}) error {
	if workspace.Code == nil {
		return fmt.Errorf("未获取到流水线空间代码信息")
	}
	envs["PIPELINE_CODE_URL"] = workspace.Code.CloneUrl
	paramBranch, ok := params["branch"]
	if ok {
		envs["PIPELINE_CODE_BRANCH"] = paramBranch
	} else {
		return fmt.Errorf("未获取到代码分支参数")
	}
	branch, ok := paramBranch.(string)
	if !ok {
		return fmt.Errorf("获取分支参数类型错误")
	}
	if !r.MatchTriggerBranch(pipeline.Triggers, branch) {
		return fmt.Errorf("代码分支未匹配到该流水线")
	}
	secret, err := r.models.SettingsSecretManager.Get(workspace.Code.SecretId)
	if err != nil {
		return fmt.Errorf("获取代码密钥失败：" + err.Error())
	}
	gitcli, err := utilgit.NewClient(workspace.Code.Type, "", &utilgit.Secret{
		Type:        secret.Type,
		User:        secret.User,
		Password:    secret.Password,
		PrivateKey:  secret.PrivateKey,
		AccessToken: secret.AccessToken,
	})
	commit, err := gitcli.GetBranchLatestCommit(context.Background(), workspace.Code.CloneUrl, branch)
	if err != nil {
		return fmt.Errorf("获取远程分支%s失败：%s", branch, err.Error())
	}
	envs["PIPELINE_CODE_COMMIT_ID"] = commit.CommitId
	envs["PIPELINE_CODE_COMMIT_AUTHOR"] = commit.Author
	envs["PIPELINE_CODE_COMMIT_MESSAGE"] = commit.Message
	envs["PIPELINE_CODE_COMMIT_TIME"] = commit.CommitTime
	return nil
}

func (r *ServicePipelineRun) Build(buildSer *serializers.PipelineBuildSerializer, user *types.User) *utils.Response {
	pipelineObj, err := r.models.PipelineManager.Get(buildSer.PipelineId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	workspace, err := r.models.PipelineWorkspaceManager.Get(pipelineObj.WorkspaceId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	stages, err := r.models.PipelineManager.Stages(buildSer.PipelineId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if len(stages) == 0 {
		return &utils.Response{Code: code.DataNotExists, Msg: "当前流水线未配置阶段"}
	}
	envs, err := r.InitialEnvs(pipelineObj, workspace, buildSer.Params)
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
	pipelineRun := &types.PipelineRun{
		PipelineId: buildSer.PipelineId,
		Status:     types.PipelineStatusWait,
		Operator:   user.Name,
		Params:     buildSer.Params,
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
	if len(manualSer.CustomParams) > 0 {
		stageRun.CustomParams = manualSer.CustomParams
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
		klog.Infof("current stage run id=%v status is %v, not error", stageRun.ID, stageRun.Status)
		return &utils.Response{Code: code.RequestError, Msg: "current stage run id=%v status is %v, not error"}
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
