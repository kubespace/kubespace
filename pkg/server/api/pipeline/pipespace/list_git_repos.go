package pipespace

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	pipelineservice "github.com/kubespace/kubespace/pkg/service/pipeline"
	"github.com/kubespace/kubespace/pkg/utils"
)

type listGitReposHandler struct {
	models           *model.Models
	workspaceService *pipelineservice.WorkspaceService
}

func ListGitReposHandler(conf *config.ServerConfig) api.Handler {
	return &listGitReposHandler{
		models:           conf.Models,
		workspaceService: conf.ServiceFactory.Pipeline.WorkspaceService,
	}
}

type listGitReposForm struct {
	SecretId uint   `json:"secret_id" form:"secret_id"`
	GitType  string `json:"git_type" form:"git_type"`
	ApiUrl   string `json:"api_url" form:"api_url"`
}

func (h *listGitReposHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, nil, nil
}

func (h *listGitReposHandler) Handle(c *api.Context) *utils.Response {
	var form listGitReposForm
	if err := c.ShouldBind(&form); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	repos, err := h.workspaceService.ListGitRepos(form.SecretId, form.GitType, form.ApiUrl)
	return c.Response(err, repos)
}
