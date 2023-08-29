package project

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
)

type listHandler struct {
	models *model.Models
}

func ListHandler(conf *config.ServerConfig) api.Handler {
	return &listHandler{models: conf.Models}
}

type listProjectData struct {
	*types.Project `json:",inline"`
	Cluster        *types.Cluster `json:"cluster"`
}

func (h *listHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, nil, nil
}

func (h *listHandler) Handle(c *api.Context) *utils.Response {
	projects, err := h.models.ProjectManager.List()
	if err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}
	var data []*listProjectData
	clusters := make(map[string]*types.Cluster)

	for i, project := range projects {
		if !h.models.UserRoleManager.AuthRole(c.User, types.ScopeProject, project.ID, types.RoleViewer) {
			continue
		}
		cluster, ok := clusters[project.ClusterId]
		if !ok {
			cluster, err = h.models.ClusterManager.GetByName(project.ClusterId)
			if err != nil {
				klog.Errorf("get project id=%s cluster error: %s", project.ID, err.Error())
			}
			clusters[project.ClusterId] = cluster
		}
		data = append(data, &listProjectData{
			Project: projects[i],
			Cluster: cluster,
		})
	}
	return c.ResponseOK(data)
}
