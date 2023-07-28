package main

import (
	"flag"
	"github.com/kubespace/kubespace/pkg/controller"
	"github.com/kubespace/kubespace/pkg/controller/pipeline_run"
	"github.com/kubespace/kubespace/pkg/controller/pipeline_trigger"
	"github.com/kubespace/kubespace/pkg/controller/spacelet"
	"github.com/kubespace/kubespace/pkg/core/db"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
)

var (
	redisAddress  = flag.String("redis-address", utils.LookupEnvOrString("REDIS_ADDRESS", "localhost:6379"), "redis address used.")
	redisDB       = flag.Int("redis-db", utils.LookupEnvOrInt("REDIS_DB", 0), "redis db used.")
	redisPassword = flag.String("redis-password", utils.LookupEnvOrString("REDIS_PASSWORD", "123abc,.;"), "redis password used.")
	mysqlHost     = flag.String("mysql-host", utils.LookupEnvOrString("MYSQL_HOST", "localhost:3306"), "mysql address used.")
	mysqlUser     = flag.String("mysql-user", utils.LookupEnvOrString("MYSQL_USER", "root"), "mysql db user.")
	mysqlPassword = flag.String("mysql-password", utils.LookupEnvOrString("MYSQL_PASSWORD", ""), "mysql password used.")
	mysqlDbName   = flag.String("mysql-dbname", utils.LookupEnvOrString("MYSQL_DBNAME", "kubespace"), "mysql db used.")
	resyncSec     = flag.Int("resync-seconds", utils.LookupEnvOrInt("RESYNC_SECONDS", 5), "controller list resync seconds.")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	flag.VisitAll(func(flag *flag.Flag) {
		klog.Infof("FLAG: --%s=%q", flag.Name, flag.Value)
	})
	stopCh := make(chan struct{})

	dbConfig := &db.Config{
		Mysql: &db.MysqlConfig{
			Username: *mysqlUser,
			Password: *mysqlPassword,
			Host:     *mysqlHost,
			DbName:   *mysqlDbName,
		},
		Redis: &db.RedisConfig{
			Addr:     *redisAddress,
			Password: *redisPassword,
			DB:       *redisDB,
		},
	}
	controllerConfig, err := controller.NewConfig(dbConfig, *resyncSec)
	if err != nil {
		panic(err)
	}

	// 流水线构建controller
	pipelineRunController := pipeline_run.NewPipelineRunController(controllerConfig)
	pipelineRunController.Run(stopCh)

	// 流水线自动触发controller
	pipelineTriggerController := pipeline_trigger.NewPipelineTriggerController(controllerConfig)
	pipelineTriggerController.Run(stopCh)

	spaceletController := spacelet.NewSpaceletController(controllerConfig)
	spaceletController.Run(stopCh)

	<-stopCh
}
