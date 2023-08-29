package main

import (
	"flag"
	"github.com/kubespace/kubespace/pkg/server"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
)

var (
	insecurePort    = flag.Int("insecure-port", utils.LookupEnvOrInt("INSECURE_PORT", 80), "Server insecure port to listen.")
	port            = flag.Int("port", utils.LookupEnvOrInt("SECURE_PORT", 443), "Server port to listen.")
	redisAddress    = flag.String("redis-address", utils.LookupEnvOrString("REDIS_ADDRESS", "localhost:6379"), "redis address used.")
	redisDB         = flag.Int("redis-db", utils.LookupEnvOrInt("REDIS_DB", 0), "redis db used.")
	redisPassword   = flag.String("redis-password", utils.LookupEnvOrString("REDIS_PASSWORD", "123abc,.;"), "redis password used.")
	certFile        = flag.String("cert-file", utils.LookupEnvOrString("CERT_FILE", ""), "cert file path for tls used.")
	keyFile         = flag.String("cert-key-file", utils.LookupEnvOrString("CERT_KEY_FILE", ""), "cert key file path for tls used.")
	mysqlHost       = flag.String("mysql-host", utils.LookupEnvOrString("MYSQL_HOST", "localhost:3306"), "mysql address used.")
	mysqlUser       = flag.String("mysql-user", utils.LookupEnvOrString("MYSQL_USER", "root"), "mysql db user.")
	mysqlPassword   = flag.String("mysql-password", utils.LookupEnvOrString("MYSQL_PASSWORD", ""), "mysql password used.")
	mysqlDbName     = flag.String("mysql-dbname", utils.LookupEnvOrString("MYSQL_DBNAME", "kubespace"), "mysql db used.")
	agentVersion    = flag.String("agent-version", utils.LookupEnvOrString("AGENT_VERSION", "latest"), "kubespace agent version.")
	agentRepository = flag.String("agent-repository", utils.LookupEnvOrString("AGENT_REPOSITORY", "kubespace/kubespace-agent"), "kubespace agent image repository.")
	releaseVersion  = flag.String("release-version", utils.LookupEnvOrString("RELEASE_VERSION", ""), "kubespace release version.")
)

func createServerOptions() *config.ServerOptions {
	return &config.ServerOptions{
		InsecurePort:    *insecurePort,
		Port:            *port,
		RedisAddress:    *redisAddress,
		RedisDB:         *redisDB,
		RedisPassword:   *redisPassword,
		CertFilePath:    *certFile,
		KeyFilePath:     *keyFile,
		MysqlHost:       *mysqlHost,
		MysqlUser:       *mysqlUser,
		MysqlPassword:   *mysqlPassword,
		MysqlDbName:     *mysqlDbName,
		AgentVersion:    *agentVersion,
		AgentRepository: *agentRepository,
		ReleaseVersion:  *releaseVersion,
	}
}

func buildServer() (*server.Server, error) {
	serverOptions := createServerOptions()
	serverConfig, err := config.NewServerConfig(serverOptions)
	if err != nil {
		klog.Error("New server config error:", err)
		return nil, err
	}
	return server.NewServer(serverConfig)
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	flag.VisitAll(func(flag *flag.Flag) {
		klog.Infof("FLAG: --%s=%q", flag.Name, flag.Value)
	})
	var err error
	svr, err := buildServer()
	if err != nil {
		panic(err)
	}
	svr.Run()
}
