package pipeline

import (
	"context"
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/third/git"
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
	if codeType == types.WorkspaceCodeTypeGit {
		re, _ = regexp.Compile("git@[\\w\\.]+:/?([\\w/\\-_]+)[\\.git]*")
	} else {
		re, _ = regexp.Compile("http[s]?://[\\w\\.:]+/([\\w/\\-_]+)[.git]*")
	}
	codeName := re.FindStringSubmatch(codeUrl)
	if len(codeName) < 2 {
		return ""
	}
	return codeName[1]
}

func (w *WorkspaceService) checkCodeUrl(codeType string, codeUrl string) bool {
	var re *regexp.Regexp
	if codeType == types.WorkspaceCodeTypeGit {
		re, _ = regexp.Compile("git@[\\w\\.]+:/?([\\w/\\-_]+)[\\.git]*")
	} else {
		re, _ = regexp.Compile("http[s]?://[\\w\\.:]+/([\\w/\\-_]+)[.git]*")
	}
	return re.MatchString(codeUrl)
}

func (w *WorkspaceService) defaultCodePipelines(pipespace *types.PipelineWorkspace) ([]*types.Pipeline, error) {
	branchPipeline := &types.Pipeline{
		Name:       "分支流水线",
		CreateUser: pipespace.CreateUser,
		UpdateUser: pipespace.UpdateUser,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		Sources: types.PipelineSources{
			&types.PipelineSource{
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
						Name:      "构建代码镜像",
						PluginKey: types.BuiltinPluginBuildCodeToImage,
						Params:    map[string]interface{}{},
					},
				},
			},
		},
	}
	masterPipeline := &types.Pipeline{
		Name:       "主干流水线",
		CreateUser: pipespace.CreateUser,
		UpdateUser: pipespace.UpdateUser,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		Sources: types.PipelineSources{
			&types.PipelineSource{
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
						Name:      "构建代码镜像",
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

func (w *WorkspaceService) Create(pipespace *types.PipelineWorkspace) (*types.PipelineWorkspace, error) {
	if pipespace.Type == types.WorkspaceTypeCode {
		if !w.checkCodeUrl(pipespace.Code.Type, pipespace.Code.CloneUrl) {
			return nil, errors.New(code.ParamsError, "代码地址格式不正确")
		}
		pipespace.Name = w.getCodeName(pipespace.Code.Type, pipespace.Code.CloneUrl)
		secret, err := w.models.SettingsSecretManager.Get(pipespace.Code.SecretId)
		if err != nil {
			return nil, errors.New(code.DataNotExists, fmt.Sprintf("获取代码密钥失败：%v", err))
		}
		gitcli, err := git.NewClient(pipespace.Code.Type, pipespace.Code.ApiUrl, &types.Secret{
			Type:        secret.Type,
			User:        secret.User,
			Password:    secret.Password,
			PrivateKey:  secret.PrivateKey,
			AccessToken: secret.AccessToken,
		})
		if err != nil {
			return nil, errors.New(code.GitError, fmt.Sprintf("new git clint error: %v", err))
		}
		// 获取代码仓库分支，验证是否可以连通
		if _, err = gitcli.ListRepoBranches(context.Background(), pipespace.Code.CloneUrl); err != nil {
			return nil, errors.New(code.GitError, fmt.Sprintf("get git branch error: %v", err))
		}
	}
	if pipespace.Name == "" {
		return nil, errors.New(code.ParamsError, "解析代码地址失败，未获取到代码库名称")
	}
	var defaultPipeline []*types.Pipeline
	var err error
	if pipespace.Type == types.WorkspaceTypeCode {
		defaultPipeline, err = w.defaultCodePipelines(pipespace)
		if err != nil {
			return nil, errors.New(code.CreateError, "创建默认流水线失败: "+err.Error())
		}
	}
	pipespace, err = w.models.PipelineWorkspaceManager.Create(pipespace, defaultPipeline)
	if err != nil {
		return nil, errors.New(code.DBError, err)
	}
	return pipespace, nil
}

func (w *WorkspaceService) ListGitRepos(secretId uint, gitType, apiUrl string) ([]*git.Repository, error) {
	secret, err := w.models.SettingsSecretManager.Get(secretId)
	if err != nil {
		return nil, errors.New(code.DataNotExists, "获取密钥失败："+err.Error())
	}
	gitcli, err := git.NewClient(gitType, apiUrl, &types.Secret{
		Type:        secret.Type,
		User:        secret.User,
		Password:    secret.Password,
		PrivateKey:  secret.PrivateKey,
		AccessToken: secret.AccessToken,
	})
	if err != nil {
		return nil, errors.New(code.GitError, "new git clint error: "+err.Error())
	}
	repos, err := gitcli.ListRepositories(context.Background())
	if err != nil {
		return nil, errors.New(code.GitError, "list git repository error: "+err.Error())
	}
	return repos, nil
}
