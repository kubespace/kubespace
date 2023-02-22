package spacelet

type Options struct {
	HostIp    string
	Port      int
	DataDir   string
	ServerUrl string
}

type Config struct {
	// kubespace服务地址
	ServerUrl string
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
	return &Config{
		ServerUrl: options.ServerUrl,
		HostIp:    options.HostIp,
		Port:      options.Port,
		DataDir:   options.DataDir,
		Token:     "",
	}, nil
}
