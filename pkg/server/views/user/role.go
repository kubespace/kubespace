package user

import (
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"k8s.io/klog/v2"
	"net/http"
)

type Role struct {
	Views  []*views.View
	models *model.Models
}

func NewRole(models *model.Models) *Role {
	role := &Role{
		models: models,
	}
	role.Views = []*views.View{
		views.NewView(http.MethodGet, "/permissions", role.permissions),
		views.NewView(http.MethodGet, "", role.list),
		views.NewView(http.MethodPost, "", role.create),
		views.NewView(http.MethodPut, "/:rolename", role.update),
		views.NewView(http.MethodPost, "/delete", role.delete),
	}
	role.models.RoleManager.Init()
	return role
}

func (r *Role) permissions(c *views.Context) *utils.Response {
	//userName := ""
	//if user, ok := c.GetByName("user"); ok {
	//	userName = user.(*types.User).Name
	//}
	return &utils.Response{Code: code.Success, Data: types.AllPermissions}
}

func (r *Role) list(c *views.Context) *utils.Response {
	resp := &utils.Response{Code: code.Success}
	var filters map[string]interface{}

	rList, err := r.models.RoleManager.List(filters)
	if err != nil {
		resp.Code = code.GetError
		resp.Msg = err.Error()
		return resp
	}

	resp.Data = rList
	return resp
}

func (r *Role) create(c *views.Context) *utils.Response {
	var role types.Role
	resp := &utils.Response{Code: code.Success}

	if err := c.ShouldBind(&role); err != nil {
		klog.Error("params error: ", err)
		resp.Code = code.ParamsError
		resp.Msg = err.Error()
		return resp
	}

	for i, r := range role.Permissions {
		addGet := false
		for _, op := range r.Operations {
			if op == types.OpGet {
				addGet = false
				break
			}
			if op == types.OpCreate || op == types.OpUpdate || op == types.OpDelete {
				addGet = true
			}
		}
		if addGet {
			r.Operations = append(r.Operations, types.OpGet)
		}
		role.Permissions[i].Operations = r.Operations
	}

	role.CreateTime = utils.StringNow()
	role.UpdateTime = utils.StringNow()

	if err := r.models.RoleManager.Create(&role); err != nil {
		resp.Code = code.CreateError
		resp.Msg = err.Error()
		return resp
	}

	resp.Data = role
	return resp
}

func (r *Role) update(c *views.Context) *utils.Response {
	roleName := c.Param("rolename")
	var role types.Role

	resp := &utils.Response{Code: code.Success}

	if err := c.ShouldBindJSON(&role); err != nil {
		resp.Code = code.ParamsError
		resp.Msg = err.Error()
		return resp
	}

	roleObj, err := r.models.RoleManager.Get(roleName)
	if err != nil {
		resp.Code = code.GetError
		resp.Msg = err.Error()
		return resp
	}

	if role.Description != "" {
		roleObj.Description = role.Description
	}

	for i, r := range role.Permissions {
		addGet := false
		for _, op := range r.Operations {
			if op == types.OpGet {
				addGet = false
				break
			}
			if op == types.OpCreate || op == types.OpUpdate || op == types.OpDelete {
				addGet = true
			}
		}
		if addGet {
			r.Operations = append(r.Operations, types.OpGet)
		}
		role.Permissions[i].Operations = r.Operations
	}
	roleObj.Permissions = role.Permissions

	role.UpdateTime = utils.StringNow()

	if err := r.models.RoleManager.Update(roleObj); err != nil {
		resp.Code = code.UpdateError
		resp.Msg = err.Error()
		return resp
	}
	return resp
}

func (r *Role) delete(c *views.Context) *utils.Response {
	var ser []serializers.DeleteRoleSerializers
	//klog.Info(c.Request.Body)
	if err := c.ShouldBind(&ser); err != nil {
		klog.Errorf("bind params error: %s", err.Error())
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	for _, du := range ser {
		err := r.models.RoleManager.Delete(du.Name)
		if err != nil {
			klog.Errorf("delete user %s error: %s", c, err.Error())
			return &utils.Response{Code: code.DeleteError, Msg: err.Error()}
		}
	}
	return &utils.Response{Code: code.Success}
}
