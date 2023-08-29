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
	"k8s.io/klog/v2"
	"time"
)

type statusSSEHandler struct {
	models     *model.Models
	appService *projectservice.AppService
}

func StatusSSEHandler(conf *config.ServerConfig) api.Handler {
	return &statusSSEHandler{
		models:     conf.Models,
		appService: conf.ServiceFactory.Project.AppService,
	}
}

func (h *statusSSEHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	var form appStatusForm
	if err := c.ShouldBindQuery(&form); err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   form.Scope,
		ScopeId: form.ScopeId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *statusSSEHandler) Handle(c *api.Context) *utils.Response {
	var form appStatusForm
	if err := c.ShouldBindQuery(&form); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	w := c.Writer
	clientGone := w.CloseNotify()
	c.SSEvent("message", "{}")
	w.Flush()
	tick := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-clientGone:
			klog.Info("app status client gone")
			return nil
		case <-tick.C:
			tick.Stop()
			status, err := h.appService.GetAppStatus(form.Scope, form.ScopeId, form.Name)
			c.SSEvent("message", c.Response(err, status))
			w.Flush()
			tick.Reset(5 * time.Second)
		}
	}
}
