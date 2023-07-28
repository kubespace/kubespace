package spacelet

import (
	"fmt"
	"github.com/google/gopacket/routing"
	"github.com/kubespace/kubespace/pkg/third/httpclient"
	"github.com/kubespace/kubespace/pkg/utils"
	"net"
	"net/url"
	"strings"
)

type Options struct {
	HostIp    string
	Port      int
	DataDir   string
	ServerUrl string
}

type Config struct {
	// kubespace服务地址
	ServerUrl string
	Client    *httpclient.HttpClient
	// spacelet所在服务器的主机ip
	HostIp string
	// spacelet服务启动端口
	Port int
	// 执行流水线任务的数据目录
	DataDir string
	// 注册之后获取的token，用来进行认证
	Token string
}

func getServerIp(serverUrl string) (net.IP, error) {
	u, err := url.Parse(serverUrl)
	if err != nil {
		return nil, err
	}
	host := u.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	if len(ips) > 0 {
		return ips[0], nil
	}
	return nil, fmt.Errorf("not found server ip")
}

func getSpaceletHostIp(options *Options) string {
	defer utils.HandleCrash()
	serverIp, err := getServerIp(options.ServerUrl)
	if err != nil {
		return ""
	}
	if serverIp == nil {
		return ""
	}
	router, err := routing.New()
	if err != nil {
		return ""
	}
	_, _, srcIp, err := router.Route(serverIp)
	if err != nil {
		return ""
	}
	return srcIp.String()
}

func NewConfig(options *Options) (*Config, error) {
	httpcli, err := httpclient.NewHttpClient(options.ServerUrl)
	if err != nil {
		return nil, err
	}
	hostIp := options.HostIp
	if hostIp == "" {
		hostIp = getSpaceletHostIp(options)
	}
	return &Config{
		ServerUrl: options.ServerUrl,
		Client:    httpcli,
		HostIp:    hostIp,
		Port:      options.Port,
		DataDir:   options.DataDir,
		Token:     "",
	}, nil
}
