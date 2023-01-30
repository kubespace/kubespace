package config

import (
	"github.com/kubespace/kubespace/pkg/informer"
	"github.com/kubespace/kubespace/pkg/model"
)

type ServiceConfig struct {
	Models          *model.Models
	InformerFactory informer.Factory
}
