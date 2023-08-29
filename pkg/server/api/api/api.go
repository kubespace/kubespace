package api

import (
	"github.com/kubespace/kubespace/pkg/utils"
)

// ApiGroup 一个api分组下的所有api
type ApiGroup interface {
	Apis() []*Api
}

// Api 单个api
type Api struct {
	Method  string
	Path    string
	Handler Handler
}

func NewApi(method, path string, handler Handler) *Api {
	return &Api{
		Method:  method,
		Path:    path,
		Handler: handler,
	}
}

type Handler interface {
	// Auth 当前api是否需要认证，以及鉴权所需要的权限
	Auth(c *Context) (bool, *AuthPerm, error)
	// Handle api处理逻辑
	Handle(c *Context) *utils.Response
}

//type Handler func(*Context) *utils.Response
