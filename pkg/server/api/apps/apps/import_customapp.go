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
	"io/ioutil"
)

type importCustomAppHandler struct {
	models     *model.Models
	appService *projectservice.AppService
}

func ImportCustomAppHandler(conf *config.ServerConfig) api.Handler {
	return &importCustomAppHandler{
		models:     conf.Models,
		appService: conf.ServiceFactory.Project.AppService,
	}
}

func (h *importCustomAppHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	var form projectservice.ImportCustomAppForm
	if err := c.ShouldBindBodyWith(&form, binding.JSON); err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   form.Scope,
		ScopeId: form.ScopeId,
		Role:    types.RoleEditor,
	}, nil
}

func (h *importCustomAppHandler) Handle(c *api.Context) *utils.Response {
	var form projectservice.ImportCustomAppForm
	if err := c.ShouldBindBodyWith(&form, binding.JSON); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	form.User = c.User.Name

	file, err := c.FormFile("file")
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, "get chart file error: "+err.Error()))
	}
	chartIn, err := file.Open()
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, "get chart file error: "+err.Error()))
	}
	form.ChartBytes, err = ioutil.ReadAll(chartIn)
	if err != nil {
		return c.ResponseError(errors.New(code.GetError, "获取chart文件失败: "+err.Error()))
	}
	app, _, err := h.appService.ImportCustomApp(&form)
	resp := c.ResponseError(err)
	if app == nil {
		return resp
	}
	var opDetail, opScope, opScopeName, opNamespace, opResType string
	var opScopeId uint
	if app.Scope == types.ScopeProject {
		projectObj, err := h.models.ProjectManager.Get(app.ScopeId)
		if err != nil {
			return c.ResponseError(errors.New(code.DataNotExists, fmt.Sprintf("获取应用所在工作空间失败：%s", err.Error())))
		}
		opScope = types.ScopeProject
		opScopeId = projectObj.ID
		opNamespace = projectObj.Namespace
		opScopeName = projectObj.Name
		opResType = types.AuditResourceApp
		opDetail = fmt.Sprintf("导入自定义应用：%s-%s", form.Name, form.PackageVersion)
	} else {
		clusterObj, err := h.models.ClusterManager.GetById(app.ScopeId)
		if err != nil {
			return c.ResponseError(errors.New(code.DBError, fmt.Sprintf("获取集群id=%d失败：%s", app.ScopeId, err.Error())))
		}
		opScope = types.ScopeCluster
		opScopeId = clusterObj.ID
		opNamespace = app.Namespace
		opScopeName = clusterObj.Name1
		opResType = types.AuditResourceClusterComponent
		opDetail = fmt.Sprintf("导入自定义集群组件：%s-%s", form.Name, form.PackageVersion)
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationImport,
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
