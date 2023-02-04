package pipeline

import (
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"net/http"
	"strconv"
	"time"
)

type PipelineResource struct {
	Views  []*views.View
	models *model.Models
}

func NewPipelineResource(config *config.ServerConfig) *PipelineResource {
	pipelineWs := &PipelineResource{
		models: config.Models,
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "/:workspaceId", pipelineWs.list),
		views.NewView(http.MethodPost, "", pipelineWs.create),
		views.NewView(http.MethodPut, "/:id", pipelineWs.update),
		views.NewView(http.MethodDelete, "/:id", pipelineWs.delete),
	}
	pipelineWs.Views = vs
	return pipelineWs
}

func (r *PipelineResource) create(c *views.Context) *utils.Response {
	var ser serializers.PipelineResourceSerializer
	resp := &utils.Response{Code: code.Success}
	if err := c.ShouldBind(&ser); err != nil {
		resp.Code = code.ParamsError
		resp.Msg = err.Error()
		return resp
	}
	resource := &types.PipelineResource{
		WorkspaceId: ser.WorkspaceId,
		Name:        ser.Name,
		Type:        ser.Type,
		Value:       ser.Value,
		Global:      ser.Global,
		SecretId:    ser.SecretId,
		Description: ser.Description,
		CreateUser:  c.User.Name,
		UpdateUser:  c.User.Name,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
	_, err := r.models.PipelineResourceManager.Create(resource)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: "创建资源失败: " + err.Error()}
	}
	return &utils.Response{Code: code.Success}
}

func (r *PipelineResource) update(c *views.Context) *utils.Response {
	var ser serializers.PipelineResourceSerializer
	resp := &utils.Response{Code: code.Success}
	if err := c.ShouldBind(&ser); err != nil {
		resp.Code = code.ParamsError
		resp.Msg = err.Error()
		return resp
	}
	resId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	resource, err := r.models.PipelineResourceManager.Get(uint(resId))
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "获取资源失败: " + err.Error()
		return resp
	}
	resource.Value = ser.Value
	resource.Description = ser.Description
	resource.Global = ser.Global
	resource.SecretId = ser.SecretId
	resource.UpdateUser = c.User.Name
	resource.UpdateTime = time.Now()
	_, err = r.models.PipelineResourceManager.Update(resource)
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "更新资源失败: " + err.Error()
		return resp
	}
	return resp
}

func (r *PipelineResource) list(c *views.Context) *utils.Response {
	workspaceId, err := strconv.ParseUint(c.Param("workspaceId"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	resp := &utils.Response{Code: code.Success}
	resources, err := r.models.PipelineResourceManager.List(uint(workspaceId))
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = err.Error()
		return resp
	}
	resp.Data = resources
	return resp
}

func (r *PipelineResource) delete(c *views.Context) *utils.Response {
	resp := &utils.Response{Code: code.Success}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	res, err := r.models.PipelineResourceManager.Get(uint(id))
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "获取资源失败: " + err.Error()
		return resp
	}
	err = r.models.PipelineResourceManager.Delete(res)
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "删除资源失败: " + err.Error()
		return resp
	}
	return resp
}
