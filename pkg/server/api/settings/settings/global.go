package settings

import (
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
)

type globalSettingsHandler struct {
	models         *model.Models
	releaseVersion string
}

func GlobalSettingsHandler(conf *config.ServerConfig) api.Handler {
	return &globalSettingsHandler{
		models:         conf.Models,
		releaseVersion: conf.ReleaseVersion,
	}
}

func (h *globalSettingsHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, nil, nil
}

func (h *globalSettingsHandler) Handle(c *api.Context) *utils.Response {
	return c.ResponseOK(map[string]interface{}{
		"release_version": h.releaseVersion,
	})
}
