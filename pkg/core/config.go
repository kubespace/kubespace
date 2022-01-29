package core

import (
	"github.com/kubespace/kubespace/pkg/model/mysql"
	"github.com/kubespace/kubespace/pkg/options"
	"github.com/kubespace/kubespace/pkg/redis"
	"github.com/kubespace/kubespace/pkg/utils"
	"time"
)

type ServerConfig struct {
	InsecurePort int
	Port         int
	RedisOptions *redis.Options
	MysqlOptions *mysql.Options
	CertFilePath string
	KeyFilePath  string
}

func NewServerConfig(op *options.ServerOptions) (*ServerConfig, error) {
	redisOp := &redis.Options{
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
	return &ServerConfig{
		InsecurePort: op.InsecurePort,
		Port:         op.Port,
		RedisOptions: redisOp,
		MysqlOptions: mysqlOp,
		CertFilePath: certFilePath,
		KeyFilePath:  keyFilePath,
	}, nil
}
