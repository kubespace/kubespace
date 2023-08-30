package spacelet

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
	"time"
)

type updateHandler struct {
	models *model.Models
}

func UpdateHandler(conf *config.ServerConfig) api.Handler {
	return &updateHandler{models: conf.Models}
}

type updateSpaceletBody struct {
	Labels map[string]string `json:"labels"`
}

func (h *updateHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return true, &api.AuthPerm{
		Scope:   types.ScopePlatform,
		ScopeId: 0,
		Role:    types.RoleEditor,
	}, nil
}

func (h *updateHandler) Handle(c *api.Context) *utils.Response {
	id, err := utils.ParseUint(c.Param("id"))
	if err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	var body updateSpaceletBody
	if err = c.ShouldBind(&body); err != nil {
		return c.ResponseError(errors.New(code.ParamsError, err))
	}
	spaceletObj, err := h.models.SpaceletManager.Get(id)
	if err != nil {
		return c.ResponseError(errors.New(code.DataNotExists, err))
	}
	if body.Labels != nil {
		spaceletObj.Labels = body.Labels
	}
	spaceletObj.UpdateTime = time.Now()
	if err = h.models.SpaceletManager.Save(spaceletObj); err != nil {
		return c.ResponseError(errors.New(code.DBError, err))
	}
	return c.ResponseOK(nil)
}
