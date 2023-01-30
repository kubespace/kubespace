package project

import (
	"fmt"
	kubetypes "github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/service/cluster"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
	"time"
)

type ProjectService struct {
	models     *model.Models
	kubeClient *cluster.KubeClient
	appService *AppService
}

func NewProjectService(models *model.Models, kubeClient *cluster.KubeClient, appService *AppService) *ProjectService {
	return &ProjectService{
		models:     models,
		kubeClient: kubeClient,
		appService: appService,
	}
}

func (p *ProjectService) Delete(projectId uint, delResource bool) *utils.Response {
	resp := &utils.Response{Code: code.Success}
	project, err := p.models.ProjectManager.Get(projectId)
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "获取工作空间失败: " + err.Error()
		return resp
	}
	apps, err := p.appService.ListApp(types.AppVersionScopeProjectApp, projectId)
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: err.Error()}
	}
	for _, app := range apps {
		if app.Status != types.AppStatusUninstall {
			return &utils.Response{Code: code.DeleteError, Msg: "删除工作空间失败：应用" + app.Name + "正在运行"}
		}
	}
	if delResource {
		kinds := []string{"ConfigMap", "Secret", "Service", "Ingress", "PersistentVolumeClaim"}
		errs := ""
		for _, kind := range kinds {
			err = p.deleteK8sResource(project, kind)
			if err != nil {
				errs += err.Error()
			}
		}
		if errs != "" {
			return &utils.Response{Code: code.CreateError, Msg: "删除k8s以下资源失败：\n" + errs}
		}
	}
	err = p.models.ProjectManager.Delete(project)
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "删除工作空间失败: " + err.Error()
		return resp
	}
	return resp
}

func (p *ProjectService) Get(projectId uint, withDetail bool) *utils.Response {
	project, err := p.models.ProjectManager.Get(projectId)
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: "获取工作空间失败: " + err.Error()}
	}
	data := map[string]interface{}{
		"id":          project.ID,
		"name":        project.Name,
		"description": project.Description,
		"cluster_id":  project.ClusterId,
		"namespace":   project.Namespace,
		"owner":       project.Owner,
		"create_time": project.CreateTime,
		"update_time": project.UpdateTime,
	}
	clusterObj, err := p.models.ClusterManager.GetByName(project.ClusterId)
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: "获取集群信息失败: %s" + err.Error()}
	}
	data["cluster"] = clusterObj
	if withDetail {
		resp := p.kubeClient.Get(project.ClusterId, kubetypes.ClusterType, map[string]interface{}{
			"workspace": project.ID,
			"namespace": project.Namespace,
		})
		if resp.IsSuccess() {
			data["resource"] = resp.Data
		} else {
			return resp
		}
	}

	return &utils.Response{Code: code.Success, Data: data}
}

func (p *ProjectService) deleteK8sResource(project *types.Project, kind string) error {
	var resType string
	switch kind {
	case "ConfigMap":
		resType = kubetypes.ConfigMapType
	case "Secret":
		resType = kubetypes.SecretType
	case "Service":
		resType = kubetypes.ServiceType
	case "Ingress":
		resType = kubetypes.IngressType
	case "PersistentVolumeClaim":
		resType = kubetypes.PersistentVolumeClaimType
	default:
		return fmt.Errorf("not found %s kind", kind)
	}
	res := p.kubeClient.List(project.ClusterId, resType, map[string]interface{}{
		"namespace": project.Namespace,
		"labels":    map[string]string{"kubespace.cn/belong-to": "project"},
	})
	errs := ""
	if res.IsSuccess() {
		if res.Data != nil {
			var delRes []map[string]interface{}
			var resList interface{}
			klog.Info(res.Data)
			if kind == "Ingress" {
				if data, ok := res.Data.(map[string]interface{}); ok {
					resList = data["ingresses"]
					if resList == nil {
						return nil
					}
				} else {
					return fmt.Errorf("get %s data error\n", kind)
				}
			} else {
				resList = res.Data
			}
			klog.Info(resList)
			if data, ok := resList.([]interface{}); ok {
				for _, do := range data {
					d := do.(map[string]interface{})
					if name, ok := d["name"]; ok {
						delRes = append(delRes, map[string]interface{}{
							"name":      name,
							"namespace": project.Namespace,
						})
					} else {
						errs += fmt.Sprintf("not found %s object name field\n", kind)
					}
				}
				params := map[string]interface{}{
					"resources": delRes,
				}
				resObj := p.kubeClient.Delete(project.ClusterId, resType, params)
				if !resObj.IsSuccess() {
					return fmt.Errorf("delete %s resources error: %s\n", kind, resObj.Msg)
				}
			} else {
				return fmt.Errorf("get %s data error\n", kind)
			}
		}
	} else {
		return fmt.Errorf("get %s resources error: %s\n", kind, res.Msg)
	}
	if errs != "" {
		return fmt.Errorf(errs)
	}
	return nil
}

func (p *ProjectService) processServiceObj(obj *unstructured.Unstructured, kind, name string) error {
	if err := unstructured.SetNestedField(obj.Object, "", "spec", "clusterIP"); err != nil {
		klog.Errorf("set object %s/%s spec.clusterIP error: %s\n", kind, name, err.Error())
	}
	if err := unstructured.SetNestedField(obj.Object, nil, "spec", "clusterIPs"); err != nil {
		klog.Errorf("set object %s/%s spec.clusterIP error: %s\n", kind, name, err.Error())
	}
	serviceType, ok, err := unstructured.NestedString(obj.Object, "spec", "type")
	if err != nil {
		klog.Errorf("get object %s/%s spec.type error: %s", kind, name, err.Error())
	}
	if !ok {
		klog.Errorf("not get object %s/%s spec.type", kind, name)
	}
	if serviceType == "NodePort" {
		ports, ok, err := unstructured.NestedSlice(obj.Object, "spec", "ports")
		if err != nil {
			klog.Errorf("get object %s/%s spec.ports error: %s", kind, name, err.Error())
		}
		if !ok {
			klog.Errorf("not get object %s/%s spec.ports", kind, name)
		}
		for _, portObj := range ports {
			if port, ok := portObj.(map[string]interface{}); ok {
				if _, ok = port["nodePort"]; ok {
					delete(port, "nodePort")
				}
			}
		}
	}
	return nil
}

func (p *ProjectService) cloneK8sResource(oriProject, destProject *types.Project, kind string) error {
	apiVersion := "v1"
	var resType string
	switch kind {
	case "ConfigMap":
		resType = kubetypes.ConfigMapType
	case "Secret":
		resType = kubetypes.SecretType
	case "Service":
		resType = kubetypes.ServiceType
	case "Ingress":
		resType = kubetypes.IngressType
	case "PersistentVolumeClaim":
		resType = kubetypes.PersistentVolumeClaimType
	default:
		return fmt.Errorf("not found %s kind", kind)
	}
	res := p.kubeClient.List(oriProject.ClusterId, resType, map[string]interface{}{
		"namespace": oriProject.Namespace,
		"labels":    map[string]string{"kubespace.cn/belong-to": "project"},
	})
	errs := ""
	if res.IsSuccess() {
		if res.Data != nil {
			var resList interface{}
			if kind == "Ingress" {
				if data, ok := res.Data.(map[string]interface{}); ok {
					group, ok := data["group"]
					if ok {
						if group == "extensions" {
							apiVersion = "extensions/v1beta1"
						} else {
							apiVersion = "networking.k8s.io/v1"
						}
					} else {
						return fmt.Errorf("get ingress group error\n")
					}
					resList = data["ingresses"]
				} else {
					return fmt.Errorf("get %s data error\n", kind)
				}
			} else {
				resList = res.Data
			}
			if data, ok := resList.([]interface{}); ok {
				for _, do := range data {
					d := do.(map[string]interface{})
					if name, ok := d["name"]; ok {
						resObj := p.kubeClient.Get(oriProject.ClusterId, resType, map[string]interface{}{
							"name":      name,
							"namespace": oriProject.Namespace,
						})
						if !resObj.IsSuccess() {
							errs += fmt.Sprintf("get %s/%s error: %s\n", kind, name, resObj.Msg)
							continue
						}
						if obj, ok := resObj.Data.(map[string]interface{}); ok {
							unstructuredObj := unstructured.Unstructured{Object: obj}
							unstructuredObj.SetNamespace(destProject.Namespace)
							unstructuredObj.SetKind(kind)
							unstructuredObj.SetAPIVersion(apiVersion)
							unstructuredObj.SetManagedFields(nil)
							unstructuredObj.SetUID("")
							unstructuredObj.SetResourceVersion("")
							unstructuredObj.SetCreationTimestamp(metav1.Time{})
							if kind == "Service" {
								if err := p.processServiceObj(&unstructuredObj, kind, name.(string)); err != nil {
									errs += fmt.Sprintf("process object %s/%s error: %s\n", kind, name, err.Error())
								}
							}
							yamlStr, _ := yaml.Marshal(unstructuredObj.Object)
							klog.Info(string(yamlStr))
							applyRes := p.kubeClient.Apply(destProject.ClusterId, map[string]interface{}{
								"yaml": string(yamlStr),
							})
							if !applyRes.IsSuccess() {
								errs += fmt.Sprintf("create object %s/%s error: %s\n", kind, name, applyRes.Msg)
							}
						} else {
							errs += fmt.Sprintf("get %s/%s data error\n", kind, name)
						}
					} else {
						errs += fmt.Sprintf("not found %s object name field\n", kind)
					}
				}
			} else {
				return fmt.Errorf("get %s data error\n", kind)
			}
		}
	} else {
		return fmt.Errorf("get %s resources error: %s\n", kind, res.Msg)
	}
	if errs != "" {
		return fmt.Errorf(errs)
	}
	return nil
}

func (p *ProjectService) Clone(ser *serializers.ProjectCloneSerializer, user *types.User) *utils.Response {
	oriProject, err := p.models.ProjectManager.Get(ser.OriginProjectId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: "获取工作空间失败:" + err.Error()}
	}
	newProject := &types.Project{
		Name:        ser.Name,
		Description: ser.Description,
		ClusterId:   ser.ClusterId,
		Namespace:   ser.Namespace,
		Owner:       ser.Owner,
		CreateUser:  user.Name,
		UpdateUser:  user.Name,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
	newProject, err = p.models.ProjectManager.Clone(ser.OriginProjectId, newProject)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	kinds := []string{"ConfigMap", "Secret", "Service", "Ingress", "PersistentVolumeClaim"}
	errs := ""
	for _, kind := range kinds {
		err = p.cloneK8sResource(oriProject, newProject, kind)
		if err != nil {
			errs += err.Error()
		}
	}
	if errs != "" {
		return &utils.Response{Code: code.CreateError, Msg: "克隆k8s以下资源失败：\n" + errs}
	}
	return &utils.Response{Code: code.Success}
}

func (p *ProjectService) GetProjectNamespaceResources(ser *serializers.ProjectResourcesSerializer) *utils.Response {
	oriProject, err := p.models.ProjectManager.Get(ser.ProjectId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: "获取工作空间失败:" + err.Error()}
	}
	data := map[string]interface{}{}
	for _, kind := range []string{"ConfigMap", "Secret", "PersistentVolumeClaim"} {
		var resType string
		switch kind {
		case "ConfigMap":
			resType = kubetypes.ConfigMapType
		case "Secret":
			resType = kubetypes.SecretType
		case "Service":
			resType = kubetypes.ServiceType
		case "Ingress":
			resType = kubetypes.IngressType
		case "PersistentVolumeClaim":
			resType = kubetypes.PersistentVolumeClaimType
		default:
			return &utils.Response{Code: code.ParamsError, Msg: "get kuberesource error"}
		}
		res := p.kubeClient.List(oriProject.ClusterId, resType, map[string]interface{}{
			"namespace": oriProject.Namespace,
			"labels":    map[string]string{"kubespace.cn/belong-to": "project"},
		})
		if res.IsSuccess() {
			data[kind] = res.Data
		} else {
			return res
		}
	}
	return &utils.Response{Code: code.Success, Data: data}
}
