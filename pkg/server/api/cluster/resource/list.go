package resource

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/kubernetes/resource"
	kubetypes "github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/service/cluster"
	"github.com/kubespace/kubespace/pkg/utils"
)

type listHandler struct {
	models     *model.Models
	kubeClient *cluster.KubeClient
}

func ListHandler(conf *config.ServerConfig) api.Handler {
	return &listHandler{
		models:     conf.Models,
		kubeClient: conf.ServiceFactory.Cluster.KubeClient,
	}
}

func (h *listHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	clusterId, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return true, nil, errors.New(code.ParamsError, err)
	}
	return true, &api.AuthPerm{
		Scope:   types.ScopeCluster,
		ScopeId: clusterId,
		Role:    types.RoleViewer,
	}, nil
}

func (h *listHandler) Handle(c *api.Context) *utils.Response {
	var params interface{}
	if c.Param("resType") == kubetypes.CustomResourceType {
		params = &resource.CustomResourceQueryParams{}
	} else {
		params = &resource.QueryParams{}
	}
	if err := c.ShouldBind(params); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	return h.kubeClient.List(c.Param("id"), c.Param("resType"), params)
}
