package views

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
	"time"
)

type View struct {
	Method  string
	Path    string
	Handler ViewHandler
}

func NewView(method, path string, handler ViewHandler) *View {
	return &View{
		Method:  method,
		Path:    path,
		Handler: handler,
	}
}

type ViewHandler func(*Context) *utils.Response

type Context struct {
	*gin.Context
	User   *types.User
	Models *model.Models
}

// CreateAudit 审计操作入库
func (c *Context) CreateAudit(ao *types.AuditOperate) {
	ao.Operator = c.User.Name
	ao.Ip = c.ClientIP()
	ao.CreateTime = time.Now()
	jsonBytes, err := json.Marshal(ao.OperateDataInterface)
	if err != nil {
		klog.Warningf("marshal operate data error: %s, data=%v", err.Error(), ao.OperateDataInterface)
	} else {
		ao.OperateData = jsonBytes
	}
	if err = c.Models.AuditOperateManager.Create(ao); err != nil {
		klog.Warningf("create audit operation error: %s, audit=%v", err.Error(), *ao)
	}
}

func (c *Context) GenerateResponse(err error, data any) *utils.Response {
	if err == nil {
		return &utils.Response{Code: code.Success, Data: data}
	}
	switch e := err.(type) {
	case *errors.Error:
		return &utils.Response{Code: e.Code(), Msg: e.Error()}
	default:
		return &utils.Response{Code: code.UnknownError, Msg: e.Error()}
	}
}

func (c *Context) GenerateResponseError(err error) *utils.Response {
	return c.GenerateResponse(err, nil)
}

func (c *Context) GenerateResponseOK(data any) *utils.Response {
	return c.GenerateResponse(nil, data)
}
