package controller

import (
	"github.com/kubespace/kubespace/pkg/core/db"
	"github.com/kubespace/kubespace/pkg/informer"
	listwatcherconfig "github.com/kubespace/kubespace/pkg/informer/listwatcher/config"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/service"
)

type Config struct {
	Models          *model.Models
	InformerFactory informer.Factory
	ServiceFactory  *service.Factory
}

func NewConfig(dbConfig *db.Config, resyncSec int) (*Config, error) {
	dB, err := db.NewDB(dbConfig)
	if err != nil {
		return nil, err
	}
	listWatcherConfig := listwatcherconfig.NewListWatcherConfig(dB, resyncSec)
	models, err := model.NewModels(&model.Config{
		DB:                dB,
		ListWatcherConfig: listWatcherConfig,
	})
	if err != nil {
		return nil, err
	}
	informerFactory := informer.NewInformerFactory(listWatcherConfig)

	serviceFactory := service.NewServiceFactory(service.NewConfig(models))
	return &Config{
		Models:          models,
		InformerFactory: informerFactory,
		ServiceFactory:  serviceFactory,
	}, nil
}
