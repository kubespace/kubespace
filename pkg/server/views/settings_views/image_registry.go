package settings_views

import (
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
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
	vs := []*views.View{
		views.NewView(http.MethodGet, "", settings.list),
		views.NewView(http.MethodPost, "", settings.create),
		views.NewView(http.MethodPut, "/:id", settings.update),
		views.NewView(http.MethodDelete, "/:id", settings.delete),
	}
	settings.Views = vs
	return settings
}

func (s *ImageRegistry) create(c *views.Context) *utils.Response {
	var ser serializers.ImageRegistrySerializers
	resp := &utils.Response{Code: code.Success}
	if err := c.ShouldBind(&ser); err != nil {
		resp.Code = code.ParamsError
		resp.Msg = err.Error()
		return resp
	}
	secret := &types.SettingsImageRegistry{
		Registry:   ser.Registry,
		User:       ser.User,
		Password:   ser.Password,
		CreateUser: c.User.Name,
		UpdateUser: c.User.Name,
		CreateTime: time.Time{},
		UpdateTime: time.Time{},
	}
	_, err := s.models.ImageRegistryManager.Create(secret)
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "创建镜像仓库失败: " + err.Error()
		return resp
	}
	return resp
}

func (s *ImageRegistry) update(c *views.Context) *utils.Response {
	var ser serializers.ImageRegistrySerializers
	resp := &utils.Response{Code: code.Success}
	if err := c.ShouldBind(&ser); err != nil {
		resp.Code = code.ParamsError
		resp.Msg = err.Error()
		return resp
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	obj, err := s.models.ImageRegistryManager.Get(uint(id))
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "获取镜像仓库失败: " + err.Error()
		return resp
	}
	obj.User = ser.User
	obj.Password = ser.Password
	obj.UpdateUser = c.User.Name
	obj.UpdateTime = time.Now()
	_, err = s.models.ImageRegistryManager.Update(obj)
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "更新镜像仓库失败: " + err.Error()
		return resp
	}
	return resp
}

func (s *ImageRegistry) delete(c *views.Context) *utils.Response {
	resp := &utils.Response{Code: code.Success}
	secretId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	obj, err := s.models.ImageRegistryManager.Get(uint(secretId))
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "获取镜像仓库失败: " + err.Error()
		return resp
	}
	err = s.models.ImageRegistryManager.Delete(obj)
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "删除镜像参考失败: " + err.Error()
		return resp
	}
	return resp
}

func (s *ImageRegistry) list(c *views.Context) *utils.Response {
	resp := &utils.Response{Code: code.Success}
	objs, err := s.models.ImageRegistryManager.List()
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = err.Error()
		return resp
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
	resp.Data = data
	return resp
}
