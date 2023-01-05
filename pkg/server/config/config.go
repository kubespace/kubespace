package config

import (
	"github.com/kubespace/kubespace/pkg/core/mysql"
	coreRedis "github.com/kubespace/kubespace/pkg/core/redis"
	"github.com/kubespace/kubespace/pkg/informer"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/service"
	"github.com/kubespace/kubespace/pkg/service/config"
	"github.com/kubespace/kubespace/pkg/utils"
	"time"
)

type ServerConfig struct {
	InsecurePort    int
	Port            int
	RedisOptions    *coreRedis.Options
	MysqlOptions    *mysql.Options
	CertFilePath    string
	KeyFilePath     string
	Models          *model.Models
	InformerFactory informer.InformerFactory
	ServiceFactory  *service.Factory
}

func NewServerConfig(op *ServerOptions) (*ServerConfig, error) {
	redisOp := &coreRedis.Options{
		Addr:     op.RedisAddress,
		Password: op.RedisPassword,
		DB:       op.RedisDB,
	}
	mysqlOp := &mysql.Options{
		Host:     op.MysqlHost,
		Username: op.MysqlUser,
		Password: op.MysqlPassword,
		DbName:   op.MysqlDbName,
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
	models, err := model.NewModels(&model.Options{
		RedisOptions: redisOp,
		MysqlOptions: mysqlOp,
	})
	if err != nil {
		return nil, err
	}
	informerFactory := informer.NewInformerFactory(models.ListWatcherConfig)
	serviceFactory := service.NewServiceFactory(&config.ServiceConfig{
		Models:          models,
		InformerFactory: informerFactory,
	})
	return &ServerConfig{
		InsecurePort:    op.InsecurePort,
		Port:            op.Port,
		RedisOptions:    redisOp,
		MysqlOptions:    mysqlOp,
		CertFilePath:    certFilePath,
		KeyFilePath:     keyFilePath,
		Models:          models,
		InformerFactory: informerFactory,
		ServiceFactory:  serviceFactory,
	}, nil
}
