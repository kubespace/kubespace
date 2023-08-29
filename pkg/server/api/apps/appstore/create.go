package appstore

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	projectservice "github.com/kubespace/kubespace/pkg/service/project"
	"github.com/kubespace/kubespace/pkg/utils"
	"io"
)

type createHandler struct {
	models          *model.Models
	appStoreService *projectservice.AppStoreService
}

func CreateHandler(conf *config.ServerConfig) api.Handler {
	return &createHandler{
		models:          conf.Models,
		appStoreService: conf.ServiceFactory.Project.AppStoreService,
	}
}

func (h *createHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, &api.AuthPerm{
		Scope:   types.ScopePlatform,
		ScopeId: 0,
		Role:    types.RoleEditor,
	}, nil
}

func (h *createHandler) Handle(c *api.Context) *utils.Response {
	var form projectservice.CreateStoreAppForm
	if err := c.ShouldBind(&form); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	form.User = c.User.Name

	file, err := c.FormFile("file")
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, "get chart file error: "+err.Error()))
	}
	chartIn, err := file.Open()
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, "open chart file error: "+err.Error()))
	}
	form.ChartBytes, err = io.ReadAll(chartIn)
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, "read chart file error: "+err.Error()))
	}

	if icon, err := c.FormFile("icon"); err == nil {
		if iconIn, err := icon.Open(); err == nil {
			form.IconBytes, _ = io.ReadAll(iconIn)
		}
	}
	app, appVersion, err := h.appStoreService.CreateStoreApp(&form)
	resp := c.ResponseError(err)
	if app == nil {
		return resp
	}
	var resId uint
	if appVersion != nil {
		resId = appVersion.ID
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationCreate,
		OperateDetail:        fmt.Sprintf("应用商店导入新应用版本：%s-%s", form.Name, form.PackageVersion),
		Scope:                types.ScopeAppStore,
		ScopeId:              app.ID,
		ScopeName:            app.Name,
		ResourceId:           resId,
		ResourceType:         types.AuditResourceAppVersion,
		ResourceName:         form.Name + "-" + form.PackageVersion,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: form,
	})
	return resp
}
