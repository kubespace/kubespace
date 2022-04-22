package conf

type GlobalConf struct {
	PipelinePluginUrl string
	AgentVersion      string
	AgentRepository   string
}

var AppConfig = &GlobalConf{}
