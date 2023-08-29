package pipeline

import (
	"errors"
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	corerrors "github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher"
	listwatcherconfig "github.com/kubespace/kubespace/pkg/informer/listwatcher/config"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/pipeline"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
	"strings"
	"time"
)

type PipelineRunManager struct {
	DB                        *gorm.DB
	PluginManager             *PipelinePluginManager
	pipelineRunListWatcher    listwatcher.Interface
	pipelineRunJobListWatcher listwatcher.Interface
}

func NewPipelineRunManager(db *gorm.DB, pluginManager *PipelinePluginManager, listwatcherConfig *listwatcherconfig.ListWatcherConfig) *PipelineRunManager {
	return &PipelineRunManager{
		DB:                        db,
		PluginManager:             pluginManager,
		pipelineRunListWatcher:    pipeline.NewPipelineRunListWatcher(listwatcherConfig, nil),
		pipelineRunJobListWatcher: pipeline.NewPipelineRunJobListWatcher(listwatcherConfig, nil),
	}
}

type ListPipelineRunCondition struct {
	PipelineId      uint
	LastBuildNumber int
	Status          string
	Limit           int
}

func (p *PipelineRunManager) ListPipelineRun(cond ListPipelineRunCondition) ([]*types.PipelineRun, error) {
	var pipelineRuns []*types.PipelineRun
	tx := p.DB.Order("id desc").Limit(cond.Limit).Where("pipeline_id = ?", cond.PipelineId)
	if cond.LastBuildNumber != 0 {
		tx = tx.Where("build_number < ?", cond.LastBuildNumber)
	}
	if cond.Status != "" {
		tx = tx.Where("status = ?", cond.Status)
	}
	if err := tx.Find(&pipelineRuns).Error; err != nil {
		return nil, err
	}
	return pipelineRuns, nil
}

func (p *PipelineRunManager) GetLastPipelineRun(pipelineId uint) (*types.PipelineRun, error) {
	var lastPipelineRun types.PipelineRun
	if err := p.DB.Last(&lastPipelineRun, "pipeline_id = ?", pipelineId).Error; err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return nil, nil
		}
		return nil, err
	}
	return &lastPipelineRun, nil
}

func (p *PipelineRunManager) GetLastBuildNumber(pipelineId uint) (uint, error) {
	var lastPipelineRun types.PipelineRun
	if err := p.DB.Last(&lastPipelineRun, "pipeline_id = ?", pipelineId).Error; err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return 1, nil
		}
		return 0, err
	}
	return lastPipelineRun.BuildNumber + 1, nil
}

func (p *PipelineRunManager) CreatePipelineRun(pipelineRun *types.PipelineRun, stagesRun []*types.PipelineRunStage) (*types.PipelineRun, error) {
	err := p.DB.Transaction(func(tx *gorm.DB) error {
		var lastPipelineRun types.PipelineRun
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Last(&lastPipelineRun, "pipeline_id = ?", pipelineRun.PipelineId).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
		buildNum := lastPipelineRun.BuildNumber + 1
		pipelineRun.BuildNumber = buildNum
		pipelineRun.Env[types.PipelineEnvPipelineBuildNumber] = buildNum
		pipelineRun.Env[types.PipelineEnvPipelineTriggerUser] = pipelineRun.Operator
		if err := tx.Create(pipelineRun).Error; err != nil {
			return err
		}
		var prevStageRunId uint = 0
		for _, stageRun := range stagesRun {
			stageRun.PipelineRunId = pipelineRun.ID
			stageRun.PrevStageRunId = prevStageRunId
			if err := tx.Create(stageRun).Error; err != nil {
				return err
			}
			for _, jobRun := range stageRun.Jobs {
				jobRun.StageRunId = stageRun.ID
				if err := tx.Create(jobRun).Error; err != nil {
					return err
				}
			}
			prevStageRunId = stageRun.ID
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if notifyErr := p.pipelineRunListWatcher.Notify(pipelineRun); notifyErr != nil {
		klog.Errorf("notify pipelineRun object error: %s", notifyErr.Error())
	}
	return pipelineRun, nil
}

func (p *PipelineRunManager) GetStageRunJobs(stageRunId uint) (types.PipelineRunJobs, error) {
	var runJobs []types.PipelineRunJob
	if err := p.DB.Where("stage_run_id = ?", stageRunId).Find(&runJobs).Error; err != nil {
		return nil, err
	}
	var stageRunJobs types.PipelineRunJobs
	for i := range runJobs {
		stageRunJobs = append(stageRunJobs, &runJobs[i])
	}
	return stageRunJobs, nil
}

func (p *PipelineRunManager) NextStageRun(pipelineRunId uint, stageId uint) (*types.PipelineRunStage, error) {
	var err error
	var stageRun types.PipelineRunStage
	if err = p.DB.Last(&stageRun, "prev_stage_run_id = ? and pipeline_run_id = ?", stageId, pipelineRunId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if stageRun.Jobs, err = p.GetStageRunJobs(stageRun.ID); err != nil {
		return nil, err
	}
	return &stageRun, nil
}

func (p *PipelineRunManager) Get(pipelineRunId uint) (*types.PipelineRun, error) {
	var pipelineRun types.PipelineRun
	if err := p.DB.First(&pipelineRun, pipelineRunId).Error; err != nil {
		return nil, err
	}
	return &pipelineRun, nil
}

func (p *PipelineRunManager) GetJobRun(jobRunId uint) (*types.PipelineRunJob, error) {
	var jobRun types.PipelineRunJob
	if err := p.DB.First(&jobRun, jobRunId).Error; err != nil {
		return nil, err
	}
	return &jobRun, nil
}

type JobRunListCondition struct {
	WithSpacelet *bool    `json:"with_spacelet"`
	StatusIn     []string `json:"status_in"`
}

func (p *PipelineRunManager) ListJobRun(cond *JobRunListCondition) ([]*types.PipelineRunJob, error) {
	var jobRun []*types.PipelineRunJob
	tx := p.DB
	if cond.WithSpacelet != nil {
		if *cond.WithSpacelet {
			tx = tx.Where("spacelet_id != 0 and spacelet_id is not null")
		} else {
			tx = tx.Where("spacelet_id = 0 or spacelet_id is null")
		}
	}
	if len(cond.StatusIn) > 0 {
		tx = tx.Where("status in ?", cond.StatusIn)
	}
	if err := tx.Find(&jobRun).Error; err != nil {
		return nil, err
	}
	return jobRun, nil
}

func (p *PipelineRunManager) NotifyJobRun(jobRun *types.PipelineRunJob) error {
	return p.pipelineRunJobListWatcher.Notify(jobRun)
}

func (p *PipelineRunManager) UpdateJobRun(id uint, jobRun *types.PipelineRunJob) error {
	return p.DB.Model(types.PipelineRunJob{}).Where("id=?", id).Updates(jobRun).Error
}

func (p *PipelineRunManager) GetStageRun(stageId uint) (*types.PipelineRunStage, error) {
	var err error
	var stageRun types.PipelineRunStage
	if err = p.DB.First(&stageRun, stageId).Error; err != nil {
		return nil, err
	}
	if stageRun.Jobs, err = p.GetStageRunJobs(stageId); err != nil {
		return nil, err
	}
	return &stageRun, nil
}

func (p *PipelineRunManager) StagesRun(pipelineRunId uint) ([]*types.PipelineRunStage, error) {
	var stagesRun []types.PipelineRunStage
	if err := p.DB.Where("pipeline_run_id = ?", pipelineRunId).Find(&stagesRun).Error; err != nil {
		return nil, err
	}
	for i, stageRun := range stagesRun {
		stageRunJobs, err := p.GetStageRunJobs(stageRun.ID)
		if err != nil {
			return nil, err
		}
		stagesRun[i].Jobs = stageRunJobs
	}
	var seqStages []*types.PipelineRunStage
	prevStageId := uint(0)
	for {
		hasNext := false
		for i, s := range stagesRun {
			if s.PrevStageRunId == prevStageId {
				seqStages = append(seqStages, &stagesRun[i])
				prevStageId = s.ID
				hasNext = true
				break
			}
		}
		if !hasNext {
			break
		}
	}

	return seqStages, nil
}

func (p *PipelineRunManager) UpdateStageRun(stageRun *types.PipelineRunStage) error {
	return p.DB.Save(stageRun).Error
}

func (p *PipelineRunManager) UpdateStageJobRunParams(stageRun *types.PipelineRunStage, jobRuns []*types.PipelineRunJob) error {
	return p.DB.Transaction(func(tx *gorm.DB) error {
		if err := p.DB.Select("custom_params").Save(stageRun).Error; err != nil {
			return err
		}
		for i := range jobRuns {
			if err := p.DB.Select("params").Save(jobRuns[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetStageRunStatus 根据stage的所有任务的状态返回该stage的状态
//  1. 如果有doing的job，stage状态为doing；
//  2. 如果所有job的状态为error/ok/wait，则
//     a. job中有error的，则stage为error；
//     b. 所有job都为ok，则stage为ok；
//     c. job中有ok，有wait，则stage为doing；
func (p *PipelineRunManager) GetStageRunStatus(stageRun *types.PipelineRunStage) string {
	if stageRun.Status == types.PipelineStatusCancel || stageRun.Status == types.PipelineStatusCanceled {
		// 如果当前阶段状态为取消状态，不以任务状态为准
		return stageRun.Status
	}
	status := ""
	for _, jobRun := range stageRun.Jobs {
		if jobRun.Status == types.PipelineStatusDoing {
			return types.PipelineStatusDoing
		}
		if jobRun.Status == types.PipelineStatusError {
			status = types.PipelineStatusError
			continue
		}
		if jobRun.Status == types.PipelineStatusOK && status != types.PipelineStatusError {
			status = types.PipelineStatusOK
			continue
		}
		if jobRun.Status == types.PipelineStatusWait && status == types.PipelineStatusOK {
			status = types.PipelineStatusDoing
		}
	}
	if status == "" {
		status = stageRun.Status
	}
	return status
}

func (p *PipelineRunManager) GetStageRunEnv(stageRun *types.PipelineRunStage) types.Map {
	envs := make(map[string]interface{})
	for _, jobRun := range stageRun.Jobs {
		// 当前阶段所有Job合并env
		envs = utils.MergeMap(envs, jobRun.Env)
	}
	// 合并替换之前阶段的env
	stageEnvs := utils.MergeReplaceMap(stageRun.Env, envs)
	return stageEnvs
}

type UpdateStageObj struct {
	StageRunId     uint
	StageRunStatus string
	StageExecTime  *time.Time
	StageRunJobs   types.PipelineRunJobs
	CustomParams   types.Map
}

func (p *PipelineRunManager) UpdatePipelineStageRun(updateStageObj *UpdateStageObj) (*types.PipelineRun, *types.PipelineRunStage, error) {
	if updateStageObj == nil {
		klog.Info("parameter stageObj is empty")
		return nil, nil, corerrors.New(code.ParamsError, "parameter is empty")
	}
	if updateStageObj.StageRunId == 0 {
		klog.Info("parameter stageRunId is empty")
		return nil, nil, corerrors.New(code.ParamsError, "parameter stageRunId is empty")
	}
	var stageRun types.PipelineRunStage
	var pipelineRun types.PipelineRun
	err := p.DB.Transaction(func(tx *gorm.DB) error {
		var err error

		// select for update数据库锁，防止并发修改
		if err = tx.Set("gorm:query_option", "FOR UPDATE").First(&stageRun, updateStageObj.StageRunId).Error; err != nil {
			return err
		}
		if updateStageObj.StageRunStatus == types.PipelineStatusCancel && stageRun.Status != types.PipelineStatusDoing {
			return corerrors.New(code.StatusError, "stage status is not doing, can not cancel")
		}
		if updateStageObj.StageRunStatus == types.PipelineStatusCanceled && stageRun.Status != types.PipelineStatusCancel {
			// 阶段状态修改为canceled已取消，当前状态必须为cancel取消中
			return corerrors.New(code.StatusError, "stage status is not cancel, can not set canceled")
		}

		if updateStageObj.StageRunJobs != nil {
			for _, updateJobRun := range updateStageObj.StageRunJobs {
				if updateJobRun.Status == types.PipelineStatusCancel {
					jobRun, err := p.GetJobRun(updateJobRun.ID)
					if err != nil {
						return err
					}
					// 只能取消未执行完成的任务
					if jobRun.Status == types.PipelineStatusOK || jobRun.Status == types.PipelineStatusError {
						continue
					}
				}
				if err = tx.Save(updateJobRun).Error; err != nil {
					return err
				}
			}
		}
		if updateStageObj.StageExecTime != nil {
			// 更新阶段开始执行时间
			stageRun.ExecTime = *updateStageObj.StageExecTime
		}
		if updateStageObj.CustomParams != nil {
			stageRun.CustomParams = updateStageObj.CustomParams
		}

		if updateStageObj.StageRunStatus != "" {
			// 如果传了阶段状态，直接更新该状态，否则根据阶段下的所有任务状态计算出阶段状态
			stageRun.Status = updateStageObj.StageRunStatus
		} else if updateStageObj.StageRunJobs != nil &&
			stageRun.Status != types.PipelineStatusCancel && stageRun.Status != types.PipelineStatusCanceled {
			// 当前阶段状态如果为取消，则不更新阶段状态以及环境参数
			var runJobs []types.PipelineRunJob
			if err = tx.Where("stage_run_id = ?", updateStageObj.StageRunId).Find(&runJobs).Error; err != nil {
				return err
			}
			stageRun.Jobs = types.PipelineRunJobs{}
			for i := range runJobs {
				stageRun.Jobs = append(stageRun.Jobs, &runJobs[i])
			}
			stageRun.Status = p.GetStageRunStatus(&stageRun)
			// 获取所有任务合并后的参数，传递给下一个阶段
			stageEnvs := p.GetStageRunEnv(&stageRun)
			if stageEnvs != nil {
				stageRun.Env = stageEnvs
			}
		}
		if err = tx.First(&pipelineRun, stageRun.PipelineRunId).Error; err != nil {
			return err
		}
		// 流水线构建状态根据阶段状态不同而不同
		now := time.Now()
		if updateStageObj.StageRunStatus == types.PipelineStatusCancel {
			// 如果第一次更新阶段状态为取消，设置完成时间为当前时间，后续再更新，不会修改完成时间
			stageRun.FinishTime = &now
		}
		if stageRun.Status == types.PipelineStatusError {
			stageRun.FinishTime = &now
			pipelineRun.Status = types.PipelineStatusError
		} else if stageRun.Status == types.PipelineStatusDoing || stageRun.Status == types.PipelineStatusWait {
			pipelineRun.Status = types.PipelineStatusDoing
		} else if stageRun.Status == types.PipelineStatusOK {
			stageRun.FinishTime = &now
			pipelineRun.Status = types.PipelineStatusDoing
		} else if stageRun.Status == types.PipelineStatusPause {
			pipelineRun.Status = types.PipelineStatusPause
		} else if stageRun.Status == types.PipelineStatusCancel {
			pipelineRun.Status = types.PipelineStatusCancel
		} else if stageRun.Status == types.PipelineStatusCanceled {
			pipelineRun.Status = types.PipelineStatusCanceled
		}
		stageRun.UpdateTime = now
		pipelineRun.UpdateTime = now
		if err = tx.Save(&stageRun).Error; err != nil {
			return err
		}
		if err = tx.Save(&pipelineRun).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	// 发送informer watch通知
	if notifyErr := p.pipelineRunListWatcher.Notify(pipelineRun); notifyErr != nil {
		klog.Warningf("notify pipeline run error: %s", notifyErr.Error())
	}
	return &pipelineRun, &stageRun, nil
}

// ReexecStage 重新执行已执行完成的阶段
func (p *PipelineRunManager) ReexecStage(
	pipelineRunId,
	stageRunId uint,
	customParams map[string]interface{},
	jobParams map[uint]map[string]interface{}) (*types.PipelineRun, *types.PipelineRunStage, error) {
	var pipelineRun *types.PipelineRun
	var stageRun *types.PipelineRunStage
	timeNow := time.Now()
	err := p.DB.Transaction(func(tx *gorm.DB) error {
		var err error

		// select for update数据库锁，防止并发修改
		if err = tx.Set("gorm:query_option", "FOR UPDATE").First(&pipelineRun, pipelineRunId).Error; err != nil {
			return err
		}
		stageRun, err = p.GetStageRun(stageRunId)
		if err != nil {
			return err
		}
		if !utils.Contains([]string{types.PipelineStatusOK, types.PipelineStatusError, types.PipelineStatusPause},
			pipelineRun.Status) {
			return corerrors.New(code.StatusError, fmt.Sprintf("current pipeline status is %s, cannot reexecute", pipelineRun.Status))
		}
		if stageRun.Status != types.PipelineStatusOK && stageRun.Status != types.PipelineStatusError {
			return corerrors.New(code.StatusError, fmt.Sprintf("current stage status is %s, cannot reexecute", pipelineRun.Status))
		}
		// 更新该阶段状态为doing，以及开始执行时间
		stageRun.Status = types.PipelineStatusDoing
		stageRun.UpdateTime = timeNow
		stageRun.ExecTime = timeNow
		stageRun.CustomParams = customParams
		if err = tx.Select("status", "exec_time", "custom_params", "update_time").Save(&stageRun).Error; err != nil {
			return err
		}

		for i, job := range stageRun.Jobs {
			stageRun.Jobs[i].Status = types.PipelineStatusWait
			stageRun.Jobs[i].UpdateTime = timeNow
			if params, ok := jobParams[job.ID]; ok {
				stageRun.Jobs[i].Params = params
				if err = tx.Select("params", "update_time").Save(&stageRun.Jobs[i]).Error; err != nil {
					return err
				}
			}
		}

		// 更新该阶段后续所有已执行阶段的状态为wait
		pipelineRunStages, err := p.StagesRun(pipelineRunId)
		if err != nil {
			return err
		}
		needChange := false
		var changeStageIds []uint
		for _, sr := range pipelineRunStages {
			if sr.ID == stageRunId {
				// 当前阶段之后的所有阶段都需要修改状态
				needChange = true
				continue
			}
			if needChange && sr.Status == types.PipelineStatusWait {
				// 如果之后有阶段状态为wait，则之后的阶段都不需要修改
				break
			}
			if needChange {
				changeStageIds = append(changeStageIds, sr.ID)
			}
		}
		if len(changeStageIds) > 0 {
			if err = tx.Model(&types.PipelineRunStage{}).Where("id in ?", changeStageIds).Updates(&types.PipelineRunStage{
				Status:     types.PipelineStatusWait,
				Env:        make(types.Map),
				UpdateTime: timeNow,
			}).Error; err != nil {
				return err
			}
		}

		changeStageIds = append(changeStageIds, stageRunId)
		// 修改阶段的所有任务状态为wait
		if err = tx.Model(&types.PipelineRunJob{}).Where("stage_run_id in ?", changeStageIds).Select(
			"status", "update_time", "env", "result").Updates(&types.PipelineRunJob{
			Status:     types.PipelineStatusWait,
			UpdateTime: timeNow,
			Env:        make(types.Map),
			Result:     nil,
		}).Error; err != nil {
			return err
		}

		// 更新流水线构建状态
		pipelineRun.Status = types.PipelineStatusDoing
		pipelineRun.UpdateTime = timeNow
		if err = tx.Select("status", "update_time").Save(&pipelineRun).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	// 发送informer watch通知
	if notifyErr := p.pipelineRunListWatcher.Notify(pipelineRun); notifyErr != nil {
		klog.Warningf("notify pipeline run error: %s", notifyErr.Error())
	}
	return pipelineRun, stageRun, nil
}

func (p *PipelineRunManager) UpdatePipelineRun(pipelineRun *types.PipelineRun) error {
	pipelineRun.UpdateTime = time.Now()
	if err := p.DB.Save(pipelineRun).Error; err != nil {
		return err
	}
	if notifyErr := p.pipelineRunListWatcher.Notify(pipelineRun); notifyErr != nil {
		klog.Warningf("notify pipeline run error: %s", notifyErr.Error())
	}
	return nil
}

func (p *PipelineRunManager) GetEnvBeforeStageRun(stageRun *types.PipelineRunStage) (envs map[string]interface{}, err error) {
	if stageRun.PrevStageRunId == 0 {
		var pipelineRun types.PipelineRun
		if err = p.DB.Last(&pipelineRun, "id = ?", stageRun.PipelineRunId).Error; err != nil {
			return nil, err
		}
		envs = pipelineRun.Env
	} else {
		var prevStageRun types.PipelineRunStage
		if err = p.DB.Last(&prevStageRun, "id = ? and pipeline_run_id = ?", stageRun.PrevStageRunId, stageRun.PipelineRunId).Error; err != nil {
			return nil, err
		}
		envs = prevStageRun.Env
	}
	for k, v := range stageRun.CustomParams {
		envs[k] = v
	}
	return envs, nil
}

func (p *PipelineRunManager) GetJobRunLog(jobRunId uint, withLog bool) (*types.PipelineRunJobLog, error) {
	var jobLog types.PipelineRunJobLog
	if withLog {
		if err := p.DB.Last(&jobLog, "job_run_id = ?", jobRunId).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			return nil, err
		}
	} else {
		if err := p.DB.Select("id", "job_run_id", "create_time", "update_time").Last(&jobLog, "job_run_id = ?", jobRunId).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			return nil, err
		}
	}
	return &jobLog, nil
}
