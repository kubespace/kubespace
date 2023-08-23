package user

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/manager/user"
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
	u := &User{
		models: models,
	}
	u.Views = []*views.View{
		views.NewView(http.MethodGet, "", u.list),
		views.NewView(http.MethodGet, "/:id/roles", u.list),
		views.NewView(http.MethodPost, "", u.create),
		views.NewView(http.MethodPut, "/", u.updateSelf),
		views.NewView(http.MethodPut, "/:username", u.update),

		views.NewView(http.MethodGet, "/token", u.tokenUser),
		views.NewView(http.MethodPost, "/delete", u.delete),
		views.NewView(http.MethodPost, "/update_password", u.updatePassword),
	}
	return u
}

func (u *User) tokenUser(c *views.Context) *utils.Response {
	userObj, err := u.models.UserManager.GetById(c.User.ID)
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DataNotExists, err))
	}
	perms, err := u.models.UserRoleManager.GetUserRoles(userObj.ID)
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DBError, err))
	}
	return c.GenerateResponseOK(map[string]interface{}{
		"id":          c.User.ID,
		"name":        c.User.Name,
		"permissions": perms,
		"is_super":    userObj.IsSuper,
	})
}

func (u *User) update(c *views.Context) *utils.Response {
	userName := c.Param("username")
	var ser serializers.UserSerializers

	if err := c.ShouldBindJSON(&ser); err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}

	userObj, err := u.models.UserManager.GetByName(userName)
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DataNotExists, err))
	}

	if ser.Status != "" {
		userObj.Status = ser.Status
	}

	if ser.Password != "" {
		userObj.Password = utils.Encrypt(ser.Password)
	}

	if ser.Email != "" {
		if ok := utils.VerifyEmailFormat(ser.Email); !ok {
			return c.GenerateResponseError(errors.New(code.ParamsError, fmt.Sprintf("email:%s format error for user:%s", ser.Email, userName)))
		}
		userObj.Email = ser.Email
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
		OperateDataInterface: ser,
	})
	return resp
}

func (u *User) updateSelf(c *views.Context) *utils.Response {
	var ser serializers.UserSerializers
	if err := c.ShouldBindJSON(&ser); err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}

	userObj, err := u.models.UserManager.GetById(c.User.ID)
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DataNotExists, err))
	}

	if ser.Status != "" {
		userObj.Status = ser.Status
	}
	if ser.Password != "" {
		userObj.Password = utils.Encrypt(ser.Password)
	}
	if ser.Email != "" {
		if ok := utils.VerifyEmailFormat(ser.Email); !ok {
			return c.GenerateResponseError(errors.New(code.ParamsError, fmt.Sprintf("email:%s format error for user:%s", ser.Email, userObj.Name)))
		}
		userObj.Email = ser.Email
	}

	if err := u.models.UserManager.Update(userObj); err != nil {
		return c.GenerateResponseError(errors.New(code.UpdateError, err))
	}
	return c.GenerateResponseOK(nil)
}

func (u *User) list(c *views.Context) *utils.Response {
	dList, err := u.models.UserManager.List(user.UserListCondition{})
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DBError, err))
	}
	var data []map[string]interface{}

	for _, du := range dList {
		data = append(data, map[string]interface{}{
			"id":         du.ID,
			"name":       du.Name,
			"email":      du.Email,
			"status":     du.Status,
			"is_super":   du.IsSuper,
			"last_login": du.LastLogin,
		})
	}
	return c.GenerateResponseOK(data)
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
		userObj, err := u.models.UserManager.GetByName(du.Name)
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
			OperateDetail:        fmt.Sprintf("删除用户：%s", userObj.Name),
			Scope:                types.ScopePlatform,
			ResourceId:           userObj.ID,
			ResourceType:         types.AuditResourcePlatformUser,
			ResourceName:         userObj.Name,
			Code:                 resp.Code,
			Message:              resp.Msg,
			OperateDataInterface: userObj,
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
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	roles, err := u.models.UserRoleManager.GetUserRoles(uint(id))
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DBError, err))
	}
	return c.GenerateResponseOK(roles)
}

func (u *User) updatePassword(c *views.Context) *utils.Response {
	var ser serializers.UpdatePasswordSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	userObj, err := u.models.UserManager.GetById(c.User.ID)
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DataNotExists, err))
	}
	if userObj.Password != utils.Encrypt(ser.OriginPassword) {
		return c.GenerateResponseError(errors.New(code.ParamsError, "原密码不正确，请重新输入"))
	}
	userObj.Password = utils.Encrypt(ser.NewPassword)

	if err = u.models.UserManager.Update(userObj); err != nil {
		return c.GenerateResponseError(errors.New(code.UpdateError, "更新密码失败："+err.Error()))
	}
	return c.GenerateResponseOK(nil)
}
