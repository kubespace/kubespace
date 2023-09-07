package plugins

import (
	"bytes"
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	corerrors "github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/service/cluster"
	"github.com/kubespace/kubespace/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchV1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
	"strings"
	"text/template"
)

var WorkloadKinds = []string{"Deployment", "StatefulSet", "DaemonSet", "Job", "Pod", "CronJob"}

type DeployK8sPlugin struct {
	*model.Models
	KubeClient *cluster.KubeClient
}

func (p DeployK8sPlugin) Executor(params *ExecutorParams) (Executor, error) {
	return newDeployK8s(params, p.Models, p.KubeClient)
}

type deployK8sParams struct {
	Cluster   string                 `json:"cluster"`
	Namespace string                 `json:"namespace"`
	Yaml      string                 `json:"yaml"`
	Images    string                 `json:"images"`
	Env       map[string]interface{} `json:"env"`
}

type deployK8sResultResource struct {
	Namespace string   `json:"namespace"`
	Kind      string   `json:"kind"`
	Name      string   `json:"name"`
	Images    []string `json:"images"`
}

type deployK8sResult struct {
	Cluster   string                     `json:"cluster"`
	Resources []*deployK8sResultResource `json:"resources"`
}

type deployK8s struct {
	Logger
	models     *model.Models
	kubeClient *cluster.KubeClient
	params     *deployK8sParams
	images     map[string]string
	result     *deployK8sResult
}

func newDeployK8s(params *ExecutorParams, models *model.Models, kubeClient *cluster.KubeClient) (*deployK8s, error) {
	var deployParams deployK8sParams
	if err := utils.ConvertTypeByJson(params.Params, &deployParams); err != nil {
		params.Logger.Log("插件参数：%v", params.Params)
		return nil, fmt.Errorf("插件参数错误: %s", err.Error())
	}
	return &deployK8s{
		models:     models,
		kubeClient: kubeClient,
		params:     &deployParams,
		result:     &deployK8sResult{},
		Logger:     params.Logger,
	}, nil
}

func (u *deployK8s) Execute() (interface{}, error) {
	if err := u.execute(); err != nil {
		return nil, err
	}
	return u.result, nil
}

func (u *deployK8s) Cancel() error {
	return nil
}

func (u *deployK8s) execute() error {
	if u.params.Yaml == "" {
		u.Log("要部署的k8s资源内容为空")
		return nil
	}
	if u.params.Images == "" {
		u.Log("要升级的镜像列表参数为空")
	}
	u.images = stringToImage(u.params.Images)
	u.Log("升级的镜像列表：%v", u.params.Images)
	if u.params.Cluster == "" {
		u.Log("集群参数为空")
		return fmt.Errorf("集群参数为空")
	}
	clusterObj, err := u.models.ClusterManager.GetByName(u.params.Cluster)
	if err != nil {
		u.Log("获取集群失败：%s", err.Error())
		return err
	}
	u.result.Cluster = clusterObj.Name1
	if u.params.Namespace == "" {
		u.params.Namespace = "default"
	}
	yamlTpl, err := u.templateParse(u.params.Yaml)
	if err != nil {
		return err
	}
	yamlList := strings.Split(yamlTpl, "---\n")
	var destYamlList []string
	for _, yamlStr := range yamlList {
		destYaml, err := u.replaceResourceImage(yamlStr)
		if err != nil {
			return err
		}
		destYamlList = append(destYamlList, destYaml)
	}
	destYamlStr := strings.Join(destYamlList, "\n---\n")
	u.Log(destYamlStr)
	u.Log("开始部署资源到集群「%s」", clusterObj.Name1)
	resp := u.kubeClient.Apply(clusterObj.Name, map[string]string{
		"yaml": destYamlStr,
	})
	if !resp.IsSuccess() {
		u.Log("部署资源到集群失败：%s", resp.Msg)
		return corerrors.New(resp.Code, resp.Msg)
	}
	u.Log("部署资源到集群成功")
	return nil
}

func (u *deployK8s) replaceResourceImage(yamlStr string) (string, error) {
	yamlDict := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(yamlStr), &yamlDict); err != nil {
		u.Log("%s\n解析Yaml失败：%s", yamlStr, err.Error())
		return "", corerrors.New(code.MarshalError, err)
	}

	obj := unstructured.Unstructured{Object: yamlDict}
	obj.SetNamespace(u.params.Namespace)

	deployRes := &deployK8sResultResource{
		Namespace: u.params.Namespace,
		Kind:      obj.GetKind(),
		Name:      obj.GetName(),
	}

	u.result.Resources = append(u.result.Resources, deployRes)

	if !utils.Contains(WorkloadKinds, obj.GetKind()) {
		return yamlStr, nil
	}

	var resObject interface{}

	switch obj.GetKind() {
	case "Pod":
		var pod corev1.Pod
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &pod); err != nil {
			u.Log("%s\n转换Pod资源失败: %s", yamlStr, err.Error())
			return "", corerrors.New(code.ParseError, err)
		}
		deployRes.Images = u.replaceContainerImage(pod.Spec.Containers)
		resObject = &pod
	case "Deployment":
		var deployment appsv1.Deployment
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &deployment); err != nil {
			u.Log("%s\n转换Deployment资源失败: %s", yamlStr, err.Error())
			return "", corerrors.New(code.ParseError, err)
		}
		deployRes.Images = u.replaceContainerImage(deployment.Spec.Template.Spec.Containers)
		resObject = &deployment
	case "StatefulSet":
		var sts appsv1.StatefulSet
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &sts); err != nil {
			u.Log("%s\n转换StatefulSet资源失败: %s", yamlStr, err.Error())
			return "", corerrors.New(code.ParseError, err)
		}
		deployRes.Images = u.replaceContainerImage(sts.Spec.Template.Spec.Containers)
		resObject = &sts
	case "DaemonSet":
		var ds appsv1.DaemonSet
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &ds); err != nil {
			u.Log("%s\n转换DaemonSet资源失败: %s", yamlStr, err.Error())
			return "", corerrors.New(code.ParseError, err)
		}
		deployRes.Images = u.replaceContainerImage(ds.Spec.Template.Spec.Containers)
		resObject = &ds
	case "Job":
		var job batchv1.Job
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &job); err != nil {
			u.Log("%s\n转换Job资源失败: %s", yamlStr, err.Error())
			return "", corerrors.New(code.ParseError, err)
		}
		deployRes.Images = u.replaceContainerImage(job.Spec.Template.Spec.Containers)
		resObject = &job
	case "CronJob":
		var cronjob batchV1beta1.CronJob
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &cronjob); err != nil {
			u.Log("%s\n转换CronJob资源失败: %s", yamlStr, err.Error())
			return "", corerrors.New(code.ParseError, err)
		}
		deployRes.Images = u.replaceContainerImage(cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers)
		resObject = &cronjob
	}
	var err error
	if obj.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(resObject); err != nil {
		u.Log("转换%s资源失败：%s", obj.GetKind(), err.Error())
		return "", corerrors.New(code.ParseError, err)
	}

	objBytes, err := yaml.Marshal(obj.Object)
	if err != nil {
		u.Log("marshal resource error: %s", err.Error())
		return "", corerrors.New(code.MarshalError, err)
	}
	return string(objBytes), nil
}

func (u *deployK8s) replaceContainerImage(containers []corev1.Container) []string {
	var images []string
	for i, c := range containers {
		if matchImage := u.matchImage(c.Image); matchImage != "" {
			containers[i].Image = matchImage
			images = append(images, matchImage)
			u.Log("部署容器原镜像「%s」为「%s」", c.Image, matchImage)
		}
	}
	return images
}

func (u *deployK8s) matchImage(srcImage string) string {
	srcImage = utils.GetImageName(srcImage)
	for name, img := range u.images {
		if name == srcImage {
			return img
		}
	}
	return ""
}

func (u *deployK8s) templateParse(yamlTpl string) (string, error) {
	tpl, err := template.New("yaml").Parse(yamlTpl)
	if err != nil {
		u.Log("parse template error: %s", err.Error())
		return "", err
	}
	buf := bytes.Buffer{}
	err = tpl.Execute(&buf, u.params.Env)
	if err != nil {
		u.Log("execute template error: %s", err.Error())
		return "", err
	}
	return string(buf.Bytes()), nil
}
