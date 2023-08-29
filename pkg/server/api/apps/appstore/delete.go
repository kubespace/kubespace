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
)

type deleteHandler struct {
	models          *model.Models
	appStoreService *projectservice.AppStoreService
}

func DeleteHandler(conf *config.ServerConfig) api.Handler {
	return &deleteHandler{
		models:          conf.Models,
		appStoreService: conf.ServiceFactory.Project.AppStoreService,
	}
}

func (h *deleteHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, &api.AuthPerm{
		Scope:   types.ScopePlatform,
		ScopeId: 0,
		Role:    types.RoleEditor,
	}, nil
}

func (h *deleteHandler) Handle(c *api.Context) *utils.Response {
	appId, err := utils.ParseUint(c.Param("appId"))
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	versionId, err := utils.ParseUint(c.Param("versionId"))
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	app, err := h.models.AppStoreManager.GetById(appId)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, "获取应用商店应用id=%d失败："+err.Error()))
	}
	appVersion, err := h.models.AppVersionManager.GetById(versionId)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, fmt.Sprintf("获取应用商店应用%s版本id=%d失败：%s", app.Name, versionId, err.Error())))
	}
	err = h.models.AppStoreManager.DeleteStoreAppVersion(appId, versionId, c.User.Name)
	resp := c.ResponseError(err)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationDelete,
		OperateDetail:        fmt.Sprintf("应用商店删除应用版本：%s-%s", appVersion.PackageName, appVersion.PackageVersion),
		Scope:                types.ScopeAppStore,
		ScopeId:              app.ID,
		ScopeName:            app.Name,
		ResourceId:           appVersion.ID,
		ResourceType:         types.AuditResourceAppVersion,
		ResourceName:         appVersion.PackageName + "-" + appVersion.PackageVersion,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: nil,
	})
	return resp
}
