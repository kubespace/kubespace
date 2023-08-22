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

type ImageRegistry struct {
	Views  []*views.View
	models *model.Models
}

func NewImageRegistry(models *model.Models) *ImageRegistry {
	settings := &ImageRegistry{
		models: models,
	}
	settings.Views = []*views.View{
		views.NewView(http.MethodGet, "", settings.list),
		views.NewView(http.MethodPost, "", settings.create),
		views.NewView(http.MethodPut, "/:id", settings.update),
		views.NewView(http.MethodDelete, "/:id", settings.delete),
	}
	return settings
}

func (s *ImageRegistry) create(c *views.Context) *utils.Response {
	var ser serializers.ImageRegistrySerializers
	if err := c.ShouldBind(&ser); err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	registry := &types.SettingsImageRegistry{
		Registry:   ser.Registry,
		User:       ser.User,
		Password:   ser.Password,
		CreateUser: c.User.Name,
		UpdateUser: c.User.Name,
		CreateTime: time.Time{},
		UpdateTime: time.Time{},
	}
	_, err := s.models.ImageRegistryManager.Create(registry)
	if err != nil {
		err = errors.New(code.DBError, "创建镜像仓库失败: "+err.Error())
	}
	resp := c.GenerateResponse(err, nil)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationCreate,
		OperateDetail:        fmt.Sprintf("创建镜像仓库：%s", ser.Registry),
		Scope:                types.ScopePlatform,
		ResourceId:           registry.ID,
		ResourceType:         types.AuditResourcePlatformRegistry,
		ResourceName:         ser.Registry,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: ser,
	})
	return resp
}

func (s *ImageRegistry) update(c *views.Context) *utils.Response {
	var ser serializers.ImageRegistrySerializers
	if err := c.ShouldBind(&ser); err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	obj, err := s.models.ImageRegistryManager.Get(uint(id))
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DataNotExists, "获取镜像仓库失败: "+err.Error()))
	}
	obj.User = ser.User
	obj.Password = ser.Password
	obj.UpdateUser = c.User.Name
	obj.UpdateTime = time.Now()
	_, err = s.models.ImageRegistryManager.Update(obj)
	if err != nil {
		err = errors.New(code.DBError, "更新镜像仓库失败: "+err.Error())
	}
	resp := c.GenerateResponse(err, nil)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationUpdate,
		OperateDetail:        fmt.Sprintf("更新镜像仓库：%s", ser.Registry),
		Scope:                types.ScopePlatform,
		ResourceId:           obj.ID,
		ResourceType:         types.AuditResourcePlatformRegistry,
		ResourceName:         ser.Registry,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: ser,
	})
	return resp
}

func (s *ImageRegistry) delete(c *views.Context) *utils.Response {
	secretId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.GenerateResponseError(errors.New(code.ParamsError, err))
	}
	obj, err := s.models.ImageRegistryManager.Get(uint(secretId))
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DataNotExists, "获取镜像仓库失败: "+err.Error()))
	}
	err = s.models.ImageRegistryManager.Delete(obj)
	if err != nil {
		err = errors.New(code.DBError, "删除镜像仓库失败: "+err.Error())
	}
	resp := c.GenerateResponse(err, nil)
	c.CreateAudit(&types.AuditOperate{
		Operation:            types.AuditOperationDelete,
		OperateDetail:        fmt.Sprintf("删除镜像仓库：%s", obj.Registry),
		Scope:                types.ScopePlatform,
		ResourceId:           obj.ID,
		ResourceType:         types.AuditResourcePlatformRegistry,
		ResourceName:         obj.Registry,
		Code:                 resp.Code,
		Message:              resp.Msg,
		OperateDataInterface: nil,
	})
	return resp
}

func (s *ImageRegistry) list(c *views.Context) *utils.Response {
	objs, err := s.models.ImageRegistryManager.List()
	if err != nil {
		return c.GenerateResponseError(errors.New(code.DBError, err))
	}
	var data []map[string]interface{}

	for _, obj := range objs {
		data = append(data, map[string]interface{}{
			"id":          obj.ID,
			"registry":    obj.Registry,
			"user":        obj.User,
			"create_user": obj.CreateUser,
			"update_user": obj.UpdateUser,
			"create_time": obj.CreateTime,
			"update_time": obj.UpdateTime,
		})
	}
	return c.GenerateResponseOK(data)
}
