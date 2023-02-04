package main

import (
	"flag"
	"github.com/kubespace/kubespace/pkg/kubeagent"
	"github.com/kubespace/kubespace/pkg/kubeagent/config"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
)

var (
	kubeConfigFile = flag.String("kubeconfig", "", "Path to kubeconfig file with authorization and master location information.")
	agentToken     = flag.String("token", utils.LookupEnvOrString("TOKEN", "local"), "Agent token to connect to server.")
	serverHost     = flag.String("server-host", utils.LookupEnvOrString("SERVER_HOST", "kubespace"), "Server host:port agent to connect.")
)

func buildAgent() (*kubeagent.Agent, error) {
	options := &config.AgentOptions{
		KubeConfigFile: *kubeConfigFile,
		AgentToken:     *agentToken,
		ServerHost:     *serverHost,
	}
	agentConfig, err := config.NewAgentConfig(options)
	if err != nil {
		klog.Error("New agent config error:", err)
		return nil, err
	}
	return kubeagent.NewAgent(agentConfig), nil
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	flag.VisitAll(func(flag *flag.Flag) {
		klog.Infof("FLAG: --%s=%q", flag.Name, flag.Value)
	})
	agent, err := buildAgent()
	if err != nil {
		panic(err)
	}
	stopCh := make(chan struct{})
	agent.Run(stopCh)
}
