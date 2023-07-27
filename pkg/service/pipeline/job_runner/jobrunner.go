package job_runner

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/service/pipeline/job_runner/plugins"
	"k8s.io/klog/v2"
	"os"
	"runtime"
	"sync"
)

// JobRunner 流水线任务执行器
// 在pipeline controller与spacelet调用
type JobRunner interface {
	Execute(executorF plugins.ExecutorFactory, params *plugins.ExecutorParams) (interface{}, error)
	Cancel(jobId uint) error
}

type jobRunner struct {
	// 对runningJobs处理时加锁
	mu *sync.Mutex
	// 任务执行器当前正在执行的任务
	runningJobs map[uint]plugins.Executor
}

func NewJobRunner() JobRunner {
	return &jobRunner{
		mu:          &sync.Mutex{},
		runningJobs: make(map[uint]plugins.Executor),
	}
}

func (j *jobRunner) addJob(jobId uint, executor plugins.Executor) bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	if _, ok := j.runningJobs[jobId]; ok {
		// 如果已存在正在执行的任务，返回false
		return false
	}
	j.runningJobs[jobId] = executor
	return true
}

func (j *jobRunner) delJob(jobId uint) bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	if _, ok := j.runningJobs[jobId]; !ok {
		// 不存在正在执行的任务，返回false
		return false
	}
	delete(j.runningJobs, jobId)
	return true
}

func (j *jobRunner) Execute(executorF plugins.ExecutorFactory, params *plugins.ExecutorParams) (res interface{}, err error) {
	// 退出时关闭日志
	defer params.Logger.Close()

	hostname, _ := os.Hostname()
	params.Logger.Log("current node: %s", hostname)

	if params.JobId == 0 {
		return nil, fmt.Errorf("params error: job id is empty")
	}
	// 生成任务执行器
	executor, err := executorF.Executor(params)
	if err != nil {
		return nil, fmt.Errorf("create plugin executor error: %s", err.Error())
	}
	// 添加到当前正在执行的任务
	if !j.addJob(params.JobId, executor) {
		// 添加任务失败
		klog.Infof("job id=%d already has running", params.JobId)
		return nil, types.JobAlreadyRunningError
	}
	defer j.delJob(params.JobId)
	defer func() {
		if r := recover(); r != nil {
			klog.Error("error: ", r)
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			klog.Errorf("==> %s", string(buf[:n]))
			params.Logger.Log("error: %v ==> %s\n", r, string(buf[:n]))
			err = fmt.Errorf("%v", r)
		}
	}()
	return executor.Execute()
}

// Cancel 取消任务执行
func (j *jobRunner) Cancel(jobId uint) error {
	executor, ok := j.runningJobs[jobId]
	if !ok {
		return nil
	}
	return executor.Cancel()
}
