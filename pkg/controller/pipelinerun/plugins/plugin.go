package plugins

import (
	"bytes"
	"fmt"
	"github.com/kubespace/kubespace/pkg/informer"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/service/cluster"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"k8s.io/klog/v2"
	"runtime"
	"time"
)

type PluginExecutor interface {
	Execute(params *PluginParams) (interface{}, error)
}

type PluginParams struct {
	JobId     uint
	PluginKey string
	DataDir   string
	Params    map[string]interface{}
	Logger    *PluginLogger
}

type Plugins struct {
	plugins map[string]PluginExecutor
	models  *model.Models
	dataDir string
}

func NewPlugins(models *model.Models, kubeClient *cluster.KubeClient, informerFactory informer.Factory) *Plugins {
	p := &Plugins{
		plugins: make(map[string]PluginExecutor),
		models:  models,
	}
	spacelet := SpaceletJob{models: models, informerFactory: informerFactory}
	// 通过spacelet代理执行
	p.plugins[types.BuiltinPluginBuildCodeToImage] = spacelet
	p.plugins[types.BuiltinPluginExecuteShell] = spacelet
	p.plugins[types.BuiltinPluginRelease] = spacelet

	//p.plugins[types.BuiltinPluginExecuteShell] = ExecShellPlugin{}
	//p.plugins[types.BuiltinPluginRelease] = ReleaserPlugin{Models: models}
	p.plugins[types.BuiltinPluginUpgradeApp] = UpgradeAppPlugin{Models: models, KubeClient: kubeClient}
	p.plugins[types.BuiltinPluginDeployK8s] = DeployK8sPlugin{Models: models, KubeClient: kubeClient}
	return p
}

func (b *Plugins) Execute(pluginParams *PluginParams) (resp *utils.Response) {
	defer func() {
		if err := recover(); err != nil {
			klog.Error("error: ", err)
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			klog.Errorf("==> %s\n", string(buf[:n]))
			pluginParams.Logger.Log("==> %s\n", string(buf[:n]))
			resp = &utils.Response{Code: code.UnknownError, Msg: fmt.Sprintf("%v", err)}
		}
	}()
	if pluginParams.PluginKey == "" {
		return &utils.Response{Code: code.PluginError, Msg: "not found plugin key parameter"}
	}
	executor, ok := b.plugins[pluginParams.PluginKey]
	if !ok {
		return &utils.Response{Code: code.PluginError, Msg: "not found plugin executor: " + pluginParams.PluginKey}
	}

	logger := NewPluginLogger(b.models, pluginParams.JobId)
	pluginParams.Logger = logger
	go logger.FlushLogToDB()
	defer close(logger.closeCh)

	result, err := executor.Execute(pluginParams)
	if err != nil {
		klog.Errorf("execute job %d plugin %s error: %s", pluginParams.JobId, pluginParams.PluginKey, err.Error())
		return &utils.Response{Code: code.PluginError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success, Data: result}
}

type PluginLogger struct {
	*bytes.Buffer
	jobId   uint
	models  *model.Models
	closeCh chan struct{}
}

func NewPluginLogger(models *model.Models, jobId uint) *PluginLogger {
	return &PluginLogger{
		Buffer:  &bytes.Buffer{},
		jobId:   jobId,
		models:  models,
		closeCh: make(chan struct{}),
	}
}

func (l *PluginLogger) Log(format string, a ...interface{}) {
	_, err := l.Buffer.WriteString(fmt.Sprintf(format+"\n", a...))
	if err != nil {
		klog.Errorf("write job %s log to buffer error: %s", l.jobId, err.Error())
	}
}

func (l *PluginLogger) ResetWrite(format string, a ...interface{}) {
	l.Buffer.Reset()
	l.Log(format, a...)
}

func (l *PluginLogger) FlushLogToDB() {
	tick := time.NewTicker(5 * time.Second)
	logLens := 0
	for {
		logLens = l.Len()
		select {
		case <-l.closeCh:
			err := l.models.PipelineJobLogManager.UpdateLog(l.jobId, l.Buffer.String())
			if err != nil {
				klog.Errorf("update job %s log error: %s", l.jobId, err.Error())
			}
			return
		case <-tick.C:
			if logLens == l.Len() || l.Len() == 0 {
				continue
			}
			err := l.models.PipelineJobLogManager.UpdateLog(l.jobId, l.Buffer.String())
			if err != nil {
				klog.Errorf("update job %s log error: %s", l.jobId, err.Error())
			}
		}
	}
}
