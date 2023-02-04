package plugins

import (
	"errors"
	"fmt"
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
)

var WorkloadKinds = []string{"Deployment", "StatefulSet", "DaemonSet", "Job", "Pod", "CronJob"}

type DeployK8sPlugin struct {
	*model.Models
	KubeClient *cluster.KubeClient
}

func (p DeployK8sPlugin) Execute(params *PluginParams) (interface{}, error) {
	deploy, err := newDeployK8s(params, p.Models, p.KubeClient)
	if err != nil {
		return nil, err
	}
	err = deploy.execute()
	if err != nil {
		return nil, err
	}
	return deploy.result, nil
}

type deployK8sParams struct {
	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
	Yaml      string `json:"yaml"`
	Images    string `json:"images"`
}

type deployK8sResult struct {
}

type deployK8s struct {
	models     *model.Models
	kubeClient *cluster.KubeClient
	params     *deployK8sParams
	images     []string
	result     *deployK8sResult
	*PluginLogger
}

func newDeployK8s(params *PluginParams, models *model.Models, kubeClient *cluster.KubeClient) (*deployK8s, error) {
	var deployParams deployK8sParams
	if err := utils.ConvertTypeByJson(params.Params, &deployParams); err != nil {
		params.Logger.Log("插件参数：%v", params.Params)
		return nil, fmt.Errorf("插件参数错误: %s", err.Error())
	}
	return &deployK8s{
		models:       models,
		kubeClient:   kubeClient,
		params:       &deployParams,
		result:       &deployK8sResult{},
		PluginLogger: params.Logger,
	}, nil
}

func (u *deployK8s) execute() error {
	if u.params.Yaml == "" {
		u.Log("要部署的k8s资源内容为空")
		return nil
	}
	if u.params.Images == "" {
		u.Log("要升级的镜像列表参数为空")
		//return nil
	}
	u.images = strings.Split(u.params.Images, ",")
	u.Log("升级的镜像列表：%v", u.images)
	if u.params.Cluster == "" {
		u.Log("集群参数为空")
		return fmt.Errorf("集群参数为空")
	}
	clusterObj, err := u.models.ClusterManager.GetByName(u.params.Cluster)
	if err != nil {
		u.Log("获取集群失败：%s", err.Error())
		return err
	}
	if u.params.Namespace == "" {
		u.params.Namespace = "default"
	}
	yamlList := strings.Split(u.params.Yaml, "---\n")
	destYamlStr := ""
	replaced := false
	for _, yamlStr := range yamlList {
		destYaml, imageReplaced, err := u.replaceResourceImage(yamlStr)
		if err != nil {
			return err
		}
		if imageReplaced {
			replaced = true
		}
		destYamlStr += destYaml + "\n---\n"
	}
	if !replaced {
		u.Log("未匹配到可替换的镜像")
	}
	u.Log(destYamlStr)
	u.Log("开始部署资源到集群「%s」", clusterObj.Name1)
	resp := u.kubeClient.Apply(clusterObj.Name, map[string]string{
		"yaml": destYamlStr,
	})
	if !resp.IsSuccess() {
		u.Log("部署资源到集群失败：%s", resp.Msg)
		return errors.New(resp.Msg)
	} else {
		u.Log("部署资源到集群成功")
	}
	return nil
}

func (u *deployK8s) replaceResourceImage(yamlStr string) (string, bool, error) {
	yamlDict := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(yamlStr), &yamlDict)
	if err != nil {
		u.Log("%s\n解析Yaml失败：%s", yamlStr, err.Error())
		return "", false, err
	}
	obj := unstructured.Unstructured{Object: yamlDict}
	obj.SetNamespace(u.params.Namespace)
	replaced := false
	if utils.Contains(WorkloadKinds, obj.GetKind()) {
		switch obj.GetKind() {
		case "Pod":
			var pod corev1.Pod
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &pod)
			if err != nil {
				u.Log("%s\n转换Pod资源失败: %s", yamlStr, err.Error())
				return "", false, err
			}
			if replaced, err = u.replaceContainerImage(pod.Spec.Containers); err != nil {
				return "", false, err
			}
			if obj.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&pod); err != nil {
				u.Log("转换Pod资源失败：%s", err.Error())
				return "", false, err
			}
		case "Deployment":
			var deployment appsv1.Deployment
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &deployment)
			if err != nil {
				u.Log("%s\n转换Deployment资源失败: %s", yamlStr, err.Error())
				return "", false, err
			}
			if replaced, err = u.replaceContainerImage(deployment.Spec.Template.Spec.Containers); err != nil {
				return "", false, err
			}
			if obj.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&deployment); err != nil {
				u.Log("转换Deployment资源失败：%s", err.Error())
				return "", false, err
			}
		case "StatefulSet":
			var sts appsv1.StatefulSet
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &sts)
			if err != nil {
				u.Log("%s\n转换StatefulSet资源失败: %s", yamlStr, err.Error())
				return "", false, err
			}
			if replaced, err = u.replaceContainerImage(sts.Spec.Template.Spec.Containers); err != nil {
				return "", false, err
			}
			if obj.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&sts); err != nil {
				u.Log("转换StatefulSet资源失败：%s", err.Error())
				return "", false, err
			}
		case "DaemonSet":
			var ds appsv1.DaemonSet
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &ds)
			if err != nil {
				u.Log("%s\n转换DaemonSet资源失败: %s", yamlStr, err.Error())
				return "", false, err
			}
			if replaced, err = u.replaceContainerImage(ds.Spec.Template.Spec.Containers); err != nil {
				return "", false, err
			}
			if obj.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&ds); err != nil {
				u.Log("转换DaemonSet资源失败：%s", err.Error())
				return "", false, err
			}
		case "Job":
			var job batchv1.Job
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &job)
			if err != nil {
				u.Log("%s\n转换Job资源失败: %s", yamlStr, err.Error())
				return "", false, err
			}
			if replaced, err = u.replaceContainerImage(job.Spec.Template.Spec.Containers); err != nil {
				return "", false, err
			}
			if obj.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&job); err != nil {
				u.Log("转换Job资源失败：%s", err.Error())
				return "", false, err
			}
		case "CronJob":
			var cronjob batchV1beta1.CronJob
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &cronjob)
			if err != nil {
				u.Log("%s\n转换CronJob资源失败: %s", yamlStr, err.Error())
				return "", false, err
			}
			if replaced, err = u.replaceContainerImage(cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers); err != nil {
				return "", false, err
			}
			if obj.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&cronjob); err != nil {
				u.Log("转换CronJob资源失败：%s", err.Error())
				return "", false, err
			}
		}
	}
	objBytes, err := yaml.Marshal(obj.Object)
	if err != nil {
		u.Log("marshal resource error: %s", err.Error())
		return "", false, err
	}
	return string(objBytes), replaced, nil
}

func (u *deployK8s) replaceContainerImage(containers []corev1.Container) (bool, error) {
	replaced := false
	for i, c := range containers {
		if matchImage := u.matchImage(c.Image); matchImage != "" {
			containers[i].Image = matchImage
			replaced = true
			u.Log("替换容器原镜像「%s」为「%s」", c.Image, matchImage)
		}
	}
	return replaced, nil
}

func (u *deployK8s) matchImage(srcImage string) string {
	srcImage = strings.Split(srcImage, ":")[0]
	if strings.Contains(strings.Split(srcImage, "/")[0], ".") {
		srcImage = strings.Join(strings.Split(srcImage, "/")[1:], "/")
	}
	for _, image := range u.images {
		if strings.Contains(image, srcImage+":") {
			return image
		}
	}
	return ""
}
