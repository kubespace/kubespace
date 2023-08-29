package project

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
		api.NewApi(http.MethodGet, "/resources", ResourcesHandler(a.config)),
		api.NewApi(http.MethodGet, "/:id", GetHandler(a.config)),
		api.NewApi(http.MethodPost, "", CreateHandler(a.config)),
		api.NewApi(http.MethodPost, "/clone", CloneHandler(a.config)),
		api.NewApi(http.MethodPut, "/:id", UpdateHandler(a.config)),
		api.NewApi(http.MethodDelete, "/:id", DeleteHandler(a.config)),
	}
	return apis
}
