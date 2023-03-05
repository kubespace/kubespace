package pipelinerun

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/controller"
	"github.com/kubespace/kubespace/pkg/controller/pipelinerun/plugins"
	"github.com/kubespace/kubespace/pkg/core/lock"
	"github.com/kubespace/kubespace/pkg/informer"
	pipelinelistwatcher "github.com/kubespace/kubespace/pkg/informer/listwatcher/pipeline"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/manager/pipeline"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"k8s.io/klog/v2"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PipelineRunController struct {
	models           *model.Models
	pipelineInformer informer.Informer
	lock             lock.Lock
	jobPlugins       *plugins.Plugins
	dataDir          string
}

func NewPipelineRunController(config *controller.Config) *PipelineRunController {
	p := &PipelineRunController{
		models: config.Models,
		pipelineInformer: config.InformerFactory.PipelineRunInformer(&pipelinelistwatcher.PipelineRunWatchCondition{
			StatusIn: []string{types.PipelineStatusWait, types.PipelineStatusDoing},
			WithList: true,
		}),
		lock:       lock.NewMemLock(),
		jobPlugins: plugins.NewPlugins(config.Models, config.ServiceFactory.Cluster.KubeClient, config.InformerFactory),
		dataDir:    config.DataDir,
	}
	p.pipelineInformer.AddHandler(p)
	return p
}

func (p *PipelineRunController) Run(stopCh <-chan struct{}) {
	p.pipelineInformer.Run(stopCh)
}

func (p *PipelineRunController) Check(object interface{}) bool {
	pipelineRun, ok := object.(types.PipelineRun)
	if !ok {
		return false
	}
	latest, err := p.models.PipelineRunManager.Get(pipelineRun.ID)
	if err != nil {
		klog.Errorf("get latest pipeline run error: %s", err.Error())
		return false
	}
	if latest.Status != types.PipelineStatusDoing && latest.Status != types.PipelineStatusWait {
		return false
	}
	return true
}

func (p *PipelineRunController) Handle(object interface{}) (err error) {
	pipelineRun := object.(types.PipelineRun)
	if ok, _ := p.lock.Acquire(strconv.Itoa(int(pipelineRun.ID))); !ok {
		return nil
	}
	defer p.lock.Release(strconv.Itoa(int(pipelineRun.ID)))
	if latestPipelineRun, err := p.models.PipelineRunManager.Get(pipelineRun.ID); err != nil {
		return err
	} else {
		pipelineRun = *latestPipelineRun
	}
	defer utils.HandleCrash(func(r interface{}) {
		pipelineRun.Status = types.PipelineStatusError
		err := p.models.PipelineRunManager.UpdatePipelineRun(&pipelineRun)
		if err != nil {
			klog.Errorf("update pipeline run error: %v", err)
		}
	})
	var prevStageId uint = 0
	for {
		nextStage, err := p.models.PipelineRunManager.NextStageRun(pipelineRun.ID, prevStageId)
		if err != nil {
			klog.Errorf("get pipeline run id=%d next stage error, current stage id %d", pipelineRun.ID, prevStageId)
			pipelineRun.Status = types.PipelineStatusError
			return p.models.PipelineRunManager.UpdatePipelineRun(&pipelineRun)
		}
		if nextStage == nil {
			pipelineRun.Status = types.PipelineStatusOK
			return p.models.PipelineRunManager.UpdatePipelineRun(&pipelineRun)
		}
		if nextStage.Status == types.PipelineStatusOK {
			prevStageId = nextStage.ID
			continue
		}
		if err = p.executeStage(nextStage); err != nil {
			return err
		}
		nextStage, _ = p.models.PipelineRunManager.GetStageRun(nextStage.ID)
		if nextStage.Status != types.PipelineStatusOK {
			return nil
		}
		prevStageId = nextStage.ID
	}
}

func (p *PipelineRunController) executeStage(stageRun *types.PipelineRunStage) (err error) {
	if stageRun.TriggerMode == types.StageTriggerModeManual && stageRun.Status == types.PipelineStatusWait {
		klog.Infof("current stage id=%d trigger mode is manual, pausing...", stageRun.ID)
		if _, stageRun, err = p.models.PipelineRunManager.UpdatePipelineStageRun(&pipeline.UpdateStageObj{
			StageRunId:     stageRun.ID,
			StageRunStatus: types.PipelineStatusPause,
		}); err != nil {
			klog.Errorf("update stage id=%d status to pause error: %v", stageRun.ID, err)
		}
		return err
	}
	envs, _ := p.models.PipelineRunManager.GetEnvBeforeStageRun(stageRun)
	stageRun.Env = envs
	if stageRun.Status == types.PipelineStatusWait {
		stageRun.ExecTime = time.Now()
		stageRun.Status = types.PipelineStatusDoing
	}
	err = p.models.PipelineRunManager.UpdateStageRun(stageRun)
	if err != nil {
		klog.Errorf("update stage id=%d exec time error: %v", stageRun.ID, err)
		return
	}
	runJobs := stageRun.Jobs
	wg := sync.WaitGroup{}
	muSync := sync.Mutex{}
	for _, runJob := range runJobs {
		if runJob.Status == types.PipelineStatusOK {
			continue
		}
		runJob.Status = types.PipelineStatusDoing
		_, stageRun, _ = p.models.PipelineRunManager.UpdatePipelineStageRun(&pipeline.UpdateStageObj{
			StageRunId:   stageRun.ID,
			StageRunJobs: types.PipelineRunJobs{runJob},
		})
		wg.Add(1)
		go func(runJob *types.PipelineRunJob) {
			defer wg.Done()
			resp := p.executeJob(stageRun, runJob)
			runJob.Result = resp
			if !resp.IsSuccess() {
				runJob.Status = types.PipelineStatusError
			} else {
				runJob.Status = types.PipelineStatusOK
			}
			jobEnvs := p.getJobRunResultEnvs(runJob)
			if jobEnvs != nil {
				runJob.Env = jobEnvs
			}
			muSync.Lock()
			defer muSync.Unlock()
			_, stageRun, _ = p.models.PipelineRunManager.UpdatePipelineStageRun(&pipeline.UpdateStageObj{
				StageRunId:   stageRun.ID,
				StageRunJobs: types.PipelineRunJobs{runJob},
			})
		}(runJob)
	}
	wg.Wait()
	return
}

func (p *PipelineRunController) executeJob(stageRun *types.PipelineRunStage, runJob *types.PipelineRunJob) (resp *utils.Response) {
	plugin, err := p.models.PipelinePluginManager.GetByKey(runJob.PluginKey)
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
		executeParams[pluginParam.ParamName], err = p.getJobExecParam(stageRun.Env, runJob.Params, pluginParam)
		if err != nil {
			return &utils.Response{Code: code.ParamsError, Msg: fmt.Sprintf("获取执行参数异常：%s", err.Error())}
		}
	}
	pluginParams := &plugins.PluginParams{
		JobId:     runJob.ID,
		PluginKey: plugin.Key,
		Params:    executeParams,
		DataDir:   p.dataDir,
	}
	return p.jobPlugins.Execute(pluginParams)
}

func (p *PipelineRunController) getJobExecParam(
	envs map[string]interface{},
	jobParams map[string]interface{},
	pluginParam *types.PipelinePluginParamsSpec) (interface{}, error) {
	if pluginParam == nil {
		return nil, nil
	}
	res := pluginParam.Default
	switch pluginParam.From {
	case types.PluginParamsFromPipelineEnv:
		res = envs
	case types.PluginParamsFromEnv:
		if pp, ok := envs[pluginParam.FromName]; ok {
			res = pp
		}
	case types.PluginParamsFromJob:
		if jp, ok := jobParams[pluginParam.FromName]; ok {
			res = jp
		}
	case types.PluginParamsFromCodeSecret:
		res = nil
		workspaceId, err := strconv.Atoi(fmt.Sprintf("%v", envs["PIPELINE_WORKSPACE_ID"]))
		if err != nil {
			return nil, fmt.Errorf("获取流水线空间参数错误：" + err.Error())
		}
		workspace, err := p.models.PipelineWorkspaceManager.Get(uint(workspaceId))
		if err != nil {
			return nil, fmt.Errorf("获取流水线空间「id=%d」失败：%s", workspaceId, err.Error())
		}
		if workspace.Code != nil && workspace.Code.SecretId == 0 {
			return nil, nil
		}
		secret, _ := p.models.SettingsSecretManager.Get(workspace.Code.SecretId)
		if secret != nil {
			res = map[string]interface{}{
				"type":         secret.Type,
				"user":         secret.User,
				"password":     secret.Password,
				"private_key":  secret.PrivateKey,
				"access_token": secret.AccessToken,
			}
		}
	case types.PluginParamsFromImageRegistry:
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
			return nil, nil
		}
		registryId, err := strconv.ParseUint(fmt.Sprintf("%v", imageRegistry), 10, 64)
		if err != nil {
			klog.Errorf("parse registry to int error: %s", err.Error())
			return nil, nil
		}
		registry, err := p.models.ImageRegistryManager.Get(uint(registryId))
		if err != nil {
			klog.Errorf("get image registry error: %s", err.Error())
			return nil, nil
		}
		res = map[string]interface{}{
			"registry": registry.Registry,
			"user":     registry.User,
			"password": registry.Password,
		}
	case types.PluginParamsFromPipelineResource:
		resourceParam, ok := jobParams[pluginParam.FromName]
		if !ok {
			return nil, nil
		}
		resourceId, err := strconv.ParseUint(fmt.Sprintf("%v", resourceParam), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("获取流水线资源参数错误：%s", err.Error())
		}
		resource, err := p.models.PipelineResourceManager.Get(uint(resourceId))
		if err != nil {
			return nil, fmt.Errorf("获取流水线资源id=%d失败：%s", resourceId, err.Error())
		}
		resMap := map[string]interface{}{
			"type":  resource.Type,
			"value": resource.Value,
		}
		if resource.Secret != nil {
			resMap["secret"] = map[string]string{
				"type":         resource.Secret.Type,
				"user":         resource.Secret.User,
				"password":     resource.Secret.Password,
				"private_key":  resource.Secret.PrivateKey,
				"access_token": resource.Secret.AccessToken,
			}
		}
		res = resMap
	}
	return res, nil
}

func (p *PipelineRunController) getJobRunResultEnvs(jobRun *types.PipelineRunJob) map[string]interface{} {
	if jobRun == nil {
		return nil
	}
	if jobRun.Status != types.PipelineStatusOK {
		return nil
	}
	if jobRun.Result.Data == nil {
		return nil
	}
	plugin, err := p.models.PipelinePluginManager.GetByKey(jobRun.PluginKey)
	if err != nil {
		klog.Errorf("get jobRun %s(%s) plugin error: %s", jobRun.ID, jobRun.Name, err.Error())
		return nil
	}
	var envs = map[string]interface{}{}
	if plugin.Key == types.BuiltinPluginExecuteShell {
		// 执行脚本，更新当前阶段的环境变量
		stageRun, err := p.models.PipelineRunManager.GetStageRun(jobRun.StageRunId)
		if err != nil {
			klog.Errorf("get callback job stage run error: %s", err.Error())
			return nil
		}
		var stageEnvKeys []string
		for envKey := range stageRun.Env {
			stageEnvKeys = append(stageEnvKeys, envKey)
		}
		resData, ok := jobRun.Result.Data.(map[string]interface{})
		if ok {
			for k, v := range resData {
				if utils.Contains(stageEnvKeys, k) {
					envs[k] = v
				}
			}
		} else {
			klog.Errorf("get job run id=%d result data error, data=%+v", jobRun.ID, jobRun.Result)
		}
	} else {
		if len(plugin.ResultEnv.EnvPath) == 0 {
			return nil
		}
		resMap := make(map[string]interface{})
		if err = utils.ConvertTypeByJson(jobRun.Result.Data, &resMap); err != nil {
			klog.Errorf("get job run id=%d result data error, data=%+v", jobRun.ID, jobRun.Result)
			return envs
		}
		for _, envPath := range plugin.ResultEnv.EnvPath {
			if v, ok := resMap[envPath.ResultName]; ok {
				envs[envPath.EnvName] = v
			}
		}
	}
	return envs
}
