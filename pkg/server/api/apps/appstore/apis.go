package appstore

import (
	"github.com/kubespace/kubespace/pkg/server/api/api"
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
		api.NewApi(http.MethodGet, "", ListHandler(a.config)),
		api.NewApi(http.MethodGet, "/:id", GetHandler(a.config)),
		api.NewApi(http.MethodPost, "/resolve", ResolveChartHandler(a.config)),
		api.NewApi(http.MethodPost, "/create", CreateHandler(a.config)),
		api.NewApi(http.MethodDelete, "/:appId/:versionId", DeleteHandler(a.config)),
	}
	return apis
}
