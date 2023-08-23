package pipeline_run

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/service/pipeline/schemas"
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

func (r *PipelineRunService) List(pipelineId uint, lastBuildNumber int, status string, limit int) *utils.Response {
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

func (r *PipelineRunService) Get(pipelineRunId uint) *utils.Response {
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

// JobCallback spacelet节点执行完成任务后进行回调，不写数据库，通知controller-manager
func (r *PipelineRunService) JobCallback(params *schemas.JobCallbackParams) *utils.Response {
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
