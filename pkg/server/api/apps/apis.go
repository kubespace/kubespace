package apps

import (
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/api/apps/apps"
	"github.com/kubespace/kubespace/pkg/server/api/apps/version"
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
		api.NewApi(http.MethodGet, "", apps.ListHandler(a.config)),
		api.NewApi(http.MethodGet, "/status", apps.StatusHandler(a.config)),
		api.NewApi(http.MethodGet, "/status_sse", apps.StatusSSEHandler(a.config)),
		api.NewApi(http.MethodGet, "/download", apps.DownloadHandler(a.config)),
		api.NewApi(http.MethodGet, "/:id", apps.GetHandler(a.config)),
		api.NewApi(http.MethodPost, "", apps.CreateHandler(a.config)),
		api.NewApi(http.MethodPost, "/install", apps.InstallHandler(a.config)),
		api.NewApi(http.MethodPost, "/destroy", apps.DestroyHandler(a.config)),
		api.NewApi(http.MethodPost, "/import_storeapp", apps.ImportStoreAppHandler(a.config)),
		api.NewApi(http.MethodPost, "/import_custom_app", apps.ImportCustomAppHandler(a.config)),
		api.NewApi(http.MethodPost, "/duplicate_app", apps.DuplicateHandler(a.config)),
		api.NewApi(http.MethodDelete, "/:id", apps.DeleteHandler(a.config)),

		api.NewApi(http.MethodGet, "/versions", version.ListHandler(a.config)),
		api.NewApi(http.MethodGet, "/version/:id", version.GetHandler(a.config)),
		api.NewApi(http.MethodGet, "/version/:id/chartfiles", version.ChartFilesHandler(a.config)),
		api.NewApi(http.MethodDelete, "/version/:id", version.DeleteHandler(a.config)),
	}
	return apis
}
