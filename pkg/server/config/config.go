package config

import (
	coredb "github.com/kubespace/kubespace/pkg/core/db"
	"github.com/kubespace/kubespace/pkg/informer"
	listwatcherconfig "github.com/kubespace/kubespace/pkg/informer/listwatcher/config"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/migrate"
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
	DB              *coredb.DB
	Models          *model.Models
	InformerFactory informer.Factory
	ServiceFactory  *service.Factory
}

func NewServerConfig(op *ServerOptions) (*ServerConfig, error) {
	db, err := coredb.NewDB(&coredb.Config{
		Mysql: &coredb.MysqlConfig{
			Username: op.MysqlUser,
			Password: op.MysqlPassword,
			Host:     op.MysqlHost,
			DbName:   op.MysqlDbName,
		},
		Redis: &coredb.RedisConfig{
			Addr:     op.RedisAddress,
			Password: op.RedisPassword,
			DB:       op.RedisDB,
		},
	})
	if err != nil {
		return nil, err
	}
	// 数据库连接之后，首先初始化迁移数据
	if err = migrate.NewMigrate(db.Instance).Do(); err != nil {
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
	listWatcherConfig := listwatcherconfig.NewListWatcherConfig(db, op.ListWatcherResyncSec)
	models, err := model.NewModels(&model.Config{
		DB:                db,
		ListWatcherConfig: listWatcherConfig,
	})
	if err != nil {
		return nil, err
	}
	informerFactory := informer.NewInformerFactory(models.ListWatcherConfig)
	serviceFactory := service.NewServiceFactory(service.NewConfig(models))
	return &ServerConfig{
		AgentVersion:    op.AgentVersion,
		AgentRepository: op.AgentRepository,
		DB:              db,
		InsecurePort:    op.InsecurePort,
		Port:            op.Port,
		CertFilePath:    certFilePath,
		KeyFilePath:     keyFilePath,
		Models:          models,
		InformerFactory: informerFactory,
		ServiceFactory:  serviceFactory,
	}, nil
}
