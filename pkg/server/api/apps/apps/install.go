package apps

import (
	"fmt"
	"github.com/gin-gonic/gin/binding"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	projectservice "github.com/kubespace/kubespace/pkg/service/project"
	"github.com/kubespace/kubespace/pkg/utils"
)

type installHandler struct {
	models     *model.Models
	appService *projectservice.AppService
}

func InstallHandler(conf *config.ServerConfig) api.Handler {
	return &installHandler{
		models:     conf.Models,
		appService: conf.ServiceFactory.Project.AppService,
	}
}

func (h *installHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	var form projectservice.InstallAppForm
	if err := c.ShouldBindBodyWith(&form, binding.JSON); err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	app, err := h.models.AppManager.GetById(form.AppId)
	if err != nil {
		return true, nil, errors.New(code.DataNotExists, err)
	}
	return true, &api.AuthPerm{
		Scope:   app.Scope,
		ScopeId: app.ScopeId,
		Role:    types.RoleEditor,
	}, nil
}

func (h *installHandler) Handle(c *api.Context) *utils.Response {
	var form projectservice.InstallAppForm
	if err := c.ShouldBindBodyWith(&form, binding.JSON); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err.Error()))
	}
	form.User = c.User.Name

	app, versionApp, err := h.appService.InstallApp(&form)
	resp := c.ResponseError(err)

	if app == nil {
		return resp
	}

	var opDetail, opScope, opScopeName, opNamespace, opResType string
	var opScopeId uint
	opStr := "安装"
	operation := types.AuditOperationInstall
	if form.Upgrade {
		opStr = "升级"
		operation = types.AuditOperationUpgrade
	}
	if app.Scope == types.ScopeProject {
		projectObj, err := h.models.ProjectManager.Get(app.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.GetError, Msg: fmt.Sprintf("获取应用所在工作空间失败：%s", err.Error())}
		}
		opScope = types.ScopeProject
		opScopeId = projectObj.ID
		opNamespace = projectObj.Namespace
		opScopeName = projectObj.Name
		opResType = types.AuditResourceApp
		opDetail = fmt.Sprintf("%s应用%s版本：%s-%s", opStr, app.Name, versionApp.PackageName, versionApp.PackageVersion)
	} else {
		clusterObj, err := h.models.ClusterManager.GetById(app.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: fmt.Sprintf("获取集群id=%d失败：%s", app.ScopeId, err.Error())}
		}
		opScope = types.ScopeCluster
		opScopeId = clusterObj.ID
		opNamespace = app.Namespace
		opScopeName = clusterObj.Name1
		opResType = types.AuditResourceClusterComponent
		opDetail = fmt.Sprintf("%s集群组件：%s-%s", opStr, versionApp.PackageName, versionApp.PackageVersion)
	}

	c.CreateAudit(&types.AuditOperate{
		Operation:            operation,
		OperateDetail:        opDetail,
		Scope:                opScope,
		ScopeId:              opScopeId,
		ScopeName:            opScopeName,
		Namespace:            opNamespace,
		ResourceId:           app.ID,
		ResourceType:         opResType,
		ResourceName:         app.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: nil,
	})
	return resp
}
