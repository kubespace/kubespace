package apps

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
	appId, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	app, err := h.models.AppManager.GetById(appId)
	if err != nil {
		return true, nil, errors.New(code.DataNotExists, "not found app")
	}
	return true, &api.AuthPerm{
		Scope:   app.Scope,
		ScopeId: app.ScopeId,
		Role:    types.RoleEditor,
	}, nil
}

func (h *deleteHandler) Handle(c *api.Context) *utils.Response {
	appId, _ := utils.ParseUint(c.Param("id"))
	app, _ := h.models.AppManager.GetById(appId)

	var opDetail string
	var opScopeName, opNamespace string
	var resType = types.AuditResourceApp
	if app.Scope == types.ScopeProject {
		if projectObj, err := h.models.ProjectManager.Get(app.ScopeId); err != nil {
			return c.ResponseError(errors.New(code.DBError, fmt.Sprintf("获取工作空间id=%d失败：%v", app.ScopeId, err)))
		} else {
			opScopeName = projectObj.Name
			opNamespace = projectObj.Namespace
		}
		opDetail = fmt.Sprintf("删除应用%s，以及所有版本", app.Name)
	} else if app.Scope == types.ScopeCluster {
		if clusterObj, err := h.models.ClusterManager.GetById(app.ScopeId); err != nil {
			return c.ResponseError(errors.New(code.DBError, fmt.Sprintf("获取集群id=%d失败：%s", app.ScopeId, err)))
		} else {
			opScopeName = clusterObj.Name1
		}
		opNamespace = app.Namespace
		opDetail = fmt.Sprintf("删除集群组件%s", app.Name)
		resType = types.AuditResourceClusterComponent
	}
	err := h.models.AppManager.DeleteApp(appId)
	if err != nil {
		err = errors.New(code.DBError, err)
	}
	resp := c.ResponseError(err)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationDelete,
		OperateDetail:        opDetail,
		Scope:                app.Scope,
		ScopeId:              app.ScopeId,
		ScopeName:            opScopeName,
		Namespace:            opNamespace,
		ResourceId:           app.ID,
		ResourceType:         resType,
		ResourceName:         app.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: nil,
	})
	return resp
}
