package main

import (
	"flag"
	"github.com/kubespace/kubespace/pkg/core"
	"github.com/kubespace/kubespace/pkg/options"
	"k8s.io/klog"
)

var (
	insecurePort  = flag.Int("insecure-port", 80, "Server insecure port to listen.")
	port          = flag.Int("port", 443, "Server port to listen.")
	redisAddress  = flag.String("redis-address", "localhost:6379", "redis address used.")
	redisDB       = flag.Int("redis-db", 0, "redis db used.")
	redisPassword = flag.String("redis-password", "", "redis password used.")
	certFile      = flag.String("cert-file", "", "cert file path for tls used.")
	keyFile       = flag.String("cert-key-file", "", "cert key file path for tls used.")
	mysqlHost     = flag.String("mysql-host", "127.0.0.1:3306", "mysql address used.")
	mysqlUser     = flag.String("mysql-user", "root", "mysql db user.")
	mysqlPassword = flag.String("mysql-password", "123abc,.;", "mysql password used.")
	mysqlDbName   = flag.String("mysql-dbname", "kubespace", "mysql db used.")
)

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
	server, err := buildServer()
	if err != nil {
		panic(err)
	}
	server.Run()
	//rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
	//	Name: "origin",
	//	URLs: []string{"https://github.com/lzeen/testapp"},
	//})
	//
	//log.Print("Fetching tags...")
	//
	//// We can then use every Remote functions to retrieve wanted information
	//refs, err := rem.ListContext(context.Background(), &git.ListOptions{})
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//// Filters the references list and only keeps tags
	//var tags []string
	//for _, ref := range refs {
	//	if ref.Name().IsTag() {
	//		tags = append(tags, ref.Name().Short())
	//	}
	//}
	//
	//if len(tags) == 0 {
	//	log.Println("No tags!")
	//	return
	//}
	//log.Println(tags)
}
