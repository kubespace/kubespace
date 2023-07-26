package spacelet

import (
	"github.com/kubespace/kubespace/pkg/model"
	spaceletmanager "github.com/kubespace/kubespace/pkg/model/manager/spacelet"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
)

type SpaceletService struct {
	models *model.Models
}

func NewSpaceletService(models *model.Models) *SpaceletService {
	return &SpaceletService{models: models}
}

func (s *SpaceletService) List() *utils.Response {
	spacelets, err := s.models.SpaceletManager.List(&spaceletmanager.SpaceletListCondition{})
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	for _, sp := range spacelets {
		sp.Token = ""
	}
	return &utils.Response{Code: code.Success, Data: spacelets}
}
