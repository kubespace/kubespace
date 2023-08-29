package apps

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	projectservice "github.com/kubespace/kubespace/pkg/service/project"
	"github.com/kubespace/kubespace/pkg/utils"
	"net/http"
	"strings"
)

type downloadHandler struct {
	models     *model.Models
	appService *projectservice.AppService
}

func DownloadHandler(conf *config.ServerConfig) api.Handler {
	return &downloadHandler{
		models:     conf.Models,
		appService: conf.ServiceFactory.Project.AppService,
	}
}

type downloadChartForm struct {
	Scope   string `json:"scope" form:"scope"`
	ScopeId uint   `json:"scope_id" form:"scope_id"`
	Path    string `json:"path" form:"path"`
}

func (h *downloadHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	var form downloadChartForm
	if err := c.ShouldBindQuery(&form); err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	if form.Scope == types.ScopeAppStore {
		return true, nil, nil
	}
	return true, &api.AuthPerm{
		Scope:   form.Scope,
		ScopeId: form.ScopeId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *downloadHandler) Handle(c *api.Context) *utils.Response {
	var form downloadChartForm
	if err := c.ShouldBindQuery(&form); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	chartPath := form.Path
	if chartPath == "" {
		return c.ResponseError(errors.New(code.ParamsError, "not found chart path"))
	}

	appChart, err := h.models.AppVersionManager.GetChart(chartPath)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, err))
	}
	chartName := chartPath
	s := strings.Split(chartPath, "/")
	if len(s) >= 2 {
		chartName = s[len(s)-1]
	}
	fileContentDisposition := "attachment;filename=\"" + chartName + "\""
	c.Header("Content-Disposition", fileContentDisposition)
	c.Data(http.StatusOK, "application/x-tar", appChart.Content)
	return nil
}
