package project_views

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	projectservice "github.com/kubespace/kubespace/pkg/project"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"k8s.io/klog"
	"net/http"
	"strconv"
	"time"
)

type Project struct {
	Views          []*views.View
	models         *model.Models
	projectService *projectservice.ServiceProject
}

func NewProject(models *model.Models, projectService *projectservice.ServiceProject) *Project {
	pipelineWs := &Project{
		models:         models,
		projectService: projectService,
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "", pipelineWs.list),
		views.NewView(http.MethodPost, "", pipelineWs.create),
		views.NewView(http.MethodDelete, "/:id", pipelineWs.delete),
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
		Name:        ser.Name,
		Description: ser.Description,
		ClusterId:   ser.ClusterId,
		Namespace:   ser.Namespace,
		Owner:       ser.Owner,
		CreateUser:  c.User.Name,
		UpdateUser:  c.User.Name,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
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
	clusters := make(map[string]*types.Cluster)

	for _, project := range projects {
		cluster, ok := clusters[project.ClusterId]
		if !ok {
			cluster, err = p.models.ClusterManager.GetByName(project.ClusterId)
			if err != nil {
				klog.Errorf("get project id=%s cluster error: %s", project.ID, err.Error())
			}
			clusters[project.ClusterId] = cluster
		}
		data = append(data, map[string]interface{}{
			"id":          project.ID,
			"name":        project.Name,
			"description": project.Description,
			"cluster_id":  project.ClusterId,
			"cluster":     cluster,
			"namespace":   project.Namespace,
			"owner":       project.Owner,
			"update_user": project.UpdateUser,
			"create_time": project.CreateTime,
			"update_time": project.UpdateTime,
		})
	}
	resp.Data = data
	return resp
}

func (p *Project) delete(c *views.Context) *utils.Response {
	projectId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return p.projectService.Delete(uint(projectId))
}

func (p *Project) get(c *views.Context) *utils.Response {
	projectId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return p.projectService.Get(uint(projectId))
}
