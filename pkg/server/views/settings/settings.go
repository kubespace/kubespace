package settings

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/utils"
	"net/http"
)

type Settings struct {
	Views          []*views.View
	models         *model.Models
	releaseVersion string
}

func NewSettings(conf *config.ServerConfig) *Settings {
	secret := &Settings{
		models:         conf.Models,
		releaseVersion: conf.ReleaseVersion,
	}
	secret.Views = []*views.View{
		views.NewView(http.MethodGet, "/global", secret.globalSettings),
	}
	return secret
}

func (s *Settings) globalSettings(c *views.Context) *utils.Response {
	return &utils.Response{Code: code.Success, Data: map[string]interface{}{
		"release_version": s.releaseVersion,
	}}
}
