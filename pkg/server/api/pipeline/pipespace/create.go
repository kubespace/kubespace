package pipespace

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	pipelineservice "github.com/kubespace/kubespace/pkg/service/pipeline"
	"github.com/kubespace/kubespace/pkg/utils"
	"time"
)

type createHandler struct {
	models           *model.Models
	workspaceService *pipelineservice.WorkspaceService
}

func CreateHandler(conf *config.ServerConfig) api.Handler {
	return &createHandler{
		models:           conf.Models,
		workspaceService: conf.ServiceFactory.Pipeline.WorkspaceService,
	}
}

type createPipespaceBody struct {
	Name         string `json:"name" form:"name"`
	Type         string `json:"type" form:"type"`
	Description  string `json:"description" form:"description"`
	ApiUrl       string `json:"api_url" form:"api_url"`
	CodeUrl      string `json:"code_url" form:"code_url"`
	CodeType     string `json:"code_type" form:"code_type"`
	CodeSecretId uint   `json:"code_secret_id" form:"code_secret_id"`
}

func (h *createHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, &api.AuthPerm{
		Scope:   types.ScopePlatform,
		ScopeId: 0,
		Role:    types.RoleEditor,
	}, nil
}

func (h *createHandler) Handle(c *api.Context) *utils.Response {
	var body createPipespaceBody
	if err := c.ShouldBind(&body); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	workspace := &types.PipelineWorkspace{
		Name:        body.Name,
		Description: body.Description,
		Type:        body.Type,
		Code: &types.PipelineWorkspaceCode{
			Type:     body.CodeType,
			ApiUrl:   body.ApiUrl,
			CloneUrl: body.CodeUrl,
			SecretId: body.CodeSecretId,
		},
		CreateUser: c.User.Name,
		UpdateUser: c.User.Name,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	workspace, err := h.workspaceService.Create(workspace)
	resp := c.Response(err, workspace)
	var scopeId uint
	if workspace != nil {
		scopeId = workspace.ID
	}
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationCreate,
		OperateDetail:        "创建流水线空间:" + body.Name,
		Scope:                types.ScopePipeline,
		ScopeId:              scopeId,
		ScopeName:            body.Name,
		ResourceId:           scopeId,
		ResourceType:         types.AuditResourcePipeSpace,
		ResourceName:         body.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: body,
	})
	return resp
}
