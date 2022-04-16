package conf

type GlobalConf struct {
	PipelinePluginUrl string
	AgentVersion      string
}

var AppConfig = &GlobalConf{}
