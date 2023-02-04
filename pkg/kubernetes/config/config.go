package config

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/kubernetes/kubeclient"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeYaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/restmapper"
)

type Options struct {
	InCluster        bool
	KubeConfigFile   string
	KubeConfigString string
}

type KubeConfig struct {
	Client          kubeclient.Client
	DecUnstructured runtime.Serializer
	RestMapper      *restmapper.DeferredDiscoveryRESTMapper
}

func NewKubeConfig(options *Options) (c *KubeConfig, err error) {
	var client kubeclient.Client
	if options.InCluster {
		client, err = kubeclient.NewClientByInCluster()
	} else if options.KubeConfigString != "" {
		client, err = kubeclient.NewClientByKubeConfig(options.KubeConfigString)
	} else if options.KubeConfigFile != "" {
		client, err = kubeclient.NewClientByKubeConfigFile(options.KubeConfigFile)
	} else {
		return nil, fmt.Errorf("no config for kubernetes client")
	}
	if err != nil {
		return
	}
	return &KubeConfig{
		Client:          client,
		DecUnstructured: runtimeYaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme),
		RestMapper:      restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(client.Discovery())),
	}, nil
}
