package views

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"k8s.io/klog"
	"net/http"
	"time"
)

type User struct {
	Views  []*View
	models *model.Models
}

func NewUser(models *model.Models) *User {
	user := &User{
		models: models,
	}
	views := []*View{
		NewView(http.MethodGet, "", user.list),
		NewView(http.MethodPost, "", user.create),
		//NewView(http.MethodPost, "/admin", user.create),
		NewView(http.MethodPut, "/:username", user.update),

		NewView(http.MethodGet, "/token", user.tokenUser),
		NewView(http.MethodPost, "/delete", user.delete),
	}
	user.Views = views
	return user
}

func (u *User) tokenUser(c *Context) *utils.Response {
	user, err := u.models.UserManager.Get(c.User.Name)
	if err != nil {
		return &utils.Response{
			Code: code.ParamsError,
			Msg:  err.Error(),
		}
	}
	perms, err := u.models.UserManager.Permissions(user)
	if err != nil {
		return &utils.Response{
			Code: code.ParamsError,
			Msg:  err.Error(),
		}
	}
	return &utils.Response{Code: code.Success,
		Data: map[string]interface{}{
			"name":        c.User.Name,
			"permissions": perms,
			"is_super":    user.IsSuper,
		}}
}

func (u *User) update(c *Context) *utils.Response {
	userName := c.Param("username")
	var user serializers.UserSerializers

	resp := &utils.Response{Code: code.Success}

	if err := c.ShouldBindJSON(&user); err != nil {
		resp.Code = code.ParamsError
		resp.Msg = err.Error()
		return resp
	}

	userObj, err := u.models.UserManager.Get(userName)
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
			resp.Msg = fmt.Sprintf("email:%s format error for user:%s", user.Email, userName)
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

func (u *User) list(c *Context) *utils.Response {
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
			"name":        du.Name,
			"email":       du.Email,
			"status":      du.Status,
			"is_super":    du.IsSuper,
			"last_login":  du.LastLogin,
			//"roles":       du.Roles,
			"permissions": perms,
		})
	}
	resp.Data = data
	return resp
}

func (u *User) create(c *Context) *utils.Response {
	var ser serializers.UserCreateSerializers
	resp := &utils.Response{Code: code.Success}

	if err := c.ShouldBind(&ser); err != nil {
		resp.Code = code.ParamsError
		resp.Msg = err.Error()
		return resp
	}
	isSuper := false
	if ser.Name == "" {
		ser.Name = types.ADMIN
		isSuper = true
	} else {
		if ok := utils.VerifyEmailFormat(ser.Email); !ok {
			resp.Code = code.ParamsError
			resp.Msg = fmt.Sprintf("email:%s format error for user:%s", ser.Email, ser.Name)
			return resp
		}
	}

	userObj := types.User{
		Name:     ser.Name,
		Password: utils.Encrypt(ser.Password),
		Email:    ser.Email,
		IsSuper:  isSuper,
		Status:   "normal",
		//Roles:    ser.Roles,
		LastLogin: time.Now(),
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	if err := u.models.UserManager.Create(&userObj); err != nil {
		resp.Code = code.CreateError
		resp.Msg = err.Error()
		return resp
	}

	resp.Data = map[string]interface{}{
		"name":     userObj.Name,
		"password": userObj.Password,
		"status":   userObj.Status,
	}
	return resp
}

func (u *User) delete(c *Context) *utils.Response {
	var ser []serializers.DeleteUserSerializers
	if err := c.ShouldBind(&ser); err != nil {
		klog.Errorf("bind params error: %s", err.Error())
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	klog.Info(ser)
	for _, du := range ser {
		err := u.models.UserManager.Delete(du.Name)
		if err != nil {
			klog.Errorf("delete user %s error: %s", c, err.Error())
			return &utils.Response{Code: code.DeleteError, Msg: err.Error()}
		}
	}
	return &utils.Response{Code: code.Success}
}
