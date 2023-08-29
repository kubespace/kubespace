package cluster

import (
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/api/cluster/agent"
	"github.com/kubespace/kubespace/pkg/server/api/cluster/cluster"
	"github.com/kubespace/kubespace/pkg/server/api/cluster/resource"
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
		api.NewApi(http.MethodGet, "", cluster.ListHandler(a.config)),
		api.NewApi(http.MethodPost, "", cluster.CreateHandler(a.config)),
		api.NewApi(http.MethodPut, "/:id", cluster.UpdateHandler(a.config)),
		api.NewApi(http.MethodDelete, "/:id", cluster.DeleteHandler(a.config)),

		// agent连接请求
		api.NewApi(http.MethodGet, "/agent/connect", agent.ConnectHandler(a.config)),
		api.NewApi(http.MethodGet, "/agent/response", agent.ResponseHandler(a.config)),

		// 集群kubernetes资源操作
		api.NewApi(http.MethodPost, "/:id/apply", resource.ApplyHandler(a.config)),
		api.NewApi(http.MethodPost, "/:id/:resType/list", resource.ListHandler(a.config)),
		api.NewApi(http.MethodGet, "/:id/:resType/watch", resource.WatchHandler(a.config)),
		api.NewApi(http.MethodGet, "/:id/:resType/namespace/:namespace/:name", resource.GetHandler(a.config)),
		api.NewApi(http.MethodGet, "/:id/:resType/:name", resource.GetHandler(a.config)),
		api.NewApi(http.MethodPost, "/:id/:resType/delete", resource.DeleteHandler(a.config)),
		api.NewApi(http.MethodGet, "/:id/pod/exec/:namespace/:pod", resource.PodExecHandler(a.config)),
		api.NewApi(http.MethodGet, "/:id/pod/log/:namespace/:pod", resource.PodLogHandler(a.config)),
		api.NewApi(http.MethodPut, "/:id/:resType/namespace/:namespace/:name", resource.UpdateHandler(a.config)),
		api.NewApi(http.MethodPut, "/:id/:resType/:name", resource.UpdateHandler(a.config)),
		api.NewApi(http.MethodPost, "/:id/:resType/patch", resource.PatchHandler(a.config)),
	}
	return apis
}
