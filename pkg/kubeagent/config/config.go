package config

import (
	"github.com/kubespace/kubespace/pkg/kubernetes/config"
)

type AgentOptions struct {
	KubeConfigFile string
	AgentToken     string
	ServerHost     string
}

type AgentConfig struct {
	Token      string
	KubeConfig *config.KubeConfig
	ServerHost string
}

func NewAgentConfig(options *AgentOptions) (a *AgentConfig, err error) {
	a = &AgentConfig{
		Token:      options.AgentToken,
		ServerHost: options.ServerHost,
	}
	kubeOptions := &config.Options{}
	if options.KubeConfigFile != "" {
		kubeOptions.KubeConfigFile = options.KubeConfigFile
	} else {
		kubeOptions.InCluster = true
	}
	if a.KubeConfig, err = config.NewKubeConfig(kubeOptions); err != nil {
		return nil, err
	}
	return
}
