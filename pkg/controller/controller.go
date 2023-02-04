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
	DataDir         string
}

func NewConfig(dbConfig *db.Config, resyncSec int, dataDir string) (*Config, error) {
	dB, err := db.NewDb(dbConfig)
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

	serviceFactory := service.NewServiceFactory(&service.Config{
		Models:          models,
		InformerFactory: informerFactory,
	})
	return &Config{
		Models:          models,
		InformerFactory: informerFactory,
		ServiceFactory:  serviceFactory,
		DataDir:         dataDir,
	}, nil
}
