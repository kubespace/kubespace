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

type importStoreAppHandler struct {
	models     *model.Models
	appService *projectservice.AppService
}

func ImportStoreAppHandler(conf *config.ServerConfig) api.Handler {
	return &importStoreAppHandler{
		models:     conf.Models,
		appService: conf.ServiceFactory.Project.AppService,
	}
}

func (h *importStoreAppHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	var form projectservice.ImportStoreAppForm
	if err := c.ShouldBindBodyWith(&form, binding.JSON); err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   form.Scope,
		ScopeId: form.ScopeId,
		Role:    types.RoleEditor,
	}, nil
}

func (h *importStoreAppHandler) Handle(c *api.Context) *utils.Response {
	var form projectservice.ImportStoreAppForm
	if err := c.ShouldBindBodyWith(&form, binding.JSON); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	form.User = c.User.Name
	app, appVersion, err := h.appService.ImportStoreApp(&form)
	resp := c.ResponseError(err)
	if app == nil {
		return resp
	}
	var opDetail, opScopeName, opNamespace, opResType string
	if app.Scope == types.ScopeProject {
		projectObj, err := h.models.ProjectManager.Get(app.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.GetError, Msg: fmt.Sprintf("获取应用所在工作空间失败：%s", err.Error())}
		}
		opNamespace = projectObj.Namespace
		opScopeName = projectObj.Name
		opResType = types.AuditResourceApp
		opDetail = fmt.Sprintf("从应用商店导入应用：%s-%s", appVersion.PackageName, appVersion.PackageVersion)
	} else {
		clusterObj, err := h.models.ClusterManager.GetById(app.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.DBError, Msg: fmt.Sprintf("获取集群id=%d失败：%s", app.ScopeId, err.Error())}
		}
		opNamespace = app.Namespace
		opScopeName = clusterObj.Name1
		opResType = types.AuditResourceClusterComponent
		opDetail = fmt.Sprintf("从应用商店导入集群组件：%s-%s", appVersion.PackageName, appVersion.PackageVersion)
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationImport,
		OperateDetail:        opDetail,
		Scope:                app.Scope,
		ScopeId:              app.ScopeId,
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
