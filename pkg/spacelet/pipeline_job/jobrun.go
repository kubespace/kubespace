package pipeline_job

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	corerrors "github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/service/pipeline/job_runner"
	"github.com/kubespace/kubespace/pkg/service/pipeline/job_runner/plugins"
	"github.com/kubespace/kubespace/pkg/service/pipeline/schemas"
	"github.com/kubespace/kubespace/pkg/third/httpclient"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
	"os"
	"path"
	"strconv"
)

const PipelineCallbackUri = "/api/v1/spacelet/pipeline/callback"

// SpaceletJobRun Spacelet流水线任务插件执行处理
type SpaceletJobRun struct {
	jobRunner job_runner.JobRunner
	plugins   map[string]plugins.ExecutorFactory
	// 任务执行所在目录
	dataDir string
	// 对server进行调用
	client *httpclient.HttpClient
}

func NewSpaceletJobRun(dataDir string, client *httpclient.HttpClient) *SpaceletJobRun {
	p := &SpaceletJobRun{
		jobRunner: job_runner.NewJobRunner(),
		plugins:   make(map[string]plugins.ExecutorFactory),
		dataDir:   dataDir,
		client:    client,
	}
	p.plugins[types.BuiltinPluginBuildCodeToImage] = plugins.CodeBuilderPlugin{}
	p.plugins[types.BuiltinPluginExecuteShell] = plugins.ExecShellPlugin{}
	p.plugins[types.BuiltinPluginRelease] = plugins.NewReleasePlugin(client)
	return p
}

// GetRootDir 获取pipeline任务执行的根目录，如果目录不存在则创建一个
func (b *SpaceletJobRun) GetRootDir(jobId uint) (string, error) {
	rootPath := path.Join(b.dataDir, "pipeline", strconv.Itoa(int(jobId)))
	if err := os.MkdirAll(rootPath, 0755); err != nil {
		return "", err
	}
	return rootPath, nil
}

// getKubeSpaceFile 获取执行任务的kubespace生成的文件
func (b *SpaceletJobRun) getKubeSpaceFile(jobId uint, filename string) (string, error) {
	rootPath, err := b.GetRootDir(jobId)
	if err != nil {
		return "", err
	}
	ksPath := path.Join(rootPath, ".kubespace")
	if err = os.MkdirAll(ksPath, 0755); err != nil {
		return "", err
	}
	return path.Join(ksPath, filename), err
}

// GetJobLogFile 获取任务执行的日志文件
func (b *SpaceletJobRun) GetJobLogFile(jobId uint) (string, error) {
	return b.getKubeSpaceFile(jobId, "log")
}

// GetJobStatusFile 获取任务执行的结果以及状态文件
func (b *SpaceletJobRun) GetJobStatusFile(jobId uint) (string, error) {
	return b.getKubeSpaceFile(jobId, "status")
}

func (b *SpaceletJobRun) Cancel(jobId uint) error {
	klog.Infof("cancel job id=%d", jobId)
	return b.jobRunner.Cancel(jobId)
}

// Execute 执行任务插件，任务开启一个协程后台执行，该方法立即返回，后续的任务执行状态通过回调接口上报
func (b *SpaceletJobRun) Execute(jobId uint, pluginKey string, params map[string]interface{}) (resp *utils.Response) {
	if pluginKey == "" {
		return &utils.Response{Code: code.PluginError, Msg: "not found plugin key parameter"}
	}
	// 任务插件执行器
	executorF, ok := b.plugins[pluginKey]
	if !ok {
		return &utils.Response{Code: code.PluginError, Msg: "not found plugin executor: " + pluginKey}
	}
	rootDir, err := b.GetRootDir(jobId)
	if err != nil {
		return &utils.Response{Code: code.PluginError, Msg: "get job root dir error: " + err.Error()}
	}
	// 任务日志文件
	logFile, err := b.GetJobLogFile(jobId)
	if err != nil {
		return &utils.Response{Code: code.PluginError, Msg: "get job log file error: " + err.Error()}
	}
	// 设置任务执行的日志
	fileLogger, err := NewFileLogger(jobId, logFile)
	if err != nil {
		return &utils.Response{Code: code.PluginError, Msg: "get job logger error: " + err.Error()}
	}

	// 任务状态文件
	statusFile, err := b.GetJobStatusFile(jobId)
	if err != nil {
		return &utils.Response{Code: code.PluginError, Msg: "get job status file error: " + err.Error()}
	}
	jobStatus := NewJobStatus(statusFile)

	// 设置任务状态为doing，写入状态文件
	if err = jobStatus.Set(&StatusResult{Status: types.PipelineStatusDoing}); err != nil {
		return &utils.Response{Code: code.PluginError, Msg: "set status error: " + err.Error()}
	}

	jobParams := plugins.NewExecutorParams(jobId, pluginKey, rootDir, params, fileLogger)

	// 后台执行任务，执行完成后将状态以及结果保存到文件，并回调任务完成接口
	go b.execute(executorF, jobParams, jobStatus)

	return &utils.Response{Code: code.Success}
}

func (b *SpaceletJobRun) execute(executorF plugins.ExecutorFactory, params *plugins.ExecutorParams, jobStatus *JobStatus) {
	// 执行任务
	hostname, _ := os.Hostname()
	params.Logger.Log("current node: %s", hostname)
	result, err := b.jobRunner.Execute(executorF, params)
	if errors.Is(err, types.JobAlreadyRunningError) {
		return
	}
	var status = types.PipelineStatusOK
	var resp *utils.Response
	if err != nil {
		// 执行失败
		status = types.PipelineStatusError
		klog.Errorf("execute job=%d error: %s", params.JobId, err.Error())
		resp = utils.NewResponseWithError(corerrors.New(code.PluginError, err))
	} else {
		resp = utils.NewResponseOk(result)
	}
	// 将执行结果记录到当前任务的状态文件中
	if err = jobStatus.Set(&StatusResult{Status: status, Result: resp}); err != nil {
		klog.Errorf("set job=%d status result error: %s", params.JobId, err.Error())
	}
	callbackParams := &schemas.JobCallbackParams{JobId: params.JobId, Status: status}
	// 任务执行完成之后回调，回调失败不影响，pipeline controller有轮询机制定期查询任务状态
	if _, err = b.client.Post(PipelineCallbackUri, callbackParams, nil, httpclient.RequestOptions{}); err != nil {
		klog.Errorf("callback job=%d error: %s", params.JobId, err.Error())
	}
}

// StatusLog 任务状态以及日志
type StatusLog struct {
	StatusResult *StatusResult `json:"status"`
	Log          string        `json:"log"`
}

// GetStatusLog 从pipelline当前任务目录中的status文件获取状态信息
func (b *SpaceletJobRun) GetStatusLog(jobId uint, withLog bool) (*StatusLog, error) {
	// 当前任务的状态文件
	statusFile, err := b.GetJobStatusFile(jobId)
	if err != nil {
		return nil, err
	}
	jobStatus := NewJobStatus(statusFile)
	statusResult, err := jobStatus.Get()
	if err != nil {
		return nil, err
	}
	statusLog := &StatusLog{StatusResult: statusResult}
	if withLog {
		// 获取log文件日志内容
		logFile, err := b.GetJobLogFile(jobId)
		if err != nil {
			return nil, err
		}
		log, err := os.ReadFile(logFile)
		if err != nil {
			return nil, err
		}
		statusLog.Log = string(log)
	}
	return statusLog, nil
}

// Cleanup 清理任务目录
func (b *SpaceletJobRun) Cleanup(jobId uint) error {
	rootDir, err := b.GetRootDir(jobId)
	if err != nil {
		return err
	}
	klog.Infof("remove job id=%d root dir=%s", jobId, rootDir)
	return os.RemoveAll(rootDir)
}

// FileLogger 任务执行时日志写入文件
type FileLogger struct {
	jobId uint
	*os.File
}

func NewFileLogger(jobId uint, logFilePath string) (*FileLogger, error) {
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	return &FileLogger{
		jobId: jobId,
		File:  logFile,
	}, nil
}

// Log 日志存储到文件
func (l *FileLogger) Log(format string, a ...interface{}) {
	if _, err := fmt.Fprintf(l.File, format+"\n", a...); err != nil {
		klog.Errorf("job=%d write log error: %v", l.jobId, err)
	}
}

// Reset 从头日志重写
func (l *FileLogger) Reset(format string, a ...interface{}) {
	if _, err := l.WriteAt([]byte(fmt.Sprintf(format, a)), 0); err != nil {
		klog.Errorf("job=%d reset write log error: %v", l.jobId, err)
	}
}

func (l *FileLogger) Close() error {
	return l.File.Close()
}

// StatusResult 任务状态以及执行结果存储查询
type StatusResult struct {
	Status string          `json:"status"`
	Result *utils.Response `json:"result"`
}

type JobStatus struct {
	filePath string
}

func NewJobStatus(filePath string) *JobStatus {
	return &JobStatus{filePath: filePath}
}

// Set 将任务状态以及执行结果写入到文件
func (s *JobStatus) Set(sr *StatusResult) error {
	statusBytes, err := json.Marshal(sr)
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, statusBytes, 0644)
}

// Get 从文件获取任务状态以及执行结果
func (s *JobStatus) Get() (*StatusResult, error) {
	// 读取文件内容
	statusBytes, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, err
	}
	var sr StatusResult
	if err = json.Unmarshal(statusBytes, &sr); err != nil {
		return nil, err
	}
	return &sr, nil
}
