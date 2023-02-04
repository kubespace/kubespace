package resource

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kubespace/kubespace/pkg/kubernetes/config"
	"github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
	"strings"
	"sync"
	"time"
)

type Helm struct {
	*Resource
}

func NewHelm(conf *config.KubeConfig) *Helm {
	p := &Helm{
		Resource: NewResource(conf, "", nil, nil),
	}
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.CreateAction: p.Create,
		types.DeleteAction: p.Delete,
		types.UpdateAction: p.Update,
	}
	return p
}

func (h *Helm) actionConfig(namespace string) (*action.Configuration, error) {
	insecure := true
	configFlags := &genericclioptions.ConfigFlags{
		Insecure:    &insecure,
		Namespace:   &namespace,
		APIServer:   utils.StringPtr(h.client.RestConfig().Host),
		BearerToken: utils.StringPtr(h.client.RestConfig().BearerToken),
		WrapConfigFn: func(rc *rest.Config) *rest.Config {
			return h.client.RestConfig()
		},
	}
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(configFlags, namespace, "", klog.Infof); err != nil {
		return nil, err
	}
	return actionConfig, nil
}

type HelmListParams struct {
	Name       string   `json:"name"`
	Names      []string `json:"names"`
	Namespace  string   `json:"namespace"`
	WithStatus bool     `json:"with_status"`
}

func (h *Helm) List(params interface{}) *utils.Response {
	var query HelmListParams
	if err := utils.ConvertTypeByJson(params, &query); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	actionConfig, err := h.actionConfig(query.Namespace)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	client := action.NewList(actionConfig)
	results, err := client.Run()
	if err != nil {
		klog.Errorf("list helm error: %s", err)
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	var res []map[string]interface{}
	wg := &sync.WaitGroup{}
	for _, r := range results {
		if len(query.Names) > 0 && !utils.Contains(query.Names, r.Name) {
			continue
		}
		data := map[string]interface{}{
			"name":          r.Name,
			"namespace":     r.Namespace,
			"version":       r.Version,
			"status":        r.Info.Status,
			"chart_name":    r.Chart.Name() + "-" + r.Chart.Metadata.Version,
			"chart_version": r.Chart.Metadata.Version,
			"app_version":   r.Chart.AppVersion(),
			"last_deployed": r.Info.LastDeployed,
		}
		if query.WithStatus {
			wg.Add(1)
			go func(releaseDetail *release.Release) {
				defer wg.Done()
				state, _ := h.GetReleaseRuntimeState(releaseDetail, false)
				data["runtime_status"] = state.Status
			}(r)
		}
		res = append(res, data)
	}
	wg.Wait()
	return &utils.Response{Code: code.Success, Msg: "Success", Data: res}
}

type HelmGetParams struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	WithResource bool   `json:"with_resource"`
}

func (h *Helm) Get(params interface{}) *utils.Response {
	queryParams := &HelmGetParams{}
	if err := utils.ConvertTypeByJson(params, queryParams); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	if queryParams.Name == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "Name is blank"}
	}
	if queryParams.Namespace == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "Namespace is blank"}
	}
	actionConfig, err := h.actionConfig(queryParams.Namespace)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	client := action.NewGet(actionConfig)
	releaseDetail, err := client.Run(queryParams.Name)
	if err != nil {
		klog.Errorf("get releaseDetail error: %s", err)
		return &utils.Response{Code: code.GetError, Msg: err.Error()}
	}
	var data map[string]interface{}
	state, err := h.GetReleaseRuntimeState(releaseDetail, queryParams.WithResource)
	if err != nil {
		klog.Errorf("get release runtime namespace=%s name=%s error: %s", queryParams.Namespace, queryParams.Name, err.Error())
	}
	data = map[string]interface{}{
		"objects":        state.Resources,
		"name":           releaseDetail.Name,
		"namespace":      releaseDetail.Namespace,
		"version":        releaseDetail.Version,
		"status":         releaseDetail.Info.Status,
		"runtime_status": state.Status,
		"chart_name":     releaseDetail.Chart.Name() + "-" + releaseDetail.Chart.Metadata.Version,
		"chart_version":  releaseDetail.Chart.Metadata.Version,
		"app_version":    releaseDetail.Chart.AppVersion(),
		"last_deployed":  releaseDetail.Info.LastDeployed,
	}

	return &utils.Response{Code: code.Success, Msg: "Success", Data: data}
}

const (
	ReleaseStatusRunning      = "Running"
	ReleaseStatusNotReady     = "NotReady"
	ReleaseStatusRunningFault = "RunningFault"
)

type ReleaseRuntimeState struct {
	Name      string                       `json:"name"`
	Namespace string                       `json:"namespace"`
	Status    string                       `json:"status"`
	Resources []*unstructured.Unstructured `json:"resources"`
}

var WorkloadGVRMap = map[string]*schema.GroupVersionResource{
	"Deployment":  DeploymentGVR,
	"StatefulSet": StatefulSetGVR,
	"DaemonSet":   DaemonSetGVR,
	"Job":         JobGVR,
	"CronJob":     CronJobGVR,
}

func (h *Helm) GetReleaseObjects(release *release.Release) []*unstructured.Unstructured {
	var objects []*unstructured.Unstructured
	yamlList := strings.SplitAfter(release.Manifest, "\n---")
	for _, yamlStr := range yamlList {
		obj := &unstructured.Unstructured{}
		yamlBytes := []byte(yamlStr)
		decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewBuffer(yamlBytes), len(yamlBytes))
		if err := decoder.Decode(obj); err != nil {
			klog.Errorf("decode object yaml: %s, error: %s", yamlStr, err.Error())
			continue
		} else {
			objects = append(objects, obj)
		}
	}
	return objects
}

func (h *Helm) GetReleaseRuntimeState(rel *release.Release, withResource bool) (*ReleaseRuntimeState, error) {
	var resources []*unstructured.Unstructured
	state := &ReleaseRuntimeState{
		Name:      rel.Name,
		Namespace: rel.Namespace,
		Status:    ReleaseStatusNotReady,
	}
	objects := h.GetReleaseObjects(rel)
	for i, object := range objects {
		switch object.GetKind() {
		case "Deployment", "StatefulSet", "DaemonSet", "Job":
			if wr, err := h.addWorkloadObject(object, rel.Namespace, withResource); err != nil {
				return state, err
			} else {
				resources = append(resources, wr...)
			}
		case "Pod":
			if pod, err := h.client.Dynamic().Resource(*PodGVR).Namespace(rel.Namespace).Get(
				context.Background(), object.GetName(), metav1.GetOptions{}); err != nil {
				return state, err
			} else {
				resources = append(resources, pod)
			}
		case "Service":
			if withResource {
				svc, err := h.client.Dynamic().Resource(*ServiceGVR).Namespace(rel.Namespace).Get(
					context.Background(), object.GetName(), metav1.GetOptions{})
				if err != nil {
					return state, err
				}
				resources = append(resources, svc)
			}
		default:
			if withResource {
				resources = append(resources, objects[i])
			}
		}
	}
	if withResource {
		state.Resources = resources
	}
	status, err := h.releaseRuntimeStatus(resources)
	if err != nil {
		return state, err
	}
	state.Status = status
	return state, nil
}

func (h *Helm) releaseRuntimeStatus(releaseResources []*unstructured.Unstructured) (string, error) {
	for _, object := range releaseResources {
		if object.GetKind() == "Pod" {
			pod := &corev1.Pod{}
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(object.Object, pod); err != nil {
				return "", err
			}
			if !h.isPodReady(pod) {
				deltaSec := time.Now().Sub(pod.CreationTimestamp.Time).Seconds()
				if deltaSec > 600 {
					// 超过10分钟pod未就绪，则认为运行失败
					return ReleaseStatusRunningFault, nil
				} else {
					// 10分钟内未就绪
					return ReleaseStatusNotReady, nil
				}
			}
		}
	}
	return ReleaseStatusRunning, nil
}

func (h *Helm) isPodReady(pod *corev1.Pod) bool {
	if pod.Status.Phase == corev1.PodSucceeded {
		return true
	}
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func (h *Helm) addWorkloadObject(
	object *unstructured.Unstructured,
	namespace string,
	withResource bool) (resources []*unstructured.Unstructured, err error) {
	obj, err := h.client.Dynamic().Resource(*WorkloadGVRMap[object.GetKind()]).Namespace(namespace).Get(
		context.Background(), object.GetName(), metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get namespace %s workload %s/%s error: %s", namespace, object.GetKind(), object.GetName(), err.Error())
	}
	if withResource {
		resources = append(resources, obj)
	}
	podLabels, ok, err := unstructured.NestedStringMap(obj.Object, "spec", "selector", "matchLabels")
	if err != nil {
		return nil, fmt.Errorf("get namespace %s workload %s/%s labels error: %s", namespace, object.GetKind(), object.GetName(), err.Error())
	}
	if !ok {
		return nil, fmt.Errorf("get namespace %s workload %s/%s labels error", namespace, object.GetKind(), object.GetName())
	}
	pods, err := h.client.Dynamic().Resource(*PodGVR).Namespace(namespace).List(
		context.Background(), metav1.ListOptions{LabelSelector: labels.Set(podLabels).AsSelector().String()})
	if err != nil {
		return nil, fmt.Errorf("get namespace %s workload %s/%s pods error: %s", namespace, object.GetKind(), object.GetName(), err.Error())
	}
	for idx := range pods.Items {
		resources = append(resources, &pods.Items[idx])
	}
	return resources, nil
}

type HelmObjectParams struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Values     string `json:"values"`
	ChartBytes []byte `json:"chart_bytes"`
}

func (h *Helm) Create(requestParams interface{}) *utils.Response {
	createParams := &HelmObjectParams{}
	if err := utils.ConvertTypeByJson(requestParams, &createParams); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	if createParams.Name == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "Name is blank"}
	}
	if createParams.Namespace == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "Namespace is blank"}
	}
	chart, err := loader.LoadArchive(bytes.NewReader(createParams.ChartBytes))
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	actionConfig, err := h.actionConfig(createParams.Namespace)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}

	actionConfig.Releases.MaxHistory = 3
	clientInstall := action.NewInstall(actionConfig)
	clientInstall.ReleaseName = createParams.Name
	clientInstall.Namespace = createParams.Namespace

	values := make(map[string]interface{})
	err = yaml.Unmarshal([]byte(createParams.Values), &values)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: "values参数解析错误：" + err.Error()}
	}
	_, err = clientInstall.Run(chart, values)

	if err != nil {
		klog.Errorf("install release error: %s", err)
		clientUnInstall := action.NewUninstall(actionConfig)
		_, err1 := clientUnInstall.Run(createParams.Name)
		if err1 != nil {
			klog.Errorf("uninstall release error: %s", err)
		}
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success, Msg: "Success"}
}

func (h *Helm) Update(requestParams interface{}) *utils.Response {
	updateParams := &HelmObjectParams{}
	if err := utils.ConvertTypeByJson(requestParams, &updateParams); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	if updateParams.Name == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "Name is blank"}
	}
	if updateParams.Namespace == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "Namespace is blank"}
	}
	chart, err := loader.LoadArchive(bytes.NewReader(updateParams.ChartBytes))
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	actionConfig, err := h.actionConfig(updateParams.Namespace)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}

	actionConfig.Releases.MaxHistory = 3
	clientInstall := action.NewUpgrade(actionConfig)
	clientInstall.Namespace = updateParams.Namespace

	values := make(map[string]interface{})
	err = yaml.Unmarshal([]byte(updateParams.Values), &values)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: "values参数解析错误：" + err.Error()}
	}
	_, err = clientInstall.Run(updateParams.Name, chart, values)

	if err != nil {
		klog.Errorf("upgrade release error: %s", err)
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success, Msg: "Success"}
}

func (h *Helm) Delete(requestParams interface{}) *utils.Response {
	delParams := &HelmObjectParams{}
	if err := utils.ConvertTypeByJson(requestParams, &delParams); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	if delParams.Name == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "Name is blank"}
	}
	if delParams.Namespace == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "Namespace is blank"}
	}
	actionConfig, err := h.actionConfig(delParams.Namespace)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}

	clientInstall := action.NewUninstall(actionConfig)
	_, err = clientInstall.Run(delParams.Name)

	if err != nil {
		klog.Errorf("uninstall release error: %s", err)
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success, Msg: "Success"}
}
