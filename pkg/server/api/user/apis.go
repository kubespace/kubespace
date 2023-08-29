package user

import (
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/api/user/role"
	"github.com/kubespace/kubespace/pkg/server/api/user/user"
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
		api.NewApi(http.MethodPost, "/login", user.LoginHandler(a.config)),
		api.NewApi(http.MethodGet, "/has_admin", user.HasAdminHandler(a.config)),
		api.NewApi(http.MethodPost, "/admin", user.CreateAdminHandler(a.config)),
		api.NewApi(http.MethodPost, "/logout", user.LogoutHandler(a.config)),

		api.NewApi(http.MethodGet, "", user.ListHandler(a.config)),
		api.NewApi(http.MethodPost, "", user.CreateHandler(a.config)),
		//api.NewApi(http.MethodPut, "", user.UpdateSelfHandler(a.config)),
		api.NewApi(http.MethodPut, "/:username", user.UpdateHandler(a.config)),

		api.NewApi(http.MethodGet, "/token", user.TokenHandler(a.config)),
		api.NewApi(http.MethodPost, "/delete", user.DeleteHandler(a.config)),
		api.NewApi(http.MethodPost, "/update_password", user.UpdatePasswordHandler(a.config)),

		api.NewApi(http.MethodGet, "/role", role.ListHandler(a.config)),
		api.NewApi(http.MethodPost, "/role", role.UpdateHandler(a.config)),
		api.NewApi(http.MethodDelete, "/role/:id", role.DeleteHandler(a.config)),
	}
	return apis
}
