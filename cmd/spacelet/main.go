package main

import (
	"flag"
	"github.com/kubespace/kubespace/pkg/spacelet"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
)

var (
	port      = flag.Int("port", utils.LookupEnvOrInt("PORT", 7520), "spacelet port to listen.")
	hostIp    = flag.String("host-ip", utils.LookupEnvOrString("HOST_IP", ""), "spacelet host ip.")
	dataDir   = flag.String("data-dir", utils.LookupEnvOrString("DATA_DIR", "/data"), "data directory.")
	serverUrl = flag.String("server-url", utils.LookupEnvOrString("SERVER_URL", "http://kubespace"), "kubespace server url.")
)

func buildServer() (*spacelet.Server, error) {
	config, err := spacelet.NewConfig(&spacelet.Options{
		HostIp:    *hostIp,
		Port:      *port,
		DataDir:   *dataDir,
		ServerUrl: *serverUrl,
	})
	if err != nil {
		klog.Error("New server config error:", err)
		return nil, err
	}
	return spacelet.NewServer(config)
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
	select {}
}
