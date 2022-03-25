package project

import (
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
)

type ServiceProject struct {
	models *model.Models
	*kube_resource.KubeResources
	appService *AppService
}

func NewProjectService(models *model.Models, kr *kube_resource.KubeResources, appService *AppService) *ServiceProject {
	return &ServiceProject{
		models:        models,
		KubeResources: kr,
		appService:    appService,
	}
}

func (p *ServiceProject) Delete(projectId uint) *utils.Response {
	resp := &utils.Response{Code: code.Success}
	project, err := p.models.ProjectManager.Get(projectId)
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "获取工作空间失败: " + err.Error()
		return resp
	}
	apps, err := p.appService.ListApp(projectId)
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: err.Error()}
	}
	for _, app := range apps {
		if app.Status != types.AppStatusUninstall {
			return &utils.Response{Code: code.DeleteError, Msg: "删除工作空间失败：应用" + app.Name + "正在运行"}
		}
	}
	err = p.models.ProjectManager.Delete(project)
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "删除工作空间失败: " + err.Error()
		return resp
	}
	return resp
}

func (p *ServiceProject) Get(projectId uint) *utils.Response {
	project, err := p.models.ProjectManager.Get(projectId)
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: "获取工作空间失败: " + err.Error()}
	}
	return &utils.Response{Code: code.Success, Data: project}
}
