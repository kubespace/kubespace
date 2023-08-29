package settings

import (
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/api/settings/image_registry"
	"github.com/kubespace/kubespace/pkg/server/api/settings/secret"
	"github.com/kubespace/kubespace/pkg/server/api/settings/settings"
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
		api.NewApi(http.MethodGet, "/global", settings.GlobalSettingsHandler(a.config)),

		api.NewApi(http.MethodGet, "/secret", secret.ListHandler(a.config)),
		api.NewApi(http.MethodPost, "/secret", secret.CreateHandler(a.config)),
		api.NewApi(http.MethodPut, "/secret/:id", secret.UpdateHandler(a.config)),
		api.NewApi(http.MethodDelete, "/secret/:id", secret.DeleteHandler(a.config)),

		api.NewApi(http.MethodGet, "/image_registry", image_registry.ListHandler(a.config)),
		api.NewApi(http.MethodPost, "/image_registry", image_registry.CreateHandler(a.config)),
		api.NewApi(http.MethodPut, "/image_registry/:id", image_registry.UpdateHandler(a.config)),
		api.NewApi(http.MethodDelete, "/image_registry/:id", image_registry.DeleteHandler(a.config)),
	}
	return apis
}
