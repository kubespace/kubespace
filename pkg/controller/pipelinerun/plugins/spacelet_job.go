package plugins

import (
	"errors"
	"fmt"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/manager/pipeline"
	spaceletmanager "github.com/kubespace/kubespace/pkg/model/manager/spacelet"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/service/spacelet"
	"github.com/kubespace/kubespace/pkg/spacelet/pipeline_job"
	"k8s.io/klog/v2"
	"time"
)

// SpaceletJob spacelet节点执行pipeline job
type SpaceletJob struct {
	models *model.Models
}

func (s SpaceletJob) Execute(params *PluginParams) (interface{}, error) {
	client, err := s.getSpaceletClient(params.JobId)
	if err != nil {
		return nil, err
	}
	sj, err := newSpaceletJob(client, params)
	if err != nil {
		return nil, err
	}
	return sj.execute()
}

func (s SpaceletJob) getSpaceletClient(jobId uint) (spacelet.Client, error) {
	job, err := s.models.PipelineRunManager.GetJobRun(jobId)
	if err != nil {
		return nil, err
	}
	var sp *types.Spacelet
	if job.SpaceletId == 0 {
		if sp, err = s.chooseSpacelet(); err != nil {
			return nil, err
		}
	} else {
		if sp, err = s.models.SpaceletManager.Get(job.SpaceletId); err != nil {
			return nil, err
		}
	}
	return spacelet.NewClient(sp)
}

// chooseSpacelet 选择一个spacelet节点
func (s SpaceletJob) chooseSpacelet() (*types.Spacelet, error) {
	spacelets, err := s.models.SpaceletManager.List(&spaceletmanager.SpaceletListCondition{
		Status: "online",
	})
	if err != nil {
		return nil, err
	}
	withSpacelet := true
	spaceletJobs, err := s.models.PipelineRunManager.ListJobRun(&pipeline.JobRunListCondition{
		WithSpacelet: &withSpacelet,
		StatusIn:     []string{types.PipelineStatusWait, types.PipelineStatusDoing},
	})
	if err != nil {
		return nil, err
	}
	spaceletJobMaps := make(map[uint]int)
	for _, job := range spaceletJobs {
		if _, ok := spaceletJobMaps[job.SpaceletId]; ok {
			spaceletJobMaps[job.SpaceletId] += 1
		} else {
			spaceletJobMaps[job.SpaceletId] = 1
		}
	}
	var leastSpacelet *types.Spacelet
	var leastSpaceletNum = 100
	for i, sp := range spacelets {
		num, _ := spaceletJobMaps[sp.ID]
		if num < leastSpaceletNum {
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
}

func newSpaceletJob(client spacelet.Client, params *PluginParams) (*spaceletJob, error) {
	return &spaceletJob{
		params:         params,
		spaceletClient: client,
		logger:         params.Logger,
	}, nil
}

func (s *spaceletJob) execute() (interface{}, error) {
	jobRunParams := &pipeline_job.JobRunParams{
		JobId:  s.params.JobId,
		Plugin: s.params.PluginKey,
		Params: s.params.Params,
	}
	if err := s.spaceletClient.PipelineJobExecute(jobRunParams); err != nil {
		return nil, err
	}
	defer func() {
		if err := s.spaceletClient.PipelineJobCleanup(&pipeline_job.JobCleanParams{JobId: s.params.JobId}); err != nil {
			klog.Errorf("clean pipeline job error: %s", err.Error())
		}
	}()
	tick := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-tick.C:
			statusLog, err := s.spaceletClient.PipelineJobStatus(&pipeline_job.JobStatusParams{
				JobId:   s.params.JobId,
				WithLog: true,
			})
			if err != nil {
				return nil, err
			}
			klog.Infof("status: %s", statusLog.StatusResult.Status)
			klog.Infof("status result: %s", *statusLog.StatusResult.Result)
			klog.Infof("status log: %s", statusLog.Log)
			s.logger.Log(statusLog.Log)
			if statusLog.StatusResult.Status == types.PipelineStatusOK {
				return statusLog.StatusResult.Result.Data, nil
			}
			if statusLog.StatusResult.Status == types.PipelineStatusError {
				return nil, errors.New(statusLog.StatusResult.Result.Msg)
			}
		}
	}
}
