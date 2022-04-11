package plugins

import (
	"bytes"
	"fmt"
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"github.com/kubespace/kubespace/pkg/views/serializers"
	"k8s.io/klog"
	"runtime"
)

type PluginExecutor interface {
	Execute(params *PluginParams) (interface{}, error)
}

type PluginLogger struct {
	jobId  uint
	models *model.Models
	*bytes.Buffer
}

func (l *PluginLogger) Log(format string, a ...interface{}) {
	_, err := l.Buffer.WriteString(fmt.Sprintf(format+"\n", a...))
	if err != nil {
		klog.Errorf("write job %s log to buffer error: %s", l.jobId, err.Error())
	} else {
		err = l.models.PipelineJobLogManager.UpdateLog(l.jobId, l.Buffer.String())
		if err != nil {
			klog.Errorf("update job %s log to db errror: %s", l.jobId, err.Error())
		}
	}
}

type PluginParams struct {
	JobId     uint
	PluginKey string
	Params    map[string]interface{}
	Logger    *PluginLogger
}

type PluginCallback func(callbackSer serializers.PipelineCallbackSerializer) *utils.Response

type Plugins struct {
	Plugins  map[string]PluginExecutor
	callback PluginCallback
	*kube_resource.KubeResources
	*model.Models
}

func NewPlugins(models *model.Models, kr *kube_resource.KubeResources, callback PluginCallback) *Plugins {
	p := &Plugins{
		Plugins:       make(map[string]PluginExecutor),
		callback:      callback,
		Models:        models,
		KubeResources: kr,
	}
	p.Plugins[types.BuiltinPluginUpgradeApp] = UpgradeAppPlugin{Models: models, KubeResources: kr}
	return p
}

func (b *Plugins) Execute(pluginParams *PluginParams) *utils.Response {
	if pluginParams.PluginKey == "" {
		return &utils.Response{Code: code.PluginError, Msg: "not found plugin key parameter"}
	}
	executor, ok := b.Plugins[pluginParams.PluginKey]
	if !ok {
		return &utils.Response{Code: code.PluginError, Msg: "not found plugin executor: " + pluginParams.PluginKey}
	}
	pluginParams.Logger = &PluginLogger{
		jobId:  pluginParams.JobId,
		models: b.Models,
		Buffer: new(bytes.Buffer),
	}
	go b.doExecute(executor, pluginParams)
	return &utils.Response{Code: code.Success}
}

func (b *Plugins) doExecute(executor PluginExecutor, pluginParams *PluginParams) {
	defer func() {
		if err := recover(); err != nil {
			klog.Error("error: ", err)
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			klog.Errorf("==> %s\n", string(buf[:n]))
			pluginParams.Logger.Log("==> %s\n", string(buf[:n]))
			b.Callback(pluginParams.JobId, &utils.Response{Code: code.UnknownError, Msg: fmt.Sprintf("%v", err)})
		}
	}()
	result, err := executor.Execute(pluginParams)
	if err != nil {
		klog.Errorf("execute job %s plugin %s error: %s", pluginParams.JobId, pluginParams.PluginKey, err.Error())
		b.Callback(pluginParams.JobId, &utils.Response{Code: code.PluginError, Msg: err.Error()})
		return
	}
	b.Callback(pluginParams.JobId, &utils.Response{Code: code.Success, Data: result})
}

func (b *Plugins) Callback(jobId uint, resp *utils.Response) {
	klog.Infof("job=%d callback response: %v", jobId, resp)

	res := b.callback(serializers.PipelineCallbackSerializer{JobId: jobId, Result: resp})
	klog.Infof("job=%d callback to pipeline return: %v", jobId, res)
}
