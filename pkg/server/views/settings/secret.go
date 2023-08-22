package settings

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

type SettingsSecret struct {
	Views  []*views.View
	models *model.Models
}

func NewSettingsSecret(models *model.Models) *SettingsSecret {
	secret := &SettingsSecret{
		models: models,
	}
	secret.Views = []*views.View{
		views.NewView(http.MethodGet, "", secret.list),
		views.NewView(http.MethodPost, "", secret.create),
		views.NewView(http.MethodPut, "/:id", secret.update),
		views.NewView(http.MethodDelete, "/:id", secret.delete),
	}
	return secret
}

func (s *SettingsSecret) create(c *views.Context) *utils.Response {
	var ser serializers.SecretsSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	secret := &types.SettingsSecret{
		Name:        ser.Name,
		Description: ser.Description,
		Type:        ser.SecretType,
		User:        ser.User,
		Password:    ser.Password,
		PrivateKey:  ser.PrivateKey,
		AccessToken: ser.AccessToken,
		CreateUser:  c.User.Name,
		UpdateUser:  c.User.Name,
		CreateTime:  time.Time{},
		UpdateTime:  time.Time{},
	}
	_, err := s.models.SettingsSecretManager.Create(secret)
	if err != nil {
		err = errors.New(code.DBError, err)
	}
	resp := c.GenerateResponse(err, nil)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationCreate,
		OperateDetail:        fmt.Sprintf("创建密钥：%s", ser.Name),
		Scope:                types.ScopePlatform,
		ResourceId:           secret.ID,
		ResourceType:         types.AuditResourcePlatformSecret,
		ResourceName:         ser.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: ser,
	})
	return resp
}

func (s *SettingsSecret) update(c *views.Context) *utils.Response {
	var ser serializers.SecretsSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	secretId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	secret, err := s.models.SettingsSecretManager.Get(uint(secretId))
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DataNotExists, "获取密钥失败: "+err.Error()))
	}
	secret.Description = ser.Description
	secret.Type = ser.SecretType
	secret.User = ser.User
	secret.Password = ser.Password
	secret.PrivateKey = ser.PrivateKey
	secret.AccessToken = ser.AccessToken
	secret.UpdateUser = c.User.Name
	secret.UpdateTime = time.Now()
	_, err = s.models.SettingsSecretManager.Update(secret)
	if err != nil {
		err = errors.New(code.DBError, "更新密钥失败: "+err.Error())
	}
	resp := c.GenerateResponse(err, nil)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationUpdate,
		OperateDetail:        fmt.Sprintf("更新密钥：%s", secret.Name),
		Scope:                types.ScopePlatform,
		ResourceId:           secret.ID,
		ResourceType:         types.AuditResourcePlatformSecret,
		ResourceName:         secret.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: ser,
	})
	return resp
}

func (s *SettingsSecret) delete(c *views.Context) *utils.Response {
	secretId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	secret, err := s.models.SettingsSecretManager.Get(uint(secretId))
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DataNotExists, "获取密钥失败: "+err.Error()))
	}
	err = s.models.SettingsSecretManager.Delete(secret)
	if err != nil {
		err = errors.New(code.DBError, "删除密钥失败: "+err.Error())
	}
	resp := c.GenerateResponse(err, nil)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationDelete,
		OperateDetail:        fmt.Sprintf("删除密钥：%s", secret.Name),
		Scope:                types.ScopePlatform,
		ResourceId:           secret.ID,
		ResourceType:         types.AuditResourcePlatformSecret,
		ResourceName:         secret.Name,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: nil,
	})
	return resp
}

func (s *SettingsSecret) list(c *views.Context) *utils.Response {
	resp := &utils.Response{Code: code.Success}
	secrets, err := s.models.SettingsSecretManager.List()
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = err.Error()
		return resp
	}
	var data []map[string]interface{}

	for _, secret := range secrets {
		data = append(data, map[string]interface{}{
			"id":          secret.ID,
			"name":        secret.Name,
			"description": secret.Description,
			"type":        secret.Type,
			"user":        secret.User,
			"create_user": secret.CreateUser,
			"update_user": secret.UpdateUser,
			"create_time": secret.CreateTime,
			"update_time": secret.UpdateTime,
		})
	}
	resp.Data = data
	return resp
}
