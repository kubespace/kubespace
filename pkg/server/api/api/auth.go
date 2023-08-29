package api

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/config"
)

const (
	SessionId = "sessionId"
)

// AuthPerm 鉴权所需要的权限
type AuthPerm struct {
	Scope   string
	ScopeId uint
	Role    string
}

type Auth struct {
	models *model.Models
}

func NewAuth(conf *config.ServerConfig) *Auth {
	return &Auth{models: conf.Models}
}

// Authenticate 认证，
// a. 通过session获取用户信息
// b. todo: 通过用户token认证
func (a *Auth) Authenticate(c *Context) (*types.User, error) {
	sessionId, err := c.Cookie(SessionId)
	if err != nil {
		return nil, errors.New(code.CookieError, fmt.Sprintf("get auth cookie session error: %v", err))
	}
	session, err := a.models.SessionManager.Get(sessionId)
	if err != nil {
		return nil, errors.New(code.DataNotExists, fmt.Sprintf("get auth session error: %v", err))
	}

	user, err := a.models.UserManager.GetByName(session.UserName)
	if err != nil {
		return nil, errors.New(code.DataNotExists, fmt.Sprintf("get auth user error: %v", err))
	}
	return user, nil
}

// Authorize 鉴权，用户是否有该perm权限
func (a *Auth) Authorize(c *Context, perm *AuthPerm) (bool, error) {
	ok := a.models.UserRoleManager.AuthRole(c.User, perm.Scope, perm.ScopeId, perm.Role)
	return ok, nil
}
