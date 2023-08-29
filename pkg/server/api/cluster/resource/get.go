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

func (h *getHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
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

func (h *getHandler) Handle(c *api.Context) *utils.Response {
	var params interface{}
	if c.Param("resType") == kubetypes.CustomResourceType {
		params = &resource.CustomResourceQueryParams{
			Name:      c.Param("name"),
			Namespace: c.Param("namespace"),
		}
	} else {
		params = &resource.QueryParams{
			Name:      c.Param("name"),
			Namespace: c.Param("namespace"),
		}
	}
	if err := c.ShouldBindQuery(params); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return h.kubeClient.Get(c.Param("id"), c.Param("resType"), params)
}
