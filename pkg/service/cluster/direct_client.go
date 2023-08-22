package cluster

import (
	"errors"
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/kubernetes"
	"github.com/kubespace/kubespace/pkg/kubernetes/config"
	kubetypes "github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
)

type directClient struct{}

func (d *directClient) getKubeFactory(cluster *types.Cluster) (kubernetes.KubeFactory, error) {
	kubeConfig, err := config.NewKubeConfig(&config.Options{KubeConfigString: cluster.KubeConfig})
	if err != nil {
		return nil, err
	}
	kubeFactory := kubernetes.NewKubeFactory(kubeConfig)
	return kubeFactory, nil
}

func (d *directClient) request(cluster *types.Cluster, resType, action string, params interface{}) *utils.Response {
	kubeFactory, err := d.getKubeFactory(cluster)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	if resHandler, err := kubeFactory.GetResource(resType); err != nil {
		return &utils.Response{Code: code.GetError, Msg: err.Error()}
	} else {
		return resHandler.Handle(action, params)
	}
}

func (d *directClient) watch(cluster *types.Cluster, resType string, params interface{}) (Outer, error) {
	kubeFactory, err := d.getKubeFactory(cluster)
	if err != nil {
		return nil, err
	}
	resHandler, err := kubeFactory.GetResource(resType)
	if err != nil {
		return nil, err
	}
	o := &directOuter{outer: newOuter()}
	resp := resHandler.Watch(params, o)
	if !resp.IsSuccess() {
		return nil, fmt.Errorf(resp.Msg)
	}
	return o, nil

}

func (d *directClient) pods(cluster *types.Cluster) (PodClient, error) {
	kubeFactory, err := d.getKubeFactory(cluster)
	if err != nil {
		return nil, err
	}
	if podHandler, err := kubeFactory.GetPod(); err != nil {
		return nil, err
	} else {
		return &directPod{podHandler}, nil
	}
}

type directOuter struct {
	*outer
}

func (d *directOuter) Write(out interface{}) (err error) {
	defer utils.HandleCrash(func(r interface{}) {
		if d.stopped {
			err = fmt.Errorf("writer closed")
		} else {
			err = fmt.Errorf("writer crashed")
		}
	})
	if d.stopped {
		return fmt.Errorf("writer closed")
	}
	d.outCh <- out
	return nil
}

type directPod struct {
	kubernetes.PodHandler
}

func (d *directPod) Exec(params interface{}) (PodExec, error) {
	exec := &directPodExec{
		pod: d,
		directOuter: &directOuter{
			outer: newOuter(),
		},
	}
	resp := d.PodHandler.Exec(params, exec)
	if !resp.IsSuccess() {
		return nil, fmt.Errorf(resp.Msg)
	}
	return exec, nil
}

func (d *directPod) Log(params interface{}) (Outer, error) {
	logOuter := &directOuter{outer: newOuter()}
	resp := d.PodHandler.Log(params, logOuter)
	if !resp.IsSuccess() {
		return nil, fmt.Errorf(resp.Msg)
	}
	return logOuter, nil
}

type directPodExec struct {
	*directOuter
	pod *directPod
}

func (d *directPodExec) Stdin(params interface{}) error {
	resp := d.pod.Handle(kubetypes.StdinAction, params)
	if !resp.IsSuccess() {
		return errors.New(resp.Msg)
	}
	return nil
}
