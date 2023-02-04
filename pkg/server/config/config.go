package config

import (
	"github.com/kubespace/kubespace/pkg/core/db"
	"github.com/kubespace/kubespace/pkg/informer"
	listwatcherconfig "github.com/kubespace/kubespace/pkg/informer/listwatcher/config"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/service"
	"github.com/kubespace/kubespace/pkg/utils"
	"time"
)

type ServerConfig struct {
	InsecurePort    int
	Port            int
	CertFilePath    string
	KeyFilePath     string
	AgentVersion    string
	AgentRepository string
	DB              *db.DB
	Models          *model.Models
	InformerFactory informer.Factory
	ServiceFactory  *service.Factory
}

func NewServerConfig(op *ServerOptions) (*ServerConfig, error) {
	dB, err := db.NewDb(&db.Config{
		Mysql: &db.MysqlConfig{
			Username: op.MysqlUser,
			Password: op.MysqlPassword,
			Host:     op.MysqlHost,
			DbName:   op.MysqlDbName,
		},
		Redis: &db.RedisConfig{
			Addr:     op.RedisAddress,
			Password: op.RedisPassword,
			DB:       op.RedisDB,
		},
	})
	if err != nil {
		return nil, err
	}
	certFilePath := op.CertFilePath
	keyFilePath := op.KeyFilePath
	if op.CertFilePath == "" || op.KeyFilePath == "" {
		certFilePath = "cert.pem"
		keyFilePath = "key.pem"
		if !utils.PathExist(certFilePath) || !utils.PathExist(keyFilePath) {
			err := utils.GenerateCert(
				"localhost,127.0.0.1,*",
				time.Hour*24*365*10,
				false,
				"P256")
			if err != nil {
				return nil, err
			}
		}
	}
	listWatcherConfig := listwatcherconfig.NewListWatcherConfig(dB, op.ListWatcherResyncSec)
	models, err := model.NewModels(&model.Config{
		DB:                dB,
		ListWatcherConfig: listWatcherConfig,
	})
	if err != nil {
		return nil, err
	}
	informerFactory := informer.NewInformerFactory(models.ListWatcherConfig)
	serviceFactory := service.NewServiceFactory(&service.Config{
		Models:          models,
		InformerFactory: informerFactory,
	})
	return &ServerConfig{
		AgentVersion:    op.AgentVersion,
		AgentRepository: op.AgentRepository,
		DB:              dB,
		InsecurePort:    op.InsecurePort,
		Port:            op.Port,
		CertFilePath:    certFilePath,
		KeyFilePath:     keyFilePath,
		Models:          models,
		InformerFactory: informerFactory,
		ServiceFactory:  serviceFactory,
	}, nil
}
