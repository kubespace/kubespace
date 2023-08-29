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

type createHandler struct {
	models     *model.Models
	appService *projectservice.AppService
}

func CreateHandler(conf *config.ServerConfig) api.Handler {
	return &createHandler{
		models:     conf.Models,
		appService: conf.ServiceFactory.Project.AppService,
	}
}

func (h *createHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	var form projectservice.CreateAppForm
	if err := c.ShouldBindBodyWith(&form, binding.JSON); err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   form.Scope,
		ScopeId: form.ScopeId,
		Role:    types.RoleEditor,
	}, nil
}

func (h *createHandler) Handle(c *api.Context) *utils.Response {
	var form projectservice.CreateAppForm
	if err := c.ShouldBindBodyWith(&form, binding.JSON); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	form.User = c.User.Name
	app, err := h.appService.CreateProjectApp(&form)
	resp := c.Response(err, app)
	var opScope, scopeName, opDetail, opNamespace string
	var resId uint
	if app != nil {
		resId = app.ID
	}
	if form.Scope == types.ScopeProject {
		opDetail = fmt.Sprintf("应用%s创建新版本：%s-%s", form.Name, form.Name, form.Version)
		projectObj, err := h.models.ProjectManager.Get(form.ScopeId)
		if err != nil {
			return &utils.Response{Code: code.GetError, Msg: err.Error()}
		}
		scopeName = projectObj.Name
		opScope = types.ScopeProject
		opNamespace = projectObj.Namespace
	}
	form.ChartFiles = nil
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationCreate,
		OperateDetail:        opDetail,
		Scope:                opScope,
		ScopeId:              form.ScopeId,
		ScopeName:            scopeName,
		Namespace:            opNamespace,
		ResourceId:           resId,
		ResourceType:         types.AuditResourceAppVersion,
		ResourceName:         form.Name + "-" + form.Version,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: form,
	})
	return resp
}
