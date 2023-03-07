package spacelet

import (
	"github.com/kubespace/kubespace/pkg/third/httpclient"
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

func NewConfig(options *Options) (*Config, error) {
	httpcli, err := httpclient.NewHttpClient(options.ServerUrl)
	if err != nil {
		return nil, err
	}
	return &Config{
		ServerUrl: options.ServerUrl,
		Client:    httpcli,
		HostIp:    options.HostIp,
		Port:      options.Port,
		DataDir:   options.DataDir,
		Token:     "",
	}, nil
}
