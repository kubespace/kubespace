package resource

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/kubespace/kubespace/pkg/kubernetes/config"
	"github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog/v2"
	"sync"
)

var PodGVR = &schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "pods",
}

type Pod struct {
	*Resource
	execSession *sync.Map
}

func NewPod(config *config.KubeConfig) *Pod {
	p := &Pod{
		execSession: &sync.Map{},
	}
	p.Resource = NewResource(config, types.PodType, PodGVR, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.StdinAction:  p.ExecStdIn,
		types.DeleteAction: p.Delete,
	}
	return p
}

type BuildContainer struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Restarts int32  `json:"restarts"`
	Ready    bool   `json:"ready"`
}

type BuildPod struct {
	UID             string            `json:"uid"`
	Labels          map[string]string `json:"labels"`
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	Containers      []*BuildContainer `json:"containers"`
	InitContainers  []*BuildContainer `json:"init_containers"`
	Controlled      string            `json:"controlled"`
	ControlledName  string            `json:"controlled_name"`
	Qos             string            `json:"qos"`
	Created         metav1.Time       `json:"created"`
	Status          string            `json:"status"`
	Ip              string            `json:"ip"`
	NodeName        string            `json:"node_name"`
	ResourceVersion string            `json:"resource_version"`
	ContainerNum    int               `json:"containerNum"`
	Restarts        int32             `json:"restarts"`
}

func (p *Pod) ToBuildContainer(statuses []corev1.ContainerStatus, container *corev1.Container) *BuildContainer {
	bc := &BuildContainer{
		Name: container.Name,
	}
	for _, s := range statuses {
		if s.Name == container.Name {
			bc.Restarts = s.RestartCount
			if s.State.Running != nil {
				bc.Status = "running"
			} else if s.State.Terminated != nil {
				bc.Status = "terminated"
			} else if s.State.Waiting != nil {
				bc.Status = "waiting"
			}
			bc.Ready = s.Ready
			break
		}
	}
	return bc
}

func (p *Pod) ToBuildPod(pod *corev1.Pod) *BuildPod {
	if pod == nil {
		return nil
	}
	var restarts = int32(0)
	var containers []*BuildContainer
	for _, container := range pod.Spec.Containers {
		bc := p.ToBuildContainer(pod.Status.ContainerStatuses, &container)
		containers = append(containers, bc)
		if bc.Restarts > restarts {
			restarts = bc.Restarts
		}
	}
	cn := len(containers)
	var initContainers []*BuildContainer
	for _, container := range pod.Spec.InitContainers {
		bc := p.ToBuildContainer(pod.Status.InitContainerStatuses, &container)
		initContainers = append(initContainers, bc)
		if bc.Restarts > restarts {
			restarts = bc.Restarts
		}
	}
	cn += len(initContainers)
	var controlled = ""
	var controlledName = ""
	if len(pod.ObjectMeta.OwnerReferences) > 0 {
		controlled = pod.ObjectMeta.OwnerReferences[0].Kind
		controlledName = pod.ObjectMeta.OwnerReferences[0].Name
	}
	return &BuildPod{
		UID:             string(pod.UID),
		Labels:          pod.Labels,
		Name:            pod.Name,
		Namespace:       pod.Namespace,
		Containers:      containers,
		InitContainers:  initContainers,
		Controlled:      controlled,
		ControlledName:  controlledName,
		Qos:             string(pod.Status.QOSClass),
		Status:          string(pod.Status.Phase),
		Ip:              pod.Status.PodIP,
		Created:         pod.GetCreationTimestamp(),
		NodeName:        pod.Spec.NodeName,
		ResourceVersion: pod.ResourceVersion,
		ContainerNum:    cn,
		Restarts:        restarts,
	}
}

func (p *Pod) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	pod := &corev1.Pod{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, pod); err != nil {
		return nil, err
	}
	if len(query.Names) > 0 && !utils.Contains(query.Names, pod.Name) {
		return nil, nil
	}
	return p.ToBuildPod(pod), nil
}

type PodExecParams struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Container string `json:"container"`
	SessionId string `json:"session_id"`
	Rows      string `json:"rows"`
	Cols      string `json:"cols"`
}

// Exec Todo 添加注释
func (p *Pod) Exec(params interface{}, writer OutWriter) *utils.Response {
	var execParams PodExecParams
	if err := utils.ConvertTypeByJson(params, &execParams); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	execCmd := []string{"/bin/sh", "-c",
		fmt.Sprintf(`export LINES=%s; export COLUMNS=%s; 
	 TERM=xterm-256color; export TERM;
	 [ -x /bin/bash ] && ([ -x /usr/bin/script ] && /usr/bin/script -q -c \"/bin/bash\" /dev/null || exec /bin/bash) || exec /bin/sh`,
			execParams.Rows, execParams.Cols)}
	sshReq := p.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(execParams.Name).
		Namespace(execParams.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: execParams.Container,
			Command:   execCmd,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(p.client.RestConfig(), "POST", sshReq.URL())
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	handler := &streamHandler{
		resizeEvent: make(chan remotecommand.TerminalSize),
		stdinCh:     make(chan string),
		writer:      writer,
	}
	sessionId := execParams.SessionId
	p.execSession.Store(sessionId, handler)
	go func() {
		defer p.execSession.Delete(sessionId)
		defer handler.Close()
		if err := executor.Stream(remotecommand.StreamOptions{
			Stdin:             handler,
			Stdout:            handler,
			Stderr:            handler,
			TerminalSizeQueue: handler,
			Tty:               true,
		}); err != nil {
			klog.Errorf("exec pod container error session %s: %v", sessionId, err)
			handler.Write([]byte(err.Error()))
		}
		klog.Info("connection closed")
		handler.Write([]byte(fmt.Sprint("\n\rConnection closed")))
	}()
	return &utils.Response{Code: code.Success}
}

// todo 添加注释
type streamHandler struct {
	stdinCh     chan string
	resizeEvent chan remotecommand.TerminalSize
	writer      OutWriter
}

func (s *streamHandler) Read(p []byte) (size int, err error) {
	select {
	case inData, ok := <-s.stdinCh:
		if !ok {
			return 0, fmt.Errorf("stream closed")
		}
		d, err := base64.StdEncoding.DecodeString(inData)
		if err != nil {
			klog.Errorf("decode stream input data error: %s", err.Error())
		} else {
			size = len(d)
			copy(p, d)
		}
	case <-s.writer.StopCh():
		return 0, fmt.Errorf("writer closed")
	}
	return
}

func (s *streamHandler) Write(p []byte) (size int, err error) {
	copyData := make([]byte, len(p))
	copy(copyData, p)
	size = len(p)
	err = s.writer.Write(string(copyData))
	return
}

// Next executor回调获取web是否resize
func (s *streamHandler) Next() (size *remotecommand.TerminalSize) {
	ret := <-s.resizeEvent
	size = &ret
	return
}

func (s *streamHandler) Close() {
	s.writer.Close()
}

type StdInParams struct {
	SessionId string `json:"session_id"`
	Input     string `json:"input"`
	Width     uint16 `json:"width"`
	Height    uint16 `json:"height"`
}

// ExecStdIn todo 添加注释
func (p *Pod) ExecStdIn(params interface{}) *utils.Response {
	var inParams StdInParams
	if err := utils.ConvertTypeByJson(params, &inParams); err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	handlerObj, ok := p.execSession.Load(inParams.SessionId)
	if !ok {
		return &utils.Response{Code: code.RequestError, Msg: fmt.Sprintf("no open terminal")}
	}
	handler, ok := handlerObj.(*streamHandler)
	if !ok {
		return &utils.Response{Code: code.RequestError, Msg: fmt.Sprintf("open terminal error")}
	}
	if inParams.Width > 0 && inParams.Height > 0 {
		handler.resizeEvent <- remotecommand.TerminalSize{Width: inParams.Width, Height: inParams.Height}
	}
	handler.stdinCh <- inParams.Input
	return &utils.Response{Code: code.Success, Msg: "Success"}
}

type PodLogParams struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Container string `json:"container"`
}

func (p *Pod) Log(params interface{}, writer OutWriter) *utils.Response {
	var logParams PodLogParams
	if err := utils.ConvertTypeByJson(params, &logParams); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	tailLines := int64(100)
	podLogOpts := &corev1.PodLogOptions{
		Container: logParams.Container,
		Follow:    true,
		TailLines: &tailLines,
	}
	req := p.client.CoreV1().Pods(logParams.Namespace).GetLogs(logParams.Name, podLogOpts)
	logStream, err := req.Stream(context.Background())
	if err != nil {
		klog.Errorf("open pod %s container %s log stream error: %s", logParams.Name, logParams.Container, err.Error())
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	handler := &streamHandler{writer: writer}
	go func() {
		if _, err = io.Copy(handler, logStream); err != nil {
			klog.Errorf("io copy log stream error: %s", err.Error())
			handler.Close()
		}
		klog.V(1).Infof("io copy stopped")
	}()
	go func() {
		select {
		case <-handler.writer.StopCh():
			klog.V(1).Infof("out stream writer stopped")
			logStream.Close()
		}
	}()
	return &utils.Response{Code: code.Success}
}
