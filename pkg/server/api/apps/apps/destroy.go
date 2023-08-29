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

type destroyHandler struct {
	models     *model.Models
	appService *projectservice.AppService
}

func DestroyHandler(conf *config.ServerConfig) api.Handler {
	return &destroyHandler{
		models:     conf.Models,
		appService: conf.ServiceFactory.Project.AppService,
	}
}

type destroyAppForm struct {
	AppId uint `json:"app_id" form:"app_id"`
}

func (h *destroyHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	var form destroyAppForm
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

func (h *destroyHandler) Handle(c *api.Context) *utils.Response {
	var ser destroyAppForm
	if err := c.ShouldBindBodyWith(&ser, binding.JSON); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err.Error()))
	}
	app, err := h.appService.DestroyApp(ser.AppId, c.User.Name)
	resp := c.ResponseError(err)
	if app == nil {
		return resp
	}
	var opDetail, opScopeName, opNamespace, opResType string
	if app.Scope == types.ScopeProject {
		projectObj, err := h.models.ProjectManager.Get(app.ScopeId)
		if err != nil {
			return c.ResponseError(errors.New(code.DataNotExists, fmt.Sprintf("获取应用所在工作空间失败：%s", err.Error())))
		}
		opNamespace = projectObj.Namespace
		opScopeName = projectObj.Name
		opResType = types.AuditResourceApp
		opDetail = fmt.Sprintf("销毁应用：%s", app.Name)
	} else {
		clusterObj, err := h.models.ClusterManager.GetById(app.ScopeId)
		if err != nil {
			return c.ResponseError(errors.New(code.DataNotExists, fmt.Sprintf("获取集群id=%d失败：%s", app.ScopeId, err.Error())))
		}
		opNamespace = app.Namespace
		opScopeName = clusterObj.Name1
		opResType = types.AuditResourceClusterComponent
		opDetail = fmt.Sprintf("销毁集群组件：%s", app.Name)
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationDestroy,
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
