package settings

import (
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"io"
	"k8s.io/klog/v2"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type SettingsLdap struct {
	Views        []*views.View
	models       *model.Models
	progressChan chan int
	lock         sync.Mutex
	syncTime     uint64
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
		views.NewView(http.MethodGet, "/sync_progress/:id/:timeStamp", ldap.SyncLdap2Db),
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
		resp.Msg = "create ldap error: " + err.Error()
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
		resp.Msg = "get ldap error: " + err.Error()
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
		resp.Msg = "update ldap error: " + err.Error()
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
		resp.Msg = "get ldap error: " + err.Error()
		return resp
	}
	err = s.models.LdapManager.Delete(ldap)
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "delete ldap error: " + err.Error()
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

// SyncLdap2Db 同步ldap中的用户到本地数据库中，同时只允许一个人进行同步操作，有人同步时，建返回客户端错误事件
func (s *SettingsLdap) SyncLdap2Db(c *views.Context) *utils.Response {
	timeStamp, err := strconv.ParseUint(c.Param("timeStamp"), 10, 64)
	if err != nil {
		c.SSEvent("error", err.Error())
		return nil
	}
	defer func() {
		if s.syncTime == timeStamp {
			s.lock.Unlock()
		}
	}()
	if !s.lock.TryLock() {
		c.SSEvent("error", "Please wait for others to finish syncing")
		return nil
	}
	s.syncTime = timeStamp

	ldapId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.SSEvent("error", err.Error())
		return nil
	}

	ldap, err := s.models.LdapManager.Get(uint(ldapId))
	if err != nil {
		c.SSEvent("error", err.Error())
		return nil
	}

	err, result := utils.WithLDAPConn(&utils.LdapConfig{
		Url:      ldap.Url,
		User:     ldap.AdminDN,
		Password: ldap.AdminDNPass,
		BaseDN:   ldap.BaseDN,
	}, &utils.LdapConfig{
		BaseDN: ldap.BaseDN,
	}, utils.SearchLdapUsersFunc)

	if err != nil {
		c.SSEvent("error", err.Error())
		return nil
	}

	entries := result.([]map[string]string)
	s.progressChan = make(chan int)
	var count = 0
	var wg sync.WaitGroup
	wg.Add(1)

	// 开启同步ldap用户到本地数据协程
	go func() {
		for _, entry := range entries {
			userObj := types.User{
				Name:     entry["uid"],
				Password: utils.Encrypt(entry["uid"]),
				Email:    entry["email"],
				IsSuper:  false,
				Status:   "normal",
				//Roles:    ser.Roles,
				LastLogin:  time.Now(),
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
			}
			err := s.models.UserManager.Create(&userObj)
			if err != nil {
				klog.Error("ldap user:(%v),err:(%v) sync error", userObj, err)
			}
			count++
			progress := (float32(count) / float32(len(entries))) * float32(100)
			if count == len(entries) {
				s.progressChan <- 100
			} else {
				s.progressChan <- int(progress)
			}
			if count < 10 {
				time.Sleep(1 * time.Second)
			}
		}
		wg.Done()
	}()

	// 开启同步进度EventSource协程
	wg.Add(1)
	go func() {
		c.Stream(func(w io.Writer) bool {
			select {
			case p := <-s.progressChan:
				c.SSEvent("progress", p)
				if p >= 100 {
					return false
				}
			case <-c.Writer.CloseNotify():
				return false
			}
			return true
		})
		wg.Done()
	}()
	wg.Wait()

	defer close(s.progressChan)
	c.SSEvent("success", "success")
	return nil
}
