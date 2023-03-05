package plugins

import (
	"errors"
	"fmt"
	"github.com/kubespace/kubespace/pkg/informer"
	pipelinelistwatcher "github.com/kubespace/kubespace/pkg/informer/listwatcher/pipeline"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/manager/pipeline"
	spaceletmanager "github.com/kubespace/kubespace/pkg/model/manager/spacelet"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/service/spacelet"
	"github.com/kubespace/kubespace/pkg/spacelet/pipeline_job"
	"k8s.io/klog/v2"
	"time"
)

const SpaceletJobStatusInterval = time.Second * 5

// SpaceletJob spacelet节点执行pipeline job
type SpaceletJob struct {
	models          *model.Models
	informerFactory informer.Factory
}

// Execute 创建一个spaceletJob，通过调用spacelet执行pipeline接口执行任务
func (s SpaceletJob) Execute(params *PluginParams) (interface{}, error) {
	client, err := s.getSpaceletClient(params.JobId)
	if err != nil {
		return nil, err
	}
	sj, err := newSpaceletJob(client, params)
	if err != nil {
		return nil, err
	}
	// 监听PipelineRunJob，当spaceletJob执行完成回调时，该informer监听处理
	pipelineRunJobInformer := s.informerFactory.PipelineRunJobInformer(&pipelinelistwatcher.PipelineRunJobWatchCondition{
		WithList: false,
		Id:       params.JobId,
		StatusIn: nil,
	})
	pipelineRunJobInformer.AddHandler(sj)
	stopCh := make(chan struct{})
	// 退出时停止监听
	defer close(stopCh)
	// 开始监听PipelineRunJob对象
	go pipelineRunJobInformer.Run(stopCh)
	return sj.execute()
}

func (s SpaceletJob) getSpaceletClient(jobId uint) (spacelet.Client, error) {
	job, err := s.models.PipelineRunManager.GetJobRun(jobId)
	if err != nil {
		return nil, err
	}
	var sp *types.Spacelet
	if job.SpaceletId == 0 {
		// 分配一个执行任务数最少的spacelet节点
		if sp, err = s.chooseSpacelet(); err != nil {
			return nil, err
		}
		// 更新jobRun spaceletId
		if err = s.models.PipelineRunManager.UpdateJobRun(jobId, &types.PipelineRunJob{SpaceletId: sp.ID}); err != nil {
			return nil, err
		}
	} else {
		if sp, err = s.models.SpaceletManager.Get(job.SpaceletId); err != nil {
			return nil, err
		}
	}
	return spacelet.NewClient(sp)
}

// chooseSpacelet 选择一个spacelet执行任务数最少的在线节点
func (s SpaceletJob) chooseSpacelet() (*types.Spacelet, error) {
	// 查询所有在线的spacelet节点
	spacelets, err := s.models.SpaceletManager.List(&spaceletmanager.SpaceletListCondition{
		Status: types.SpaceletStatusOnline,
	})
	if err != nil {
		return nil, err
	}
	withSpacelet := true
	// 查询所有spacelet节点正在执行的pipeline job
	spaceletJobs, err := s.models.PipelineRunManager.ListJobRun(&pipeline.JobRunListCondition{
		WithSpacelet: &withSpacelet,
		StatusIn:     []string{types.PipelineStatusWait, types.PipelineStatusDoing},
	})
	if err != nil {
		return nil, err
	}
	// 每个spacelet节点执行的任务数，分配一个任务数最少的节点
	spaceletJobMaps := make(map[uint]int)
	for _, job := range spaceletJobs {
		if _, ok := spaceletJobMaps[job.SpaceletId]; ok {
			spaceletJobMaps[job.SpaceletId] += 1
		} else {
			spaceletJobMaps[job.SpaceletId] = 1
		}
	}
	var leastSpacelet *types.Spacelet
	var leastSpaceletNum = -1
	// spacelet节点任务数最少的
	for i, sp := range spacelets {
		num, _ := spaceletJobMaps[sp.ID]
		if leastSpaceletNum == -1 || num < leastSpaceletNum {
			leastSpaceletNum = num
			leastSpacelet = spacelets[i]
		}
	}
	if leastSpacelet == nil {
		return nil, fmt.Errorf("no spacelet node")
	}
	return leastSpacelet, nil
}

type spaceletJob struct {
	params         *PluginParams
	spaceletClient spacelet.Client
	logger         *PluginLogger
	watchCh        chan struct{}
}

func newSpaceletJob(client spacelet.Client, params *PluginParams) (*spaceletJob, error) {
	return &spaceletJob{
		params:         params,
		spaceletClient: client,
		logger:         params.Logger,
		watchCh:        make(chan struct{}),
	}, nil
}

func (s *spaceletJob) execute() (interface{}, error) {
	jobRunParams := &pipeline_job.JobRunParams{
		JobId:  s.params.JobId,
		Plugin: s.params.PluginKey,
		Params: s.params.Params,
	}
	// spacelet执行流水线任务
	if err := s.spaceletClient.PipelineJobExecute(jobRunParams); err != nil {
		return nil, err
	}
	defer func() {
		// 退出时清理流水线任务
		if err := s.spaceletClient.PipelineJobCleanup(&pipeline_job.JobCleanParams{JobId: s.params.JobId}); err != nil {
			klog.Errorf("clean pipeline job error: %s", err.Error())
		}
	}()
	tick := time.NewTicker(SpaceletJobStatusInterval)
	for {
		select {
		case <-tick.C:
			// 定时轮询spacelet节点任务状态
			klog.Infof("interval get pipeline job=%d status", s.params.JobId)
		case <-s.watchCh:
			// 收到spacelet回调
			klog.Infof("watch pipeline job changed and get pipeline job=%d status", s.params.JobId)
		}
		// 查询spacelet节点任务状态接口
		statusLog, err := s.spaceletClient.PipelineJobStatus(&pipeline_job.JobStatusParams{
			JobId:   s.params.JobId,
			WithLog: true,
		})
		if err != nil {
			return nil, err
		}
		// 重置日志内容
		s.logger.ResetWrite(statusLog.Log)
		if statusLog.StatusResult.Status == types.PipelineStatusOK {
			return statusLog.StatusResult.Result.Data, nil
		}
		if statusLog.StatusResult.Status == types.PipelineStatusError {
			return nil, errors.New(statusLog.StatusResult.Result.Msg)
		}
	}
}

func (s *spaceletJob) Check(obj interface{}) bool {
	return true
}

func (s *spaceletJob) Handle(obj interface{}) error {
	// 收到来自spacelet任务完成回调的通知
	s.watchCh <- struct{}{}
	return nil
}
