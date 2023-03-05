package settings

import (
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"net/http"
	"strconv"
)

type SettingsLdap struct {
	Views  []*views.View
	models *model.Models
}

func NewSettingsLdap(models *model.Models) *SettingsLdap {
	ldap := &SettingsLdap{
		models: models,
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "", ldap.list),
		views.NewView(http.MethodPost, "", ldap.create),
		views.NewView(http.MethodPut, "/:id", ldap.update),
		views.NewView(http.MethodDelete, "/:id", ldap.delete),
	}
	ldap.Views = vs
	return ldap
}

func (s *SettingsLdap) create(c *views.Context) *utils.Response {
	var ldp serializers.LdapSerializers
	resp := &utils.Response{Code: code.Success}
	if err := c.ShouldBind(&ldp); err != nil {
		resp.Code = code.ParamsError
		resp.Msg = err.Error()
		return resp
	}
	ldap := &types.Ldap{
		Enable:      ldp.Enable,
		Name:        ldp.Name,
		Url:         ldp.Url,
		MaxConn:     10,
		BaseDN:      ldp.BaseDN,
		AdminDN:     ldp.AdminDN,
		AdminDNPass: ldp.AdminDNPass,
	}
	_, err := s.models.LdapManager.Create(ldap)
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "创建ldap服务失败: " + err.Error()
		return resp
	}
	return resp
}

func (s *SettingsLdap) update(c *views.Context) *utils.Response {
	var ldp serializers.LdapSerializers
	resp := &utils.Response{Code: code.Success}
	if err := c.ShouldBind(&ldp); err != nil {
		resp.Code = code.ParamsError
		resp.Msg = err.Error()
		return resp
	}
	ldapId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	ldap, err := s.models.LdapManager.Get(uint(ldapId))
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "获取Ldap失败: " + err.Error()
		return resp
	}
	ldap.Enable = ldp.Enable
	ldap.Name = ldp.Name
	ldap.Url = ldp.Url
	ldap.MaxConn = 10
	ldap.BaseDN = ldp.BaseDN
	ldap.AdminDN = ldp.AdminDN
	ldap.AdminDNPass = ldp.AdminDNPass
	_, err = s.models.LdapManager.Update(ldap)
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "更新Ldap失败: " + err.Error()
		return resp
	}
	return resp
}

func (s *SettingsLdap) delete(c *views.Context) *utils.Response {
	resp := &utils.Response{Code: code.Success}
	ldapId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	ldap, err := s.models.LdapManager.Get(uint(ldapId))
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "获取ldap失败: " + err.Error()
		return resp
	}
	err = s.models.LdapManager.Delete(ldap)
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "删除ldap失败: " + err.Error()
		return resp
	}
	return resp
}

func (s *SettingsLdap) list(c *views.Context) *utils.Response {
	resp := &utils.Response{Code: code.Success}
	ldaps, err := s.models.LdapManager.List()
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = err.Error()
		return resp
	}
	var data []map[string]interface{}

	for _, ldap := range ldaps {
		data = append(data, map[string]interface{}{
			"id":          ldap.ID,
			"name":        ldap.Name,
			"enable":      ldap.Enable,
			"url":         ldap.Url,
			"baseDN":      ldap.BaseDN,
			"adminDN":     ldap.AdminDN,
			"adminDNPass": ldap.AdminDNPass,
		})
	}
	resp.Data = data
	return resp
}
