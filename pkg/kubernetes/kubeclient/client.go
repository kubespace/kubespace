package kubeclient

import (
	"fmt"
	utilversion "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"time"
)

type Client interface {
	kubernetes.Interface
	Dynamic() dynamic.Interface
	Discovery() discovery.DiscoveryInterface
	RestConfig() *rest.Config
	VersionGreaterThan(string) bool
	ServerVersion() *version.Info
}

type client struct {
	kubernetes.Interface
	dynamic       dynamic.Interface
	discovery     discovery.DiscoveryInterface
	restConfig    *rest.Config
	serverVersion *version.Info
}

func NewClientByKubeConfig(kubeConfig string) (Client, error) {
	restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeConfig))
	if err != nil {
		return nil, err
	}
	return NewClientWithRestConfig(restConfig)
}

func NewClientByKubeConfigFile(kubeConfigFile string) (Client, error) {
	if kubeConfigFile == "" {
		return nil, fmt.Errorf("no kubeconfig file")
	}
	var restConfig *rest.Config
	var err error
	klog.Infof("using kubeconfig file: %s", kubeConfigFile)
	// use the current context in kubeconfig
	restConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to build config: %s", err.Error())
	}

	return NewClientWithRestConfig(restConfig)
}

func NewClientByInCluster() (Client, error) {
	var restConfig *rest.Config
	var err error
	restConfig, err = rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build incluster config: %s", err.Error())
	}
	return NewClientWithRestConfig(restConfig)
}

func NewClientWithRestConfig(restConfig *rest.Config) (Client, error) {
	restConfig.Timeout = time.Second * 5
	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	serverVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		return nil, err
	}
	return &client{
		Interface:     clientSet,
		dynamic:       dynamicClient,
		restConfig:    restConfig,
		discovery:     discoveryClient,
		serverVersion: serverVersion,
	}, nil
}

func (c *client) RestConfig() *rest.Config {
	return c.restConfig
}

func (c *client) Dynamic() dynamic.Interface {
	return c.dynamic
}

func (c *client) Discovery() discovery.DiscoveryInterface {
	return c.discovery
}

func (c *client) ServerVersion() *version.Info {
	return c.serverVersion
}

func (c *client) VersionGreaterThan(version string) bool {
	if utilversion.MustParseSemantic(c.serverVersion.GitVersion).LessThan(utilversion.MustParseSemantic(version)) {
		return false
	}
	return true
}
