package spacelet

import (
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
)

type SpaceletService struct {
	models *model.Models
}

func NewSpaceletService(models *model.Models) *SpaceletService {
	return &SpaceletService{models: models}
}

func (s *SpaceletService) Delete(id uint) *utils.Response {
	spaceletObj, err := s.models.SpaceletManager.Get(id)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	if spaceletObj.Status == types.SpaceletStatusOnline {
		return &utils.Response{Code: code.DeleteError, Msg: "当前Spacelet节点在线，不能删除"}
	}
	if err = s.models.SpaceletManager.Delete(id); err != nil {
		return &utils.Response{Code: code.DeleteError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success}
}
