package plugins

import "io"

// ExecutorFactory 流水线任务插件执行器工厂，不同流水线任务插件产生不同的流水线执行器
type ExecutorFactory interface {
	// Executor 创建任务插件执行器
	Executor(params *ExecutorParams) (Executor, error)
}

// Executor 流水线任务插件执行处理器
type Executor interface {
	// Execute 任务插件执行处理
	Execute() (interface{}, error)

	// Cancel 取消执行
	Cancel() error
}

// ExecutorParams 任务插件执行参数
type ExecutorParams struct {
	JobId     uint
	PluginKey string
	RootDir   string
	Params    map[string]interface{}
	Logger    Logger
}

func NewExecutorParams(jobId uint, pluginKey, rootDir string, params map[string]interface{}, logger Logger) *ExecutorParams {
	return &ExecutorParams{
		JobId:     jobId,
		PluginKey: pluginKey,
		RootDir:   rootDir,
		Params:    params,
		Logger:    logger,
	}
}

// Logger 任务插件执行时调用的日志存储接口
// 当前存在两个日志存储
// 1. 在pipeline controller执行任务日志存储到db中
// 2. 在spacelet执行任务日志存储到文件中
type Logger interface {
	io.Writer

	// Log 日志追加写入
	Log(format string, a ...interface{})

	// Reset 日志覆盖重写
	Reset(format string, a ...interface{})

	Close() error
}

type stepFunc func() error
