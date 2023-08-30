package spacelet

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
		api.NewApi(http.MethodPut, "/:id", UpdateHandler(a.config)),
		api.NewApi(http.MethodDelete, "/:id", DeleteHandler(a.config)),
		api.NewApi(http.MethodGet, "/install.sh", InstallHandler(a.config)),

		api.NewApi(http.MethodPost, "/register", RegisterHandler(a.config)),
		// spacelet执行完成任务之后回调
		api.NewApi(http.MethodPost, "/pipeline/callback", CallbackHandler(a.config)),
		// 发布任务执行时添加版本号
		api.NewApi(http.MethodPost, "/pipeline/add_release", AddReleaseHandler(a.config)),
	}
	return apis
}
