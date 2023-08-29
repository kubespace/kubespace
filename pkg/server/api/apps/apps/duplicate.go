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

type duplicateHandler struct {
	models     *model.Models
	appService *projectservice.AppService
}

func DuplicateHandler(conf *config.ServerConfig) api.Handler {
	return &duplicateHandler{
		models:     conf.Models,
		appService: conf.ServiceFactory.Project.AppService,
	}
}

func (h *duplicateHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	var form projectservice.DuplicateAppForm
	if err := c.ShouldBindBodyWith(&form, binding.JSON); err != nil {
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

func (h *duplicateHandler) Handle(c *api.Context) *utils.Response {
	var ser projectservice.DuplicateAppForm
	if err := c.ShouldBindBodyWith(&ser, binding.JSON); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	ser.User = c.User.Name
	app, appVersion, err := h.appService.DuplicateApp(&ser)
	resp := c.ResponseError(err)
	if app == nil {
		return resp
	}
	var opDetail, operation string
	projectObj, err := h.models.ProjectManager.Get(app.ScopeId)
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: fmt.Sprintf("获取应用所在工作空间失败：%s", err.Error())}
	}
	if ser.Scope == types.ScopeProject {
		destProject, err := h.models.ProjectManager.Get(ser.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.GetError, Msg: fmt.Sprintf("获取目标工作空间失败：%s", err.Error())}
		}
		operation = types.AuditOperationClone
		opDetail = fmt.Sprintf("克隆应用版本：%s-%s到目标工作空间：%s", appVersion.PackageName, appVersion.PackageVersion, destProject.Name)
	} else {
		operation = types.AuditOperationRelease
		opDetail = fmt.Sprintf("发布应用版本：%s-%s到应用商店", appVersion.PackageName, appVersion.PackageVersion)
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            operation,
		OperateDetail:        opDetail,
		Scope:                types.ScopeProject,
		ScopeId:              projectObj.ID,
		ScopeName:            projectObj.Name,
		Namespace:            projectObj.Namespace,
		ResourceId:           app.ID,
		ResourceType:         types.AuditResourceApp,
		ResourceName:         app.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: nil,
	})
	return resp
}
