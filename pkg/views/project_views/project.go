package project_views

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"net/http"
	"strconv"
	"time"
)

type Project struct {
	Views            []*views.View
	models           *model.Models
}

func NewProject(models *model.Models) *Project {
	pipelineWs := &Project{
		models:           models,
	}
	vs := []*views.View{
		views.NewView(http.MethodPost, "", pipelineWs.create),
	}
	pipelineWs.Views = vs
	return pipelineWs
}

func (p *Project) create(c *views.Context) *utils.Response {
	var ser serializers.ProjectSerializer
	resp := &utils.Response{Code: code.Success}
	if err := c.ShouldBind(&ser); err != nil {
		resp.Code = code.ParamsError
		resp.Msg = err.Error()
		return resp
	}
	project := &types.Project{
		Name: ser.Name,
		Description: ser.Description,
		ClusterId: ser.ClusterId,
		Namespace: ser.Namespace,
		Owner: ser.Owner,
		CreateUser: c.User.Name,
		UpdateUser: c.User.Name,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	project, err := p.models.ProjectManager.Create(project)
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = fmt.Sprintf("创建项目空间失败:%s", err.Error())
		return resp
	}
	resp.Data = project
	return resp
}

func (p *Project) list(c *views.Context) *utils.Response {
	resp := &utils.Response{Code: code.Success}
	projects, err := p.models.ProjectManager.List()
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = err.Error()
		return resp
	}
	var data []map[string]interface{}

	for _, project := range projects {
		data = append(data, map[string]interface{}{
			"id": project.ID,
			"name": project.Name,
			"description": project.Description,
			"cluster_id": project.ClusterId,
			"namespace": project.Namespace,
			"owner": project.Owner,
			"update_user": project.UpdateUser,
			"create_time": project.CreateTime,
			"update_time": project.UpdateTime,
		})
	}
	resp.Data = data
	return resp
}

func (p *Project) delete(c *views.Context) *utils.Response {
	resp := &utils.Response{Code: code.Success}
	projectId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	project, err := p.models.ProjectManager.Get(uint(projectId))
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "获取项目空间失败: " + err.Error()
		return resp
	}
	err = p.models.ProjectManager.Delete(project)
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "删除项目空间失败: " + err.Error()
		return resp
	}
	return resp
}
