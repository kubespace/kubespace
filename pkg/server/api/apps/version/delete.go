package version

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
	models     *model.Models
	appService *projectservice.AppService
}

func DeleteHandler(conf *config.ServerConfig) api.Handler {
	return &deleteHandler{
		models:     conf.Models,
		appService: conf.ServiceFactory.Project.AppService,
	}
}

func (h *deleteHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	appVersionId, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	appVersion, err := h.models.AppVersionManager.GetById(appVersionId)
	if err != nil {
		return true, nil, errors.New(code.DataNotExists, err)
	}
	app, err := h.models.AppManager.GetById(appVersion.ScopeId)
	if err != nil {
		return true, nil, errors.New(code.DataNotExists, err)
	}
	if app.Scope == types.ScopeAppStore {
		return true, &api.AuthPerm{
			Scope:   types.ScopePlatform,
			ScopeId: 0,
			Role:    types.RoleEditor,
		}, nil
	}
	return true, &api.AuthPerm{
		Scope:   app.Scope,
		ScopeId: app.ScopeId,
		Role:    types.RoleEditor,
	}, nil
}

func (h *deleteHandler) Handle(c *api.Context) *utils.Response {
	appVersionId, _ := utils.ParseUint(c.Param("id"))
	appVersion, _ := h.models.AppVersionManager.GetById(appVersionId)

	var opScope, opScopeName, opDetail, opNamespace string
	var opScopeId uint
	if appVersion.Scope == types.ScopeProject {
		appObj, err := h.models.AppManager.GetById(appVersion.ScopeId)
		if err != nil {
			return c.ResponseError(errors.New(code.DataNotExists, fmt.Sprintf("获取应用失败：%v", err)))
		}
		projectObj, err := h.models.ProjectManager.Get(appObj.ScopeId)
		if err != nil {
			return c.ResponseError(errors.New(code.DataNotExists, fmt.Sprintf("获取应用所在工作空间失败：%v", err)))
		}
		opScope = types.ScopeProject
		opScopeId = projectObj.ID
		opScopeName = projectObj.Name
		opNamespace = projectObj.Namespace
		opDetail = fmt.Sprintf("删除应用%s所属版本：%s-%s", appObj.Name, appVersion.PackageName, appVersion.PackageVersion)
	} else if appVersion.Scope == types.ScopeAppStore {
		appStoreObj, err := h.models.AppStoreManager.GetById(appVersion.ScopeId)
		if err != nil {
			return c.ResponseError(errors.New(code.DataNotExists, fmt.Sprintf("获取应用商店应用失败：%v", err)))
		}
		opScope = types.ScopeAppStore
		opScopeId = appStoreObj.ID
		opScopeName = appStoreObj.Name
		opDetail = fmt.Sprintf("应用商店删除应用版本：%s-%s", appVersion.PackageName, appVersion.PackageVersion)
	} else if appVersion.Scope == types.ScopeCluster {
		appObj, err := h.models.AppManager.GetById(appVersion.ScopeId)
		if err != nil {
			return c.ResponseError(errors.New(code.DataNotExists, fmt.Sprintf("获取应用id=%d失败：%v", appVersion.ScopeId, err)))
		}
		clusterObj, err := h.models.ClusterManager.GetById(appObj.ScopeId)
		if err != nil {
			return c.ResponseError(errors.New(code.DBError, fmt.Sprintf("获取集群id=%d失败：%v", appObj.ScopeId, err)))
		}
		opScope = types.ScopeCluster
		opScopeId = clusterObj.ID
		opNamespace = appObj.Namespace
		opScopeName = clusterObj.Name1

		opDetail = fmt.Sprintf("删除集群组件版本：%s-%s", appVersion.PackageName, appVersion.PackageVersion)
	}
	err := h.models.AppVersionManager.Delete(appVersionId)
	if err != nil {
		err = errors.New(code.DBError, "删除应用版本失败："+err.Error())
	}
	resp := c.ResponseError(err)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationDelete,
		OperateDetail:        opDetail,
		Scope:                opScope,
		ScopeId:              opScopeId,
		ScopeName:            opScopeName,
		Namespace:            opNamespace,
		ResourceId:           appVersion.ID,
		ResourceType:         types.AuditResourceAppVersion,
		ResourceName:         appVersion.PackageName + "-" + appVersion.PackageVersion,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: nil,
	})
	return resp
}
