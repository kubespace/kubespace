package plugins

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kubespace/kubespace/pkg/utils"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog/v2"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

const (
	ResourceTypeImage = "image"
	ResourceTypeHost  = "host"
)

type ExecShellPlugin struct{}

func (b ExecShellPlugin) Executor(params *ExecutorParams) (Executor, error) {
	return newExecShellExecutor(params)
}

type ExecShellParams struct {
	JobId uint `json:"job_id"`

	Resource PipelineResource       `json:"resource"`
	Port     string                 `json:"port"`
	Shell    string                 `json:"shell"`
	Script   string                 `json:"script"`
	Env      map[string]interface{} `json:"env"`
}

type execShellExecutor struct {
	Logger
	Params     *ExecShellParams
	Result     map[string]interface{} `json:"env"`
	rootDir    string
	cancelFunc context.CancelFunc
	ctx        context.Context
	canceled   bool
	sshSession *ssh.Session
	cmd        *exec.Cmd
}

func newExecShellExecutor(params *ExecutorParams) (*execShellExecutor, error) {
	var shellParams ExecShellParams
	if err := utils.ConvertTypeByJson(params.Params, &shellParams); err != nil {
		return nil, err
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	execPlugin := &execShellExecutor{
		Params:     &shellParams,
		Logger:     params.Logger,
		Result:     make(map[string]interface{}),
		rootDir:    params.RootDir,
		cancelFunc: cancelFunc,
		ctx:        ctx,
	}

	return execPlugin, nil
}

func (b *execShellExecutor) Execute() (interface{}, error) {
	if b.Params.Resource.Type != "" && b.Params.Resource.Value == "" {
		return nil, fmt.Errorf("执行脚本目标资源参数为空，请检查流水线配置")
	}
	var err error
	if b.Params.Resource.Type == ResourceTypeImage {
		err = b.execImage()
	} else if b.Params.Resource.Type == ResourceTypeHost {
		err = b.execSsh()
	} else {
		err = b.execCmd()
	}
	if err != nil {
		return nil, err
	}
	return b.Result, nil
}

func (b *execShellExecutor) Cancel() error {
	b.canceled = true
	b.cancelFunc()
	if b.sshSession != nil {
		b.sshSession.Close()
	}
	if b.cmd != nil {
		b.cmd.Cancel()
	}
	return nil
}

func (b *execShellExecutor) execCmd() error {
	shell := b.Params.Shell
	if shell == "" {
		shell = "sh"
	}

	cmd := exec.CommandContext(b.ctx, shell, "-xc", b.Params.Script)
	stdin := bytes.NewBuffer(nil)
	cmd.Stdin = stdin
	cmd.Stdout = b.Logger
	cmd.Stderr = b.Logger
	cmd.Dir = b.rootDir
	var envs []string
	for name, val := range b.Params.Env {
		envs = append(envs, fmt.Sprintf("%s=%v", name, val))
	}
	cmd.Env = envs
	b.cmd = cmd
	if err := cmd.Run(); err != nil {
		b.Log(err.Error())
		return fmt.Errorf("execute error: %v", err)
	}

	return b.getOutput()
}

func (b *execShellExecutor) execImage() error {
	image := b.Params.Resource.Value
	shell := b.Params.Shell
	if shell == "" {
		shell = "sh"
	}

	// 脚本写入文件
	scriptFileName := ".script.sh"
	scriptFile := filepath.Join(b.rootDir, scriptFileName)
	if err := os.WriteFile(scriptFile, []byte(b.Params.Script), 0644); err != nil {
		b.Log("写入脚本错误：%v", err)
		klog.Errorf("job=%d write build error: %v", b.Params.JobId, err)
		return err
	}

	var envs []string
	for name, val := range b.Params.Env {
		envs = append(envs, fmt.Sprintf("%s='%v'", name, val))
	}
	envs = append(envs, fmt.Sprintf("WORKDIR='/pipeline'"))
	env := strings.Join(envs, " ")

	dockerRunCmd := fmt.Sprintf("docker run --net=host --rm -i -v %s:/pipeline -w /pipeline --entrypoint sh %s -c \"%s %s -x %s 2>&1\"", b.rootDir, image, env, shell, scriptFileName)
	klog.Infof("job=%d code build cmd: %s", b.Params.JobId, dockerRunCmd)
	cmd := exec.CommandContext(b.ctx, "bash", "-c", dockerRunCmd)
	cmd.Stdout = b.Logger
	cmd.Stderr = b.Logger

	if err := cmd.Run(); err != nil {
		klog.Errorf("job=%d execute error: %v", b.Params.JobId, err)
		b.Log(err.Error())
		return fmt.Errorf("execute error: %v", err)
	}
	return b.getOutput()
}

func (b *execShellExecutor) getOutput() error {
	// 读取脚本输出内容
	outputBytes, err := os.ReadFile(path.Join(b.rootDir, "output"))
	if err != nil {
		if !os.IsNotExist(err) {
			b.Log("read output error: %s", err.Error())
			return err
		}
	}
	outEnvStr := string(outputBytes)
	b.Log("output:\n%s", outEnvStr)
	if outEnvStr != "" {
		for _, line := range strings.Split(outEnvStr, "\n") {
			if strings.Contains(line, "=") {
				splits := strings.SplitN(line, "=", 2)
				key := splits[0]
				value := splits[1]
				b.Result[key] = value
			}
		}
	}

	return nil
}

func (b *execShellExecutor) execSsh() error {
	// 建立SSH客户端连接
	host := b.Params.Resource.Value
	if b.Params.Port != "" {
		host += ":" + b.Params.Port
	} else {
		host += ":22"
	}
	var auth ssh.AuthMethod
	if b.Params.Resource.Secret.Type == "key" {
		signer, err := ssh.ParsePrivateKey([]byte(b.Params.Resource.Secret.PrivateKey))
		if err != nil {
			b.Log("parse ssh public key error: %s", err.Error())
			return err
		}
		auth = ssh.PublicKeys(signer)
	} else if b.Params.Resource.Secret.Type == "password" {
		auth = ssh.Password(b.Params.Resource.Secret.Password)
	}
	client, err := ssh.Dial("tcp", host, &ssh.ClientConfig{
		User:            b.Params.Resource.Secret.User,
		Auth:            []ssh.AuthMethod{auth},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		b.Log("ssh host %s error: %s", host, err.Error())
		return err
	}
	b.Log("连接主机%s成功", host)

	// 建立新会话
	session, err := client.NewSession()
	if err != nil {
		b.Log("ssh host %s new session error: %s", host, err.Error())
		return err
	}
	defer session.Close()
	if err = session.RequestPty("vt100", 80, 25, ssh.TerminalModes{}); err != nil {
		fmt.Printf("Failed to request for pseudo terminal.err: %s\n", err.Error())
		return err
	}
	b.sshSession = session
	b.Log("建立session成功，开始执行脚本")
	session.Stdout = b.Logger
	var envs []string
	for name, val := range b.Params.Env {
		envs = append(envs, fmt.Sprintf("%s='%v'", name, val))
	}
	workDir := fmt.Sprintf("/tmp/kubespace/pipeline/%d", b.Params.JobId)
	envs = append(envs, fmt.Sprintf("WORKDIR='%s'", workDir))
	env := strings.Join(envs, " ")

	output := fmt.Sprintf("%s/output", workDir)
	cmd := fmt.Sprintf("mkdir -p %s && cd %s && rm -rf %s && %s bash -cx '%s' 2>&1", workDir, workDir, output, env, b.Params.Script)
	err = session.Run(cmd)
	if err != nil {
		b.Log("执行脚本失败: %s", err.Error())
		return err
	}

	// 建立新会话
	newSession, err := client.NewSession()
	if err != nil {
		b.Log("ssh host %s new session error: %s", host, err.Error())
		return err
	}
	defer newSession.Close()
	buffer := new(bytes.Buffer)
	newSession.Stdout = buffer
	// 获取执行脚本输出内容
	cmd = fmt.Sprintf("bash -c '[[ -f %s ]] && cat %s; rm -rf %s'", output, output, workDir)
	err = newSession.Run(cmd)
	if err != nil {
		b.Log("获取脚本输出%s失败: %s", output, err.Error())
		return err
	}
	outEnvStr := buffer.String()
	b.Log("output:\n%s", outEnvStr)
	if outEnvStr != "" {
		for _, line := range strings.Split(outEnvStr, "\n") {
			if strings.Contains(line, "=") {
				splits := strings.SplitN(line, "=", 2)
				key := splits[0]
				value := splits[1]
				b.Result[key] = value
			}
		}
	}

	return nil
}
