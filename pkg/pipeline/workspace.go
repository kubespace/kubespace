package pipeline

import (
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"regexp"
	"time"
)

type WorkspaceService struct {
	models *model.Models
}

func NewWorkspaceService(models *model.Models) *WorkspaceService {
	return &WorkspaceService{
		models: models,
	}
}

func (w *WorkspaceService) getCodeName(codeType string, codeUrl string) string {
	var re *regexp.Regexp
	if codeType == types.WorkspaceCodeTypeHttps {
		re, _ = regexp.Compile("http[s]+://[\\w\\.:]+/([\\w/\\-_]+)[.git]*")
	} else if codeType == types.WorkspaceCodeTypeGit {
		re, _ = regexp.Compile("git@[\\w\\.]+:[\\d]*/?([\\w/\\-_]+)[\\.git]*")
	} else {
		return ""
	}
	codeName := re.FindStringSubmatch(codeUrl)
	if len(codeName) < 2 {
		return ""
	}
	return codeName[1]
}

func (w *WorkspaceService) checkCodeUrl(codeType string, codeUrl string) bool {
	var re *regexp.Regexp
	if codeType == types.WorkspaceCodeTypeHttps {
		re, _ = regexp.Compile("http[s]+://[\\w.:]+/[\\w/]+[.git]+")
	} else if codeType == types.WorkspaceCodeTypeGit {
		re, _ = regexp.Compile("git@[\\w.:]+/[\\w/]+[.git]+")
	} else {
		return false
	}
	return re.MatchString(codeUrl)
}

func (w *WorkspaceService) defaultCodePipelines() ([]*types.Pipeline, error) {
	branchPipeline := &types.Pipeline{
		Name:       "分支流水线",
		CreateUser: "admin",
		UpdateUser: "admin",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		Triggers: types.PipelineTriggers{
			&types.PipelineTrigger{
				Type:       types.WorkspaceTypeCode,
				BranchType: types.PipelineBranchTypeBranch,
				Operator:   types.PipelineTriggerOperatorExclude,
				Branch:     "master",
			},
		},
		Stages: []*types.PipelineStage{
			{
				Name:        "构建代码镜像",
				TriggerMode: types.StageTriggerModeAuto,
				Jobs: types.PipelineJobs{
					&types.PipelineJob{
						Name:      "代码构建镜像",
						PluginKey: types.BuiltinPluginBuildCodeToImage,
						Params:    map[string]interface{}{},
					},
				},
			},
		},
	}
	masterPipeline := &types.Pipeline{
		Name:       "主干流水线",
		CreateUser: "admin",
		UpdateUser: "admin",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		Triggers: types.PipelineTriggers{
			&types.PipelineTrigger{
				Type:       types.WorkspaceTypeCode,
				BranchType: types.PipelineBranchTypeBranch,
				Operator:   types.PipelineTriggerOperatorEqual,
				Branch:     "master",
			},
		},
		Stages: []*types.PipelineStage{
			{
				Name:        "构建代码镜像",
				TriggerMode: types.StageTriggerModeAuto,
				Jobs: types.PipelineJobs{
					&types.PipelineJob{
						Name:      "代码构建镜像",
						PluginKey: types.BuiltinPluginBuildCodeToImage,
						Params:    map[string]interface{}{},
					},
				},
			},
			{
				Name:        "发布",
				TriggerMode: types.StageTriggerModeManual,
				Jobs: types.PipelineJobs{
					&types.PipelineJob{
						Name:      "发布",
						PluginKey: types.BuiltinPluginRelease,
						Params:    map[string]interface{}{},
					},
				},
			},
		},
	}
	return []*types.Pipeline{branchPipeline, masterPipeline}, nil
}

func (w *WorkspaceService) Create(workspaceSer *serializers.WorkspaceSerializer, user *types.User) *utils.Response {
	var err error
	if !w.checkCodeUrl(workspaceSer.CodeType, workspaceSer.CodeUrl) {
		return &utils.Response{Code: code.ParamsError, Msg: "代码地址格式不正确"}
	}
	workspace := &types.PipelineWorkspace{
		Name:         workspaceSer.Name,
		Description:  workspaceSer.Description,
		Type:         workspaceSer.Type,
		CodeType:     workspaceSer.CodeType,
		CodeUrl:      workspaceSer.CodeUrl,
		CodeSecretId: workspaceSer.CodeSecretId,
		CreateUser:   user.Name,
		UpdateUser:   user.Name,
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
	}
	if workspace.Type == types.WorkspaceTypeCode {
		workspace.Name = w.getCodeName(workspace.CodeType, workspace.CodeUrl)
		if workspace.Name == "" {
			return &utils.Response{Code: code.ParamsError, Msg: "解析代码地址失败，未获取到代码库名称"}
		}
	}
	resp := &utils.Response{Code: code.Success}
	var defaultPipeline []*types.Pipeline
	if workspace.Type == types.WorkspaceTypeCode {
		defaultPipeline, err = w.defaultCodePipelines()
		if err != nil {
			return &utils.Response{Code: code.CreateError, Msg: "创建默认流水线失败: " + err.Error()}
		}
	}
	workspace, err = w.models.PipelineWorkspaceManager.Create(workspace, defaultPipeline)
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = err.Error()
		return resp
	}
	resp.Data = workspace
	return resp
}
