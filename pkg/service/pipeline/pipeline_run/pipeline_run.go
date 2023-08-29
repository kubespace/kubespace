package pipeline_run

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/utils"
)

type PipelineRunService struct {
	models *model.Models
}

func NewPipelineRunService(models *model.Models) *PipelineRunService {
	r := &PipelineRunService{
		models: models,
	}
	return r
}

func (r *PipelineRunService) Get(pipelineRunId uint) *utils.Response {
	pipelineRun, err := r.models.PipelineRunManager.Get(pipelineRunId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	stagesRun, err := r.models.PipelineRunManager.StagesRun(pipelineRunId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	pipelineObj, err := r.models.PipelineManager.GetById(pipelineRun.PipelineId)
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

// JobCallback spacelet节点执行完成任务后进行回调，不写数据库，通知controller-manager
func (r *PipelineRunService) JobCallback(jobId uint, status string) error {
	jobRun, err := r.models.PipelineRunManager.GetJobRun(jobId)
	if err != nil {
		return errors.New(code.DataNotExists, "get job run error: "+err.Error())
	}
	jobRun.Status = status
	// 通知controller-manager
	if err = r.models.PipelineRunManager.NotifyJobRun(jobRun); err != nil {
		return errors.New(code.RedisError, err)
	}
	return nil
}
