package project

import (
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
)

type AppStoreService struct {
	*AppBaseService
}

func NewAppStoreService(appBaseService *AppBaseService) *AppStoreService {
	return &AppStoreService{
		AppBaseService: appBaseService,
	}
}

func (s *AppStoreService) CreateStoreApp(user *types.User) *utils.Response {
	return &utils.Response{Code: code.Success}
}
