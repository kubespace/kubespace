package plugins

import (
	"encoding/json"
	"fmt"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/service/pipeline/schemas"
	"github.com/kubespace/kubespace/pkg/third/httpclient"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"io/ioutil"
	"k8s.io/klog/v2"
	"os"
	"path"
	"runtime"
	"strconv"
	"sync"
)

const PipelineCallbackUri = "/api/v1/spacelet/pipeline/callback"

type PluginExecutor interface {
	Execute(params *PluginParams) (interface{}, error)
}

// PluginParams 任务执行的参数
type PluginParams struct {
	JobId     uint
	PluginKey string
	RootDir   string
	Params    map[string]interface{}
	Logger    *JobLogger
}

type Plugins struct {
	plugins     map[string]PluginExecutor
	models      *model.Models
	dataDir     string
	client      *httpclient.HttpClient
	mu          *sync.Mutex
	runningJobs map[uint]*pluginExec
}

func NewPlugins(dataDir string, client *httpclient.HttpClient) *Plugins {
	p := &Plugins{
		plugins:     make(map[string]PluginExecutor),
		dataDir:     dataDir,
		client:      client,
		mu:          &sync.Mutex{},
		runningJobs: make(map[uint]*pluginExec),
	}
	p.plugins[types.BuiltinPluginBuildCodeToImage] = CodeBuilderPlugin{}
	p.plugins[types.BuiltinPluginExecuteShell] = ExecShellPlugin{}
	p.plugins[types.BuiltinPluginRelease] = ReleaserPlugin{client: client}
	return p
}

// GetRootDir 获取pipeline任务执行的根目录，如果目录不存在则创建一个
func (b *Plugins) GetRootDir(jobId uint) (string, error) {
	rootPath := path.Join(b.dataDir, "pipeline", strconv.Itoa(int(jobId)))
	if err := os.MkdirAll(rootPath, 0755); err != nil {
		return "", err
	}
	return rootPath, nil
}

// getKubeSpaceFile 获取执行任务的kubespace生成的文件
func (b *Plugins) getKubeSpaceFile(jobId uint, filename string) (string, error) {
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
func (b *Plugins) GetJobLogFile(jobId uint) (string, error) {
	return b.getKubeSpaceFile(jobId, "log")
}

// GetJobStatusFile 获取任务执行的结果以及状态文件
func (b *Plugins) GetJobStatusFile(jobId uint) (string, error) {
	return b.getKubeSpaceFile(jobId, "status")
}

// RecordJob 记录job当前执行的实例，如果已经正在执行了，返回false
func (b *Plugins) RecordJob(jobId uint, exec *pluginExec) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.runningJobs[jobId]; ok {
		return false
	}
	b.runningJobs[jobId] = exec
	return true
}

// RemoveJob 执行完成后删除job记录
func (b *Plugins) RemoveJob(jobId uint) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.runningJobs, jobId)
}

// Execute 执行任务插件，任务开启一个协程后台执行，该方法立即返回，后续的任务执行状态通过回调接口上报
func (b *Plugins) Execute(pluginParams *PluginParams) (resp *utils.Response) {
	var err error
	pluginParams.RootDir, err = b.GetRootDir(pluginParams.JobId)
	if err != nil {
		return &utils.Response{Code: code.PluginError, Msg: "get job root dir error: " + err.Error()}
	}
	if pluginParams.PluginKey == "" {
		return &utils.Response{Code: code.PluginError, Msg: "not found plugin key parameter"}
	}
	// 任务插件执行器
	executor, ok := b.plugins[pluginParams.PluginKey]
	if !ok {
		return &utils.Response{Code: code.PluginError, Msg: "not found plugin executor: " + pluginParams.PluginKey}
	}
	// 任务日志文件
	logFile, err := b.GetJobLogFile(pluginParams.JobId)
	if err != nil {
		return &utils.Response{Code: code.PluginError, Msg: "get job log file error: " + err.Error()}
	}
	// 设置任务执行的日志
	if pluginParams.Logger, err = NewPluginLogger(pluginParams.JobId, logFile); err != nil {
		return &utils.Response{Code: code.PluginError, Msg: "get job logger error: " + err.Error()}
	}

	// 任务状态文件
	statusFile, err := b.GetJobStatusFile(pluginParams.JobId)
	if err != nil {
		return &utils.Response{Code: code.PluginError, Msg: "get job status file error: " + err.Error()}
	}
	jobStatus := NewJobStatus(statusFile)

	// 设置任务状态为doing，写入状态文件
	if err = jobStatus.Set(&StatusResult{Status: types.PipelineStatusDoing}); err != nil {
		return &utils.Response{Code: code.PluginError, Msg: "set status error: " + err.Error()}
	}

	e := &pluginExec{
		jobStatus: jobStatus,
		params:    pluginParams,
		executor:  executor,
		client:    b.client,
	}
	if b.RecordJob(pluginParams.JobId, e) {
		go func() {
			// 执行完成关闭日志文件句柄
			defer e.params.Logger.Close()
			// 执行完成删除job执行记录
			defer b.RemoveJob(pluginParams.JobId)
			e.Execute()
		}()
	} else {
		klog.Infof("job=%d already running", pluginParams.JobId)
	}

	return &utils.Response{Code: code.Success}
}

// GetStatusLog 从pipelline当前任务目录中的status文件获取状态信息
func (b *Plugins) GetStatusLog(jobId uint, withLog bool) (*StatusLog, error) {
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
		log, err := ioutil.ReadFile(logFile)
		if err != nil {
			return nil, err
		}
		statusLog.Log = string(log)
	}
	return statusLog, nil
}

// Cleanup 清理任务目录
func (b *Plugins) Cleanup(jobId uint) error {
	rootDir, err := b.GetRootDir(jobId)
	if err != nil {
		return err
	}
	klog.Infof("remove job id=%d root dir=%s", jobId, rootDir)
	return os.RemoveAll(rootDir)
}

// 每个任务一个执行实例
type pluginExec struct {
	jobStatus *JobStatus
	params    *PluginParams
	executor  PluginExecutor
	client    *httpclient.HttpClient
}

func (e *pluginExec) Execute() {
	// 执行任务
	result, err := e.execute()
	var status = types.PipelineStatusOK
	var resp *utils.Response
	if err != nil {
		// 执行失败
		status = types.PipelineStatusError
		klog.Errorf("execute job=%d error: %s", e.params.JobId, err.Error())
		resp = &utils.Response{Code: code.PluginError, Msg: err.Error()}
	} else {
		resp = &utils.Response{Code: code.Success, Data: result}
	}
	// 将执行结果记录到当前任务的状态文件中
	if err = e.jobStatus.Set(&StatusResult{Status: status, Result: resp}); err != nil {
		klog.Errorf("set job=%d status result error: %s", e.params.JobId, err.Error())
	}
	params := &schemas.JobCallbackParams{JobId: e.params.JobId, Status: status}
	// 任务执行完成之后回调，失败不影响，controller-manager有轮询机制定期查询任务状态
	if _, err = e.client.Post(PipelineCallbackUri, params, nil, httpclient.RequestOptions{}); err != nil {
		klog.Errorf("callback job=%d error: %s", e.params.JobId, err.Error())
	}
}

// 执行任务插件，抓取panic异常信息
func (e *pluginExec) execute() (result interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			klog.Error("error: ", r)
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			klog.Errorf("==> %s", string(buf[:n]))
			e.params.Logger.Log("error: %v ==> %s\n", r, string(buf[:n]))
			err = fmt.Errorf("%v", r)
		}
	}()
	return e.executor.Execute(e.params)
}

// JobLogger 任务执行时日志写入文件
type JobLogger struct {
	jobId uint
	*os.File
}

func NewPluginLogger(jobId uint, logFilePath string) (*JobLogger, error) {
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	return &JobLogger{
		jobId: jobId,
		File:  logFile,
	}, nil
}

func (l *JobLogger) Log(format string, a ...interface{}) {
	_, err := fmt.Fprintf(l.File, format+"\n", a...)
	if err != nil {
		klog.Errorf("job=%d write log error: %v", l.jobId, err)
	}
}

func (l *JobLogger) Close() error {
	return l.File.Close()
}

// StatusLog 任务状态以及日志
type StatusLog struct {
	StatusResult *StatusResult `json:"status"`
	Log          string        `json:"log"`
}

// StatusResult 任务状态以及执行结果
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
	return ioutil.WriteFile(s.filePath, statusBytes, 0644)
}

// Get 从文件获取任务状态以及执行结果
func (s *JobStatus) Get() (*StatusResult, error) {

	// 读取文件内容
	statusBytes, err := ioutil.ReadFile(s.filePath)
	if err != nil {
		return nil, err
	}
	var sr StatusResult
	if err = json.Unmarshal(statusBytes, &sr); err != nil {
		return nil, err
	}
	return &sr, nil
}
