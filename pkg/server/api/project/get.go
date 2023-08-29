package project

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	kubetypes "github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/service/cluster"
	"github.com/kubespace/kubespace/pkg/utils"
)

type getHandler struct {
	models     *model.Models
	kubeClient *cluster.KubeClient
}

func GetHandler(conf *config.ServerConfig) api.Handler {
	return &getHandler{
		models:     conf.Models,
		kubeClient: conf.ServiceFactory.Cluster.KubeClient,
	}
}

type getProjectData struct {
	*types.Project `json:",inline"`
	Cluster        *types.Cluster `json:"cluster"`
	Resource       interface{}    `json:"resource"`
}

func (h *getHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	projectId, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopeProject,
		ScopeId: projectId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *getHandler) Handle(c *api.Context) *utils.Response {
	projectId, _ := utils.ParseUint(c.Param("id"))
	project, err := h.models.ProjectManager.Get(projectId)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, "获取工作空间失败: "+err.Error()))
	}

	clusterObj, err := h.models.ClusterManager.GetByName(project.ClusterId)
	if err != nil {
		return c.ResponseError(errors.New(code.GetError, "获取集群信息失败: %s"+err.Error()))
	}
	resp := h.kubeClient.Get(project.ClusterId, kubetypes.ClusterType, map[string]interface{}{
		"workspace": project.ID,
		"namespace": project.Namespace,
	})
	if !resp.IsSuccess() {
		return resp
	}

	return c.ResponseOK(getProjectData{
		Project:  project,
		Cluster:  clusterObj,
		Resource: resp.Data,
	})
}
