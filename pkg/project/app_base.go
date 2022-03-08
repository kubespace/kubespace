package project

import (
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"helm.sh/helm/v3/pkg/chart/loader"
	"io"
)

type AppBaseService struct {
	models *model.Models
}

func NewAppBaseService(models *model.Models) *AppBaseService {
	return &AppBaseService{
		models: models,
	}
}

func (b *AppBaseService) ResolveChart(chartIn io.Reader) *utils.Response {
	charts, err := loader.LoadArchive(chartIn)
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: err.Error()}
	}
	data := map[string]interface{}{
		"package_name":    charts.Name(),
		"description":     charts.Metadata.Description,
		"package_version": charts.Metadata.Version,
		"app_version":     charts.AppVersion(),
	}
	return &utils.Response{Code: code.Success, Data: data}
}
