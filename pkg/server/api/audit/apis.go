package audit

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
	}
	return apis
}
