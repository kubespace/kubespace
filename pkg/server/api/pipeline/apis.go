package pipeline

import (
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/api/pipeline/pipeline"
	"github.com/kubespace/kubespace/pkg/server/api/pipeline/pipeline_run"
	"github.com/kubespace/kubespace/pkg/server/api/pipeline/pipespace"
	"github.com/kubespace/kubespace/pkg/server/api/pipeline/resource"
	"github.com/kubespace/kubespace/pkg/server/config"
	"net/http"
)

type apiGroup struct {
	config *config.ServerConfig
}

func ApiGroup(conf *config.ServerConfig) api.ApiGroup {
	return &apiGroup{conf}
}

func (a *apiGroup) Apis() []*api.Api {
	apis := []*api.Api{
		api.NewApi(http.MethodGet, "/workspace", pipespace.ListHandler(a.config)),
		api.NewApi(http.MethodGet, "/workspace/:id", pipespace.GetHandler(a.config)),
		api.NewApi(http.MethodPost, "/workspace", pipespace.CreateHandler(a.config)),
		api.NewApi(http.MethodPut, "/workspace/:id", pipespace.UpdateHandler(a.config)),
		api.NewApi(http.MethodDelete, "workspace/:id", pipespace.DeleteHandler(a.config)),
		// 获取代码仓库
		api.NewApi(http.MethodGet, "/workspace/list_git_repos", pipespace.ListGitReposHandler(a.config)),
		// 流水线空间最新的发布版本号
		api.NewApi(http.MethodGet, "/workspace/latest_release", pipespace.LatestReleaseHandler(a.config)),
		// 流水线空间是否已存在发布版本号
		api.NewApi(http.MethodGet, "/workspace/exists_release", pipespace.ExistReleaseHandler(a.config)),

		api.NewApi(http.MethodGet, "/pipeline", pipeline.ListHandler(a.config)),
		api.NewApi(http.MethodGet, "/pipeline/:id", pipeline.GetHandler(a.config)),
		api.NewApi(http.MethodGet, "/pipeline/:id/sse", pipeline.WatchHandler(a.config)),
		api.NewApi(http.MethodPost, "/pipeline", pipeline.CreateHandler(a.config)),
		api.NewApi(http.MethodPut, "/pipeline/:id", pipeline.UpdateHandler(a.config)),
		api.NewApi(http.MethodDelete, "/pipeline/:id", pipeline.DeleteHandler(a.config)),
		api.NewApi(http.MethodGet, "/pipeline/:id/list_repo_branch", pipeline.ListRepoBranchHandler(a.config)),

		api.NewApi(http.MethodGet, "/build/list", pipeline_run.ListHandler(a.config)),
		api.NewApi(http.MethodGet, "/build/:id", pipeline_run.GetHandler(a.config)),
		api.NewApi(http.MethodGet, "/build/:id/sse", pipeline_run.WatchHandler(a.config)),
		api.NewApi(http.MethodPost, "/build/:id", pipeline_run.BuildHandler(a.config)),
		api.NewApi(http.MethodPost, "/build/stage_action", pipeline_run.StageActionHandler(a.config)),
		api.NewApi(http.MethodGet, "/build/log/:jobRunId", pipeline_run.JobLogHandler(a.config)),
		api.NewApi(http.MethodGet, "/build/log/:jobRunId/sse", pipeline_run.JobLogStreamHandler(a.config)),

		api.NewApi(http.MethodGet, "/resource/:workspaceId", resource.ListHandler(a.config)),
		api.NewApi(http.MethodPost, "/resource", resource.CreateHandler(a.config)),
		api.NewApi(http.MethodPut, "/resource/:id", resource.UpdateHandler(a.config)),
		api.NewApi(http.MethodDelete, "/resource/:id", resource.DeleteHandler(a.config)),
	}
	return apis
}
