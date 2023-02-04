package project

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	projectservice "github.com/kubespace/kubespace/pkg/service/project"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"k8s.io/klog/v2"
	"net/http"
	"strconv"
	"time"
)

type Project struct {
	Views          []*views.View
	models         *model.Models
	projectService *projectservice.ProjectService
}

func NewProject(config *config.ServerConfig) *Project {
	projectWs := &Project{
		models:         config.Models,
		projectService: config.ServiceFactory.Project.ProjectService,
	}
	projectWs.Views = []*views.View{
		views.NewView(http.MethodGet, "", projectWs.list),
		views.NewView(http.MethodPost, "/resources", projectWs.getProjectResources),
		views.NewView(http.MethodGet, "/:id", projectWs.get),
		views.NewView(http.MethodPost, "", projectWs.create),
		views.NewView(http.MethodPost, "/clone", projectWs.clone),
		views.NewView(http.MethodPut, "/:id", projectWs.update),
		views.NewView(http.MethodDelete, "/:id", projectWs.delete),
	}
	return projectWs
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

func (p *Project) update(c *views.Context) *utils.Response {
	projectId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	var ser serializers.ProjectSerializer
	resp := &utils.Response{Code: code.Success}
	if err = c.ShouldBind(&ser); err != nil {
		resp.Code = code.ParamsError
		resp.Msg = err.Error()
		return resp
	}
	project, err := p.models.ProjectManager.Get(uint(projectId))
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: err.Error()}
	}
	project.Name = ser.Name
	project.Description = ser.Description
	project.Owner = ser.Owner
	project.UpdateTime = time.Now()
	project.UpdateUser = c.User.Name
	project, err = p.models.ProjectManager.Update(project)
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = fmt.Sprintf("更新项目空间失败:%s", err.Error())
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
		if !p.models.UserRoleManager.HasScopeRole(c.User, types.RoleScopeProject, project.ID, types.RoleTypeViewer) {
			continue
		}
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
	var ser serializers.ProjectDeleteSerializer
	if err = c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}

	return p.projectService.Delete(uint(projectId), ser.DelResource)
}

func (p *Project) get(c *views.Context) *utils.Response {
	projectId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return p.projectService.Get(uint(projectId), true)
}

func (p *Project) clone(c *views.Context) *utils.Response {
	var ser serializers.ProjectCloneSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return p.projectService.Clone(&ser, c.User)
}

func (p *Project) getProjectResources(c *views.Context) *utils.Response {
	var ser serializers.ProjectResourcesSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return p.projectService.GetProjectNamespaceResources(&ser)
}
