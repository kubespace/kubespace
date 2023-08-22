package user

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/utils"
	"net/http"
	"strconv"
	"time"
)

type User struct {
	Views  []*views.View
	models *model.Models
}

func NewUser(models *model.Models) *User {
	user := &User{
		models: models,
	}
	user.Views = []*views.View{
		views.NewView(http.MethodGet, "", user.list),
		views.NewView(http.MethodGet, "/:id/roles", user.list),
		views.NewView(http.MethodPost, "", user.create),
		//NewView(http.MethodPost, "/admin", user.create),
		views.NewView(http.MethodPut, "/", user.updateSelf),
		views.NewView(http.MethodPut, "/:username", user.update),

		views.NewView(http.MethodGet, "/token", user.tokenUser),
		views.NewView(http.MethodPost, "/delete", user.delete),
		views.NewView(http.MethodPost, "/update_password", user.updatePassword),
	}
	return user
}

func (u *User) tokenUser(c *views.Context) *utils.Response {
	user, err := u.models.UserManager.GetById(c.User.ID)
	if err != nil {
		return &utils.Response{
			Code: code.ParamsError,
			Msg:  err.Error(),
		}
	}
	perms, err := u.models.UserRoleManager.GetUserRoles(user.ID)
	if err != nil {
		return &utils.Response{
			Code: code.ParamsError,
			Msg:  err.Error(),
		}
	}
	return &utils.Response{Code: code.Success,
		Data: map[string]interface{}{
			"id":          c.User.ID,
			"name":        c.User.Name,
			"permissions": perms,
			"is_super":    user.IsSuper,
		}}
}

func (u *User) update(c *views.Context) *utils.Response {
	userName := c.Param("username")
	var user serializers.UserSerializers

	if err := c.ShouldBindJSON(&user); err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}

	userObj, err := u.models.UserManager.GetByName(userName)
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DataNotExists, err))
	}

	if user.Status != "" {
		userObj.Status = user.Status
	}

	if user.Password != "" {
		userObj.Password = utils.Encrypt(user.Password)
	}

	if user.Email != "" {
		if ok := utils.VerifyEmailFormat(user.Email); !ok {
			return c.GenerateResponseError(errors.New(code.ParamsError, fmt.Sprintf("email:%s format error for user:%s", user.Email, userName)))
		}
		userObj.Email = user.Email
	}

	err = u.models.UserManager.Update(userObj)
	if err != nil {
		err = errors.New(code.UpdateError, err)
	}
	resp := c.GenerateResponse(err, nil)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationUpdate,
		OperateDetail:        fmt.Sprintf("更新用户：%s", userObj.Name),
		Scope:                types.ScopePlatform,
		ResourceId:           userObj.ID,
		ResourceType:         types.AuditResourcePlatformUser,
		ResourceName:         userObj.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: user,
	})
	return resp
}

func (u *User) updateSelf(c *views.Context) *utils.Response {
	var user serializers.UserSerializers

	resp := &utils.Response{Code: code.Success}

	if err := c.ShouldBindJSON(&user); err != nil {
		resp.Code = code.ParamsError
		resp.Msg = err.Error()
		return resp
	}

	userObj, err := u.models.UserManager.GetById(c.User.ID)
	if err != nil {
		resp.Code = code.GetError
		resp.Msg = err.Error()
		return resp
	}

	if user.Status != "" {
		userObj.Status = user.Status
	}

	if user.Password != "" {
		userObj.Password = utils.Encrypt(user.Password)
	}

	//if user.Roles != nil {
	//	userObj.Roles = user.Roles
	//}

	if user.Email != "" {
		if ok := utils.VerifyEmailFormat(user.Email); !ok {
			resp.Code = code.ParamsError
			resp.Msg = fmt.Sprintf("email:%s format error for user:%s", user.Email, userObj.Name)
			return resp
		}
		userObj.Email = user.Email
	}

	if err := u.models.UserManager.Update(userObj); err != nil {
		resp.Code = code.UpdateError
		resp.Msg = err.Error()
		return resp
	}
	return resp
}

func (u *User) list(c *views.Context) *utils.Response {
	resp := &utils.Response{Code: code.Success}
	var filters map[string]interface{}

	dList, err := u.models.UserManager.List(filters)
	if err != nil {
		resp.Code = code.GetError
		resp.Msg = err.Error()
		return resp
	}
	var data []map[string]interface{}

	for _, du := range dList {
		perms, err := u.models.UserManager.Permissions(&du)
		if err != nil {
			resp.Code = code.GetError
			resp.Msg = err.Error()
			return resp
		}
		data = append(data, map[string]interface{}{
			"id":         du.ID,
			"name":       du.Name,
			"email":      du.Email,
			"status":     du.Status,
			"is_super":   du.IsSuper,
			"last_login": du.LastLogin,
			//"roles":       du.Roles,
			"permissions": perms,
		})
	}
	resp.Data = data
	return resp
}

func (u *User) create(c *views.Context) *utils.Response {
	var ser serializers.UserCreateSerializers

	if err := c.ShouldBind(&ser); err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	isSuper := false
	if ser.Name == "" {
		ser.Name = types.ADMIN
		isSuper = true
	} else if ok := utils.VerifyEmailFormat(ser.Email); !ok {
		return c.GenerateResponseError(errors.New(code.ParamsError, fmt.Sprintf("email:%s format error for user:%s", ser.Email, ser.Name)))
	}

	userObj := types.User{
		Name:       ser.Name,
		Password:   utils.Encrypt(ser.Password),
		Email:      ser.Email,
		IsSuper:    isSuper,
		Status:     "normal",
		LastLogin:  time.Now(),
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	err := u.models.UserManager.Create(&userObj)
	if err != nil {
		err = errors.New(code.CreateError, err)
	}
	resp := c.GenerateResponse(err, nil)

	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationCreate,
		OperateDetail:        fmt.Sprintf("创建用户：%s", ser.Name),
		Scope:                types.ScopePlatform,
		ResourceId:           userObj.ID,
		ResourceType:         types.AuditResourcePlatformUser,
		ResourceName:         ser.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: ser,
	})
	return resp
}

func (u *User) delete(c *views.Context) *utils.Response {
	var ser []serializers.DeleteUserSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	for _, du := range ser {
		user, err := u.models.UserManager.GetByName(du.Name)
		if err != nil {
			return c.GenerateResponseError(errors.New(code.DataNotExists, err))
		}
		err = u.models.UserManager.Delete(du.Name)
		if err != nil {
			err = errors.New(code.DeleteError, err)
		}
		resp := c.GenerateResponse(err, nil)
		c.CreateAudit(&types.AuditOperate{
			Operation:            types.AuditOperationDelete,
			OperateDetail:        fmt.Sprintf("删除用户：%s", user.Name),
			Scope:                types.ScopePlatform,
			ResourceId:           user.ID,
			ResourceType:         types.AuditResourcePlatformUser,
			ResourceName:         user.Name,
			Code:                 resp.Code,
			Message:              resp.Msg,
			OperateDataInterface: user,
		})
		if err != nil {
			return c.GenerateResponseError(err)
		}
	}
	return &utils.Response{Code: code.Success}
}

func (u *User) roles(c *views.Context) *utils.Response {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	roles, err := u.models.UserRoleManager.GetUserRoles(uint(id))
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success, Data: roles}
}

func (u *User) updatePassword(c *views.Context) *utils.Response {
	var ser serializers.UpdatePasswordSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	userObj, err := u.models.UserManager.GetById(c.User.ID)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if userObj.Password != utils.Encrypt(ser.OriginPassword) {
		return &utils.Response{Code: code.ParamsError, Msg: "原密码不正确，请重新输入"}
	}
	userObj.Password = utils.Encrypt(ser.NewPassword)

	if err = u.models.UserManager.Update(userObj); err != nil {
		return &utils.Response{Code: code.UpdateError, Msg: "更新密码失败：" + err.Error()}
	}
	return &utils.Response{Code: code.Success}
}
