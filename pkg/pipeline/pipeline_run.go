package pipeline

import (
	"encoding/json"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	sshgit "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/manager/pipeline"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/pipeline/plugins"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type codeCommit struct {
	CommitId   string
	Author     string
	Message    string
	CommitTime time.Time
}

type ServicePipelineRun struct {
	models         *model.Models
	builtInPlugins *plugins.Plugins
}

func NewPipelineRunService(models *model.Models, kr *kube_resource.KubeResources) *ServicePipelineRun {
	r := &ServicePipelineRun{
		models: models,
	}
	r.builtInPlugins = plugins.NewPlugins(models, kr, r.Callback)
	return r
}

func (r *ServicePipelineRun) ListPipelineRun(pipelineId uint, lastBuildNumber int) *utils.Response {
	pipelineRuns, err := r.models.ManagerPipelineRun.ListPipelineRun(pipelineId, lastBuildNumber)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	var retData []map[string]interface{}
	for _, pipelineRun := range pipelineRuns {
		stagesRun, err := r.models.ManagerPipelineRun.StagesRun(pipelineRun.ID)
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
	pipelineRun, err := r.models.ManagerPipelineRun.Get(pipelineRunId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	stagesRun, err := r.models.ManagerPipelineRun.StagesRun(pipelineRunId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	pipeline, err := r.models.ManagerPipeline.Get(pipelineRun.PipelineId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	workspace, err := r.models.PipelineWorkspaceManager.Get(pipeline.WorkspaceId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	data := map[string]interface{}{
		"pipeline":     pipeline,
		"pipeline_run": pipelineRun,
		"stages_run":   stagesRun,
		"workspace": map[string]interface{}{
			"id":       workspace.ID,
			"name":     workspace.Name,
			"type":     workspace.Type,
			"code_url": workspace.CodeUrl,
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

type BuildForPipelineParams struct {
	BuildIds []uint `json:"build_ids"`
}

func (r *ServicePipelineRun) InitialEnvs(pipeline *types.Pipeline, workspace *types.PipelineWorkspace, params map[string]interface{}) (map[string]interface{}, error) {
	envs := map[string]interface{}{}
	envs["PIPELINE_WORKSPACE_ID"] = workspace.ID
	envs["PIPELINE_WORKSPACE_NAME"] = workspace.Name
	envs["PIPELINE_PIPELINE_ID"] = pipeline.ID
	envs["PIPELINE_PIPELINE_NAME"] = pipeline.Name
	if workspace.Type == types.WorkspaceTypeCode {
		if err := r.InitialCodeEnvs(pipeline, workspace, params, envs); err != nil {
			return nil, err
		}
	} else if workspace.Type == types.WorkspaceTypePipeline {

	}

	return envs, nil
}

func (r *ServicePipelineRun) InitialCodeEnvs(pipeline *types.Pipeline, workspace *types.PipelineWorkspace, params, envs map[string]interface{}) error {
	envs["PIPELINE_CODE_URL"] = workspace.CodeUrl
	branch, ok := params["branch"]
	if ok {
		envs["PIPELINE_CODE_BRANCH"] = branch
	} else {
		return fmt.Errorf("未获取到代码分支参数")
	}
	if !r.MatchTriggerBranch(pipeline.Triggers, branch.(string)) {
		return fmt.Errorf("代码分支未匹配到该流水线")
	}
	var err error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, branchErr := r.getCodeBranchCommitId(workspace.CodeUrl, branch.(string), workspace.CodeSecretId)
		if err != nil {
			err = branchErr
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		b := branch.(string)
		commit, commitErr := r.getCodeBranchCommit(workspace.CodeUrl, b, workspace.CodeSecretId)
		if commitErr != nil {
			klog.Errorf("get code %s commit error: %v", workspace.CodeUrl, err)
			if err != nil {
				err = commitErr
			}
			return
		}
		envs["PIPELINE_CODE_COMMIT_ID"] = commit.CommitId
		envs["PIPELINE_CODE_COMMIT_AUTHOR"] = commit.Author
		envs["PIPELINE_CODE_COMMIT_MESSAGE"] = commit.Message
		envs["PIPELINE_CODE_COMMIT_TIME"] = commit.CommitTime
	}()
	wg.Wait()
	return err
}

func (r *ServicePipelineRun) Build(buildSer *serializers.PipelineBuildSerializer, user *types.User) *utils.Response {
	pipeline, err := r.models.ManagerPipeline.Get(buildSer.PipelineId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	workspace, err := r.models.PipelineWorkspaceManager.Get(pipeline.WorkspaceId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	stages, err := r.models.ManagerPipeline.Stages(buildSer.PipelineId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if len(stages) == 0 {
		return &utils.Response{Code: code.DataNotExists, Msg: "当前流水线未配置阶段"}
	}
	envs, err := r.InitialEnvs(pipeline, workspace, buildSer.Params)
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
		Env:        envs,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	pipelineRun, err = r.models.ManagerPipelineRun.CreatePipelineRun(pipelineRun, stagesRun)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	go r.Execute(pipelineRun, 0, types.StageTriggerModeAuto)
	return &utils.Response{Code: code.Success, Data: pipelineRun}
}

//func (r *ServicePipelineRun) ManualExecutePipeline(pipelineRun *types.PipelineRun, workspace *types.PipelineWorkspace) {
//	defer r.recoverExecute(pipelineRun)
//	if workspace.Type == types.WorkspaceTypeCode {
//		branch := pipelineRun.Env["PIPELINE_CODE_BRANCH"].(string)
//		commit, err := r.getCodeBranchCommit(workspace.CodeUrl, branch, workspace.CodeSecretId)
//		if err != nil {
//			klog.Errorf("get code %s commit error: %v", workspace.CodeUrl, err)
//			return
//		}
//		pipelineRun.Env["PIPELINE_CODE_COMMIT_ID"] = commit.CommitId
//		pipelineRun.Env["PIPELINE_CODE_COMMIT_AUTHOR"] = commit.Author
//		pipelineRun.Env["PIPELINE_CODE_COMMIT_MESSAGE"] = commit.Message
//		pipelineRun.Env["PIPELINE_CODE_COMMIT_TIME"] = commit.CommitTime
//		err = r.models.ManagerPipelineRun.UpdatePipelineRun(pipelineRun)
//		if err != nil {
//			klog.Errorf("update pipeline run %d envs error: %v", pipelineRun.ID, err)
//			return
//		}
//	}
//	r.Execute(pipelineRun, 0, types.StageTriggerModeAuto)
//}

func (r *ServicePipelineRun) recoverExecute(pipelineRun *types.PipelineRun) {
	if err := recover(); err != nil {
		klog.Error("error: ", err)
		var buf [4096]byte
		n := runtime.Stack(buf[:], false)
		klog.Errorf("==> %s\n", string(buf[:n]))
		pipelineRun.Status = types.PipelineStatusError
		err = r.models.ManagerPipelineRun.UpdatePipelineRun(pipelineRun)
		if err != nil {
			klog.Errorf("update pipeline run error: %v", err)
		}
	}
}

func (r *ServicePipelineRun) Execute(pipelineRun *types.PipelineRun, prevStageId uint, trigger string) {
	defer r.recoverExecute(pipelineRun)
	nextStage, err := r.models.ManagerPipelineRun.NextStageRun(pipelineRun.ID, prevStageId)
	if err != nil {
		klog.Errorf("get pipeline run id=%d next stage error, current stage id %d", pipelineRun.ID, prevStageId)
		pipelineRun.Status = types.PipelineStatusError
		err = r.models.ManagerPipelineRun.UpdatePipelineRun(pipelineRun)
		if err != nil {
			klog.Errorf("update pipeline run error: %s", err.Error())
		}
		return
	}
	if nextStage == nil {
		pipelineRun.Status = types.PipelineStatusOK
		err = r.models.ManagerPipelineRun.UpdatePipelineRun(pipelineRun)
		if err != nil {
			klog.Errorf("update pipeline run error: %s", err.Error())
		}
		return
	}
	if nextStage.TriggerMode == types.StageTriggerModeManual && trigger == types.StageTriggerModeAuto {
		klog.Infof("current stage id=%d trigger mode is manual, pausing...", nextStage.ID)
		if _, _, err = r.models.ManagerPipelineRun.UpdatePipelineStageRun(&pipeline.UpdateStageObj{
			StageRunId:     nextStage.ID,
			StageRunStatus: types.PipelineStatusPause,
		}); err != nil {
			klog.Errorf("update stage id=%d status to pause error: %v", nextStage.ID, err)
		}
		return
	}
	envs, _ := r.models.ManagerPipelineRun.GetEnvBeforeStageRun(nextStage)
	klog.Info(envs)
	nextStage.Env = envs
	nextStage.ExecTime = time.Now()
	nextStage.Status = types.PipelineStatusDoing
	err = r.models.ManagerPipelineRun.UpdateStageRun(nextStage)
	if err != nil {
		klog.Errorf("update stage id=%d exec time error: %v", nextStage.ID, err)
		return
	}
	runJobs := nextStage.Jobs
	for _, runJob := range runJobs {
		runJob.Status = types.PipelineStatusDoing
		_, nextStage, _ = r.models.ManagerPipelineRun.UpdatePipelineStageRun(&pipeline.UpdateStageObj{
			StageRunId:   nextStage.ID,
			StageRunJobs: types.PipelineRunJobs{runJob},
		})
		resp := r.ExecuteJob(nextStage, runJob)
		if !resp.IsSuccess() {
			runJob.Result = resp
			runJob.Status = types.PipelineStatusError
		}
	}
	for _, runJob := range runJobs {
		if runJob.Status == types.PipelineStatusError {
			_, nextStage, _ = r.models.ManagerPipelineRun.UpdatePipelineStageRun(&pipeline.UpdateStageObj{
				StageRunId:   nextStage.ID,
				StageRunJobs: types.PipelineRunJobs{runJob},
			})
		}
	}
}

func (r *ServicePipelineRun) getJobExecParam(envs map[string]interface{}, jobParams map[string]interface{}, pluginParam *types.PipelinePluginParamsSpec) interface{} {
	if pluginParam == nil {
		return nil
	}
	res := pluginParam.Default
	if pluginParam.From == types.PluginParamsFromPipelineEnv {
		return envs
	} else if pluginParam.From == types.PluginParamsFromEnv {
		if _, ok := envs[pluginParam.FromName]; ok {
			return envs[pluginParam.FromName]
		}
	} else if pluginParam.From == types.PluginParamsFromJob {
		if _, ok := jobParams[pluginParam.FromName]; ok {
			return jobParams[pluginParam.FromName]
		}
	} else if pluginParam.From == types.PluginParamsFromCodeSecret {
		res = nil
		workspaceId, err := strconv.ParseUint(fmt.Sprintf("%v", envs["PIPELINE_WORKSPACE_ID"]), 10, 64)
		if err != nil {
			return &utils.Response{Code: code.RequestError, Msg: "获取流水线空间失败：" + err.Error()}
		}
		workspace, _ := r.models.PipelineWorkspaceManager.Get(uint(workspaceId))
		if workspace != nil && workspace.CodeSecretId != 0 {
			secret, _ := r.models.SettingsSecretManager.Get(workspace.CodeSecretId)
			if secret != nil {
				return map[string]interface{}{
					"type":         secret.Type,
					"user":         secret.User,
					"password":     secret.Password,
					"private_key":  secret.PrivateKey,
					"access_token": secret.AccessToken,
				}
			}
		}
	} else if pluginParam.From == types.PluginParamsFromImageRegistry {
		res = nil
		var imageRegistry interface{}
		var ok bool
		if pluginParam.FromName == "" {
			imageRegistry, ok = envs["CODE_BUILD_REGISTRY_ID"]
		} else {
			imageRegistry, ok = jobParams[pluginParam.FromName]
		}
		var regId string
		if regId, ok = imageRegistry.(string); ok {
			imageRegistry = strings.Split(regId, ",")[0]
		}
		if imageRegistry == nil {
			klog.Errorf("not found image registry job params")
			return nil
		}
		registryId, err := strconv.ParseUint(fmt.Sprintf("%v", imageRegistry), 10, 64)
		if err != nil {
			klog.Errorf("parse registry to int error: %s", err.Error())
			return nil
		}
		registry, err := r.models.ImageRegistryManager.Get(uint(registryId))
		if err != nil {
			klog.Errorf("get image registry error: %s", err.Error())
			return nil
		}
		return map[string]interface{}{
			"registry": registry.Registry,
			"user":     registry.User,
			"password": registry.Password,
		}

	} else if pluginParam.From == types.PluginParamsFromPipelineResource {
		res = nil
		if resourceParam, ok := jobParams[pluginParam.FromName]; ok {
			resourceId, err := strconv.ParseUint(fmt.Sprintf("%v", resourceParam), 10, 64)
			if err == nil {
				resource, _ := r.models.PipelineResourceManager.Get(uint(resourceId))
				if resource != nil {
					res := map[string]interface{}{
						"type":  resource.Type,
						"value": resource.Value,
					}
					if resource.Secret != nil {
						res["secret"] = map[string]string{
							"type":         resource.Secret.Type,
							"user":         resource.Secret.User,
							"password":     resource.Secret.Password,
							"private_key":  resource.Secret.PrivateKey,
							"access_token": resource.Secret.AccessToken,
						}
					}
					return res
				}
			}
		}
	}
	return res
}

func (r *ServicePipelineRun) ExecuteJob(stageRun *types.PipelineRunStage, runJob *types.PipelineRunJob) (resp *utils.Response) {
	defer func() {
		if err := recover(); err != nil {
			klog.Error("error: ", err)
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			klog.Errorf("==> %s\n", string(buf[:n]))
			resp = &utils.Response{Code: code.DBError, Msg: fmt.Sprintf("执行插件错误:%s", string(buf[:n]))}
		}
	}()
	plugin, err := r.models.PipelinePluginManager.GetByKey(runJob.PluginKey)
	if err != nil {
		klog.Errorf("get plugin key=%s error: %v", runJob.PluginKey, err)
		return &utils.Response{Code: code.DBError, Msg: fmt.Sprintf("获取执行插件错误:%v", err)}
	}
	executeParams := map[string]interface{}{
		"job_id": runJob.ID,
	}
	for _, pluginParam := range plugin.Params.Params {
		if pluginParam.ParamName == "" {
			continue
		}
		executeParams[pluginParam.ParamName] = r.getJobExecParam(stageRun.Env, runJob.Params, pluginParam)
	}
	if plugin.Url == types.PipelinePluginBuiltinUrl {
		pluginParams := &plugins.PluginParams{
			JobId:     runJob.ID,
			PluginKey: plugin.Key,
			Params:    executeParams,
		}
		return r.builtInPlugins.Execute(pluginParams)
	} else {
		data, err := utils.HttpPost(plugin.Url, executeParams)
		if err != nil {
			klog.Errorf("request %s error: %v", plugin.Url, err)
			return &utils.Response{Code: code.RequestError, Msg: "请求插件接口失败:" + err.Error()}
		}
		var ret utils.Response
		err = json.Unmarshal(data, &ret)
		if err != nil {
			klog.Errorf("unmarshal data error: %v", err)
			return &utils.Response{Code: code.RequestError, Msg: "插件接口返回失败:" + err.Error()}
		}
		return &ret
	}
}

func (r *ServicePipelineRun) getJobRunResultEnvs(jobRun *types.PipelineRunJob) map[string]interface{} {
	if jobRun == nil {
		return nil
	}
	if jobRun.Status != types.PipelineStatusOK {
		return nil
	}
	if jobRun.Result.Data == nil {
		return nil
	}
	plugin, err := r.models.PipelinePluginManager.GetByKey(jobRun.PluginKey)
	if err != nil {
		klog.Errorf("get jobRun %s(%s) plugin error: %s", jobRun.ID, jobRun.Name, err.Error())
		return nil
	}
	if len(plugin.ResultEnv.EnvPath) == 0 {
		return nil
	}
	var envs = map[string]interface{}{}
	resData, ok := jobRun.Result.Data.(map[string]interface{})
	if ok {
		for _, envPath := range plugin.ResultEnv.EnvPath {
			if v, ok := resData[envPath.ResultName]; ok {
				envs[envPath.EnvName] = v
			}
		}
	} else {
		klog.Errorf("get job run id=%d result data error", jobRun.ID)
	}
	return envs
}

func (r *ServicePipelineRun) Callback(callbackSer serializers.PipelineCallbackSerializer) *utils.Response {
	callbackJobRun, err := r.models.ManagerPipelineRun.GetJobRun(callbackSer.JobId)
	if err != nil {
		klog.Errorf("get job run id=%v error: %v", callbackSer.JobId, err)
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	stageRun, err := r.models.ManagerPipelineRun.GetStageRun(callbackJobRun.StageRunId)
	if err != nil {
		klog.Errorf("get job run id=%v stage error: %v", callbackSer.JobId, err)
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if callbackSer.Result == nil {
		//klog.Infof("stage run id=%v job=%v callback return nil", stageRun.ID, callbackJobRun.JobId)
		resp := &utils.Response{Code: code.ParamsError, Msg: "stage job callback return nil"}
		callbackJobRun.Result = resp
		callbackSer.Result = resp
	} else {
		callbackJobRun.Result = callbackSer.Result
	}
	if callbackSer.Result.IsSuccess() {
		callbackJobRun.Status = types.PipelineStatusOK
	} else {
		callbackJobRun.Status = types.PipelineStatusError
	}
	envs := r.getJobRunResultEnvs(callbackJobRun)
	if envs != nil {
		callbackJobRun.Env = envs
	}
	//pipelineRun, stageRun, err := r.models.ManagerPipelineRun.UpdatePipelineStageRun(stageRun.ID, "", types.PipelineRunJobs{callbackJobRun})
	pipelineRun, stageRun, err := r.models.ManagerPipelineRun.UpdatePipelineStageRun(&pipeline.UpdateStageObj{
		StageRunId:   stageRun.ID,
		StageRunJobs: types.PipelineRunJobs{callbackJobRun},
	})
	if stageRun != nil && stageRun.Status == types.PipelineStatusOK {
		go r.Execute(pipelineRun, stageRun.ID, types.StageTriggerModeAuto)
	}
	return &utils.Response{Code: code.Success}
}

func (r *ServicePipelineRun) ManualExecuteStage(manualSer *serializers.PipelineStageManualSerializer) *utils.Response {
	stageRun, err := r.models.ManagerPipelineRun.GetStageRun(manualSer.StageRunId)
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
	if err = r.models.ManagerPipelineRun.UpdateStageJobRunParams(stageRun, stageRun.Jobs); err != nil {
		return &utils.Response{Code: code.DBError, Msg: "更新阶段任务参数失败:" + err.Error()}
	}
	pipelineRun, err := r.models.ManagerPipelineRun.Get(stageRun.PipelineRunId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	go r.Execute(pipelineRun, stageRun.PrevStageRunId, types.StageTriggerModeManual)
	return &utils.Response{Code: code.Success}
}

func (r *ServicePipelineRun) RetryStage(retrySer *serializers.PipelineStageRetrySerializer) *utils.Response {
	stageRun, err := r.models.ManagerPipelineRun.GetStageRun(retrySer.StageRunId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if stageRun.Status != types.PipelineStatusError {
		klog.Infof("current stage run id=%v status is %v, not error", stageRun.ID, stageRun.Status)
		return &utils.Response{Code: code.RequestError, Msg: "current stage run id=%v status is %v, not error"}
	}
	//pipelineRun, stageRun, err := r.models.ManagerPipelineRun.UpdatePipelineStageRun(stageRun.ID, types.PipelineStatusDoing, nil)
	pipelineRun, stageRun, err := r.models.ManagerPipelineRun.UpdatePipelineStageRun(&pipeline.UpdateStageObj{
		StageRunId: stageRun.ID,
		//StageRunStatus: types.PipelineStatusDoing,
	})
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	go r.Execute(pipelineRun, stageRun.PrevStageRunId, types.StageTriggerModeManual)
	return &utils.Response{Code: code.Success}
}

func (r *ServicePipelineRun) JobLog(jobRunId uint) *utils.Response {

	return &utils.Response{Code: code.Success}
}
