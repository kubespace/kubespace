package pipelinerun

import (
	"fmt"
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

func (p *PipelineRunController) buildLockKey(id uint) string {
	return fmt.Sprintf("pipeline_run_controller:build:run:%d", id)
}

// Check 检查流水线构建状态以及是否正在执行
func (p *PipelineRunController) buildCheck(object interface{}) bool {
	pipelineRun, ok := object.(types.PipelineRun)
	if !ok {
		return false
	}
	if pipelineRun.Status != types.PipelineStatusDoing && pipelineRun.Status != types.PipelineStatusWait {
		return false
	}
	if locked, _ := p.lock.Locked(p.buildLockKey(pipelineRun.ID)); locked {
		// 该流水线构建已存在锁，正在被执行
		return false
	}
	return true
}

func (p *PipelineRunController) build(object interface{}) (err error) {
	pipelineRun := object.(types.PipelineRun)
	// 对流水线构建执行加锁，保证只有一个goroutinue执行
	if ok, _ := p.lock.Acquire(p.buildLockKey(pipelineRun.ID)); !ok {
		return nil
	}
	// 执行完成释放锁
	defer p.lock.Release(p.buildLockKey(pipelineRun.ID))
	if latestPipelineRun, err := p.models.PipelineRunManager.Get(pipelineRun.ID); err != nil {
		return err
	} else {
		pipelineRun = *latestPipelineRun
	}

	if pipelineRun.Status != types.PipelineStatusDoing && pipelineRun.Status != types.PipelineStatusWait {
		return fmt.Errorf("pipeline run id=%d status=%s, do not run", pipelineRun.ID, pipelineRun.Status)
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
			// 下一个阶段为空，表示流水线构建已执行完成，状态置为ok
			pipelineRun.Status = types.PipelineStatusOK
			return p.models.PipelineRunManager.UpdatePipelineRun(&pipelineRun)
		}
		if nextStage.Status == types.PipelineStatusOK {
			// 阶段状态ok，执行下一个阶段
			prevStageId = nextStage.ID
			continue
		}
		// 执行当前阶段所有任务
		if err = p.executeStage(nextStage); err != nil {
			return err
		}
		nextStage, _ = p.models.PipelineRunManager.GetStageRun(nextStage.ID)
		if nextStage.Status != types.PipelineStatusOK {
			// 当前阶段执行不成功，退出
			return nil
		}
		prevStageId = nextStage.ID
	}
}

func (p *PipelineRunController) executeStage(stageRun *types.PipelineRunStage) (err error) {
	if stageRun.TriggerMode == types.StageTriggerModeManual && stageRun.Status == types.PipelineStatusWait {
		// 阶段触发状态为手动，且执行状态为wait，修改流水线构建状态为pause并退出，等待用户在页面手动点击执行继续
		// 用户手动点击执行后，会将阶段状态修改为doing
		klog.Infof("current stage id=%d trigger mode is manual, pausing...", stageRun.ID)
		if _, stageRun, err = p.models.PipelineRunManager.UpdatePipelineStageRun(&pipeline.UpdateStageObj{
			StageRunId:     stageRun.ID,
			StageRunStatus: types.PipelineStatusPause,
		}); err != nil {
			klog.Errorf("update stage id=%d status to pause error: %v", stageRun.ID, err)
		}
		return err
	}
	// 获取当前阶段之前的所有参数，并赋值给当前阶段
	envs, _ := p.models.PipelineRunManager.GetEnvBeforeStageRun(stageRun)
	stageRun.Env = envs
	if stageRun.Status == types.PipelineStatusWait {
		stageRun.ExecTime = time.Now()
		stageRun.Status = types.PipelineStatusDoing
	}
	klog.Infof("current stage id=%d envs=%v", stageRun.ID, envs)
	err = p.models.PipelineRunManager.UpdateStageRun(stageRun)
	if err != nil {
		klog.Errorf("update stage id=%d exec time error: %v", stageRun.ID, err)
		return
	}
	runJobs := stageRun.Jobs
	wg := sync.WaitGroup{}
	muSync := sync.Mutex{}
	// 并发执行阶段所有任务
	for _, runJob := range runJobs {
		if runJob.Status == types.PipelineStatusOK {
			// 任务状态ok不执行
			continue
		}
		// 修改任务状态为doing，并更新数据库
		runJob.Status = types.PipelineStatusDoing
		runJob.Result = nil
		_, stageRun, _ = p.models.PipelineRunManager.UpdatePipelineStageRun(&pipeline.UpdateStageObj{
			StageRunId:   stageRun.ID,
			StageRunJobs: types.PipelineRunJobs{runJob},
		})
		// 清空日志
		if err = p.models.PipelineJobLogManager.UpdateLog(runJob.ID, ""); err != nil {
			klog.Errorf("clear jobrun id=%d log error: %s", runJob.ID, err.Error())
		}
		wg.Add(1)
		go func(runJob *types.PipelineRunJob) {
			defer wg.Done()
			resp := p.executeJob(stageRun, runJob)
			runJob, err = p.models.PipelineRunManager.GetJobRun(runJob.ID)
			if err != nil {
				return
			}
			if runJob.Status == types.PipelineStatusCanceled {
				// 当前任务执行状态为已取消，退出
				return
			}
			if runJob.Status == types.PipelineStatusCancel {
				// 当前任务执行状态取消中，修改为已取消
				runJob.Status = types.PipelineStatusCanceled
				runJob.Result = &utils.Response{Code: code.JobCanceled}
			} else {
				runJob.Result = resp
				if !resp.IsSuccess() {
					runJob.Status = types.PipelineStatusError
				} else {
					runJob.Status = types.PipelineStatusOK
				}
				// 任务执行完成后，根据任务插件配置，获取当前任务的参数，传递给下一个阶段
				jobEnvs := p.getJobRunResultEnvs(runJob)
				if jobEnvs != nil {
					runJob.Env = jobEnvs
				}
			}
			// 更新任务需要加锁，防止多个任务同时更新互相覆盖
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
	klog.Infof("stage run id=%d envs=%v", stageRun.ID, stageRun.Env)
	// 根据任务插件配置，从阶段中获取对应参数值
	for _, pluginParam := range plugin.Params.Params {
		if pluginParam.ParamName == "" {
			continue
		}
		executeParams[pluginParam.ParamName], err = p.getJobExecParam(stageRun.Env, runJob.Params, pluginParam)
		if err != nil {
			return &utils.Response{Code: code.ParamsError, Msg: fmt.Sprintf("获取执行参数异常：%s", err.Error())}
		}
	}
	return p.jobRun.Execute(runJob.ID, plugin.Key, executeParams)
}

// 计算当前任务执行所需的参数
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
			res = secret.GetSecret()
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
		res = registry.GetImageRegistry()
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
			resMap["secret"] = resource.Secret.GetSecret()
		}
		res = resMap
	}
	return res, nil
}

// getJobRunResultEnvs 获取任务执行完成后的参数，传递给下一个阶段
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
