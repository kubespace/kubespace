package main

import (
	"flag"
	conf2 "github.com/kubespace/kubespace/pkg/conf"
	"github.com/kubespace/kubespace/pkg/core"
	"github.com/kubespace/kubespace/pkg/options"
	"k8s.io/klog"
	"os"
	"strconv"
)

var (
	insecurePort      = flag.Int("insecure-port", LookupEnvOrInt("INSECURE_PORT", 80), "Server insecure port to listen.")
	port              = flag.Int("port", LookupEnvOrInt("SECURE_PORT", 443), "Server port to listen.")
	redisAddress      = flag.String("redis-address", LookupEnvOrString("REDIS_ADDRESS", "localhost:6379"), "redis address used.")
	redisDB           = flag.Int("redis-db", LookupEnvOrInt("REDIS_DB", 0), "redis db used.")
	redisPassword     = flag.String("redis-password", LookupEnvOrString("REDIS_PASSWORD", "123abc,.;"), "redis password used.")
	certFile          = flag.String("cert-file", LookupEnvOrString("CERT_FILE", ""), "cert file path for tls used.")
	keyFile           = flag.String("cert-key-file", LookupEnvOrString("CERT_KEY_FILE", ""), "cert key file path for tls used.")
	mysqlHost         = flag.String("mysql-host", LookupEnvOrString("MYSQL_HOST", "localhost:3306"), "mysql address used.")
	mysqlUser         = flag.String("mysql-user", LookupEnvOrString("MYSQL_USER", "root"), "mysql db user.")
	mysqlPassword     = flag.String("mysql-password", LookupEnvOrString("MYSQL_PASSWORD", ""), "mysql password used.")
	mysqlDbName       = flag.String("mysql-dbname", LookupEnvOrString("MYSQL_DBNAME", "kubespace"), "mysql db used.")
	pipelinePluginUrl = flag.String("pipeline-plugin-url", LookupEnvOrString("PIPELINE_PLUGIN_URL", "http://127.0.0.1:8081/api/v1/plugin"), "pipeline plugin url.")
	agentVersion      = flag.String("agent-version", LookupEnvOrString("AGENT_VERSION", "latest"), "kubespace agent version.")
	agentRepository   = flag.String("agent-repository", LookupEnvOrString("AGENT_REPOSITORY", "kubespace/agent"), "kubespace agent version.")
)

func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func LookupEnvOrInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			klog.Fatalf("LookupEnvOrInt[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}

func createServerOptions() *options.ServerOptions {
	return &options.ServerOptions{
		InsecurePort:  *insecurePort,
		Port:          *port,
		RedisAddress:  *redisAddress,
		RedisDB:       *redisDB,
		RedisPassword: *redisPassword,
		CertFilePath:  *certFile,
		KeyFilePath:   *keyFile,
		MysqlHost:     *mysqlHost,
		MysqlUser:     *mysqlUser,
		MysqlPassword: *mysqlPassword,
		MysqlDbName:   *mysqlDbName,
	}
}

func buildServer() (*core.Server, error) {
	serverOptions := createServerOptions()
	serverConfig, err := core.NewServerConfig(serverOptions)
	if err != nil {
		klog.Error("New server config error:", err)
		return nil, err
	}
	return core.NewServer(serverConfig)
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	flag.VisitAll(func(flag *flag.Flag) {
		klog.Infof("FLAG: --%s=%q", flag.Name, flag.Value)
	})
	var err error
	conf2.AppConfig.PipelinePluginUrl = *pipelinePluginUrl
	conf2.AppConfig.AgentVersion = *agentVersion
	conf2.AppConfig.AgentRepository = *agentRepository
	server, err := buildServer()
	if err != nil {
		panic(err)
	}
	server.Run()
}
