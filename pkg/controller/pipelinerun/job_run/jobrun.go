package job_run

import (
	"bytes"
	"fmt"
	"github.com/kubespace/kubespace/pkg/informer"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/service/cluster"
	"github.com/kubespace/kubespace/pkg/service/pipeline/job_runner"
	"github.com/kubespace/kubespace/pkg/service/pipeline/job_runner/plugins"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"k8s.io/klog/v2"
	"time"
)

// JobRun pipeline controller流水线任务执行处理
type JobRun struct {
	jobRunner job_runner.JobRunner
	plugins   map[string]plugins.ExecutorFactory
	models    *model.Models
}

func NewJobRun(models *model.Models, kubeClient *cluster.KubeClient, informerFactory informer.Factory) *JobRun {
	spacelet := SpaceletJob{models: models, informerFactory: informerFactory}
	return &JobRun{
		jobRunner: job_runner.NewJobRunner(),
		models:    models,
		plugins: map[string]plugins.ExecutorFactory{
			// 比较消耗资源的任务，通过spacelet代理执行
			types.BuiltinPluginBuildCodeToImage: spacelet,
			types.BuiltinPluginExecuteShell:     spacelet,
			types.BuiltinPluginRelease:          spacelet,

			// 轻任务直接在controller内部执行
			types.BuiltinPluginDeployK8s:  plugins.DeployK8sPlugin{Models: models, KubeClient: kubeClient},
			types.BuiltinPluginUpgradeApp: plugins.UpgradeAppPlugin{Models: models, KubeClient: kubeClient},
		},
	}
}

func (b *JobRun) Execute(jobId uint, pluginKey string, params map[string]interface{}) (resp *utils.Response) {
	if pluginKey == "" {
		return &utils.Response{Code: code.PluginError, Msg: "not found plugin key parameter"}
	}
	executorF, ok := b.plugins[pluginKey]
	if !ok {
		return &utils.Response{Code: code.PluginError, Msg: "not found plugin executor: " + pluginKey}
	}

	logger := NewPluginLogger(b.models, jobId)
	// 定时刷新日志到数据库
	go logger.FlushLogToDB()

	jobParams := plugins.NewExecutorParams(jobId, pluginKey, "", params, logger)
	result, err := b.jobRunner.Execute(executorF, jobParams)

	if err != nil {
		klog.Errorf("execute job %d plugin %s error: %s", jobId, pluginKey, err.Error())
		return &utils.Response{Code: code.PluginError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success, Data: result}
}

func (b *JobRun) Cancel(jobId uint) error {
	return b.jobRunner.Cancel(jobId)
}

// PluginLogger 记录日志到数据库
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

func (l *PluginLogger) Close() error {
	close(l.closeCh)
	return nil
}

func (l *PluginLogger) Reset(format string, a ...interface{}) {
	l.Buffer.Reset()
	l.Log(format, a...)
}

// FlushLogToDB 定时刷新日志到数据库
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
