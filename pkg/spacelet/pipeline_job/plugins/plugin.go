package plugins

import (
	"encoding/json"
	"fmt"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
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

type PluginExecutor interface {
	Execute(params *PluginParams) (interface{}, error)
}

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
	mu       *sync.Mutex
	filePath string
}

// Set 将任务状态以及执行结果写入到文件
func (s *JobStatus) Set(sr *StatusResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	statusBytes, err := json.Marshal(sr)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(s.filePath, statusBytes, 0644)
}

// Get 从文件获取任务状态以及执行结果
func (s *JobStatus) Get() (*StatusResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

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

type PluginParams struct {
	JobId     uint
	PluginKey string
	RootDir   string
	Params    map[string]interface{}
	Logger    *JobLogger
}

type Plugins struct {
	plugins map[string]PluginExecutor
	models  *model.Models
	dataDir string
}

func NewPlugins(dataDir string) *Plugins {
	p := &Plugins{
		plugins: make(map[string]PluginExecutor),
		dataDir: dataDir,
	}
	p.plugins[types.BuiltinPluginBuildCodeToImage] = CodeBuilderPlugin{}
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

// Execute 执行任务插件，任务开启一个goroutinue后台执行，该方法立即返回
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
	jobStatus := &JobStatus{filePath: statusFile}
	// 设置任务状态为doing，写入状态文件
	if err = jobStatus.Set(&StatusResult{Status: types.PipelineStatusDoing}); err != nil {
		return &utils.Response{Code: code.PluginError, Msg: "write status error: " + err.Error()}
	}

	e := pluginExec{
		jobStatus: jobStatus,
		params:    pluginParams,
		executor:  executor,
	}
	// 后台执行任务
	go e.Execute()

	return &utils.Response{Code: code.Success}
}

func (b *Plugins) GetStatusLog(jobId uint, withLog bool) (*StatusLog, error) {
	// 任务状态文件
	statusFile, err := b.GetJobStatusFile(jobId)
	if err != nil {
		return nil, err
	}
	jobStatus := &JobStatus{filePath: statusFile}
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
	return os.RemoveAll(rootDir)
}

type pluginExec struct {
	jobStatus *JobStatus
	params    *PluginParams
	executor  PluginExecutor
}

func (e *pluginExec) Execute() {
	defer e.params.Logger.Close()
	var resp *utils.Response
	result, err := e.execute()
	if err != nil {
		klog.Errorf("execute job=%d error: %s", e.params.JobId, err.Error())
		resp = &utils.Response{Code: code.PluginError, Msg: err.Error()}
	} else {
		resp = &utils.Response{Code: code.Success, Data: result}
	}
	if err = e.jobStatus.Set(&StatusResult{Status: types.PipelineStatusOK, Result: resp}); err != nil {
		klog.Errorf("set job=%d status result error: %s", e.params.JobId, err.Error())
	}
}

func (e *pluginExec) execute() (result interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			klog.Error("error: ", r)
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			klog.Errorf("==> %s\n", string(buf[:n]))
			e.params.Logger.Log("==> %s\n", string(buf[:n]))
			err = fmt.Errorf("%v", r)
		}
	}()
	return e.executor.Execute(e.params)
}
