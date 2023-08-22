package audit

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/manager/audit"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/utils"
	"net/http"
)

type AuditOperate struct {
	Views  []*views.View
	models *model.Models
}

func NewAuditOperate(models *model.Models) *AuditOperate {
	a := &AuditOperate{
		models: models,
	}
	a.Views = []*views.View{
		views.NewView(http.MethodGet, "", a.list),
	}
	return a
}

func (s *AuditOperate) list(c *views.Context) *utils.Response {
	var listCond audit.AuditOperateListCondition
	if err := c.ShouldBindQuery(&listCond); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	aos, page, err := s.models.AuditOperateManager.List(&listCond)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}

	return &utils.Response{Code: code.Success, Data: map[string]interface{}{
		"data":       aos,
		"pagination": page,
	}}
}
