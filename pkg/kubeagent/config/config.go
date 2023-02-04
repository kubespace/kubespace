package config

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/kubernetes/config"
	"github.com/kubespace/kubespace/pkg/utils"
)

type AgentOptions struct {
	KubeConfigFile string
	AgentToken     string
	ServerHost     string
}

type AgentConfig struct {
	Token        string
	KubeConfig   *config.KubeConfig
	ServerHost   string
	ServerClient *utils.HttpClient
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
	if a.ServerClient, err = utils.NewHttpClient(fmt.Sprintf("http://%s", a.ServerHost)); err != nil {
		return nil, err
	}
	return
}
