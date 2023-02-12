package project

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/kubernetes/resource"
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
	project, err := p.models.ProjectManager.Get(projectId)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: "获取工作空间失败: " + err.Error()}
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
		resTypes := []string{
			kubetypes.ConfigMapType,
			kubetypes.SecretType,
			kubetypes.ServiceType,
			kubetypes.IngressType,
			kubetypes.PersistentVolumeClaimType,
		}
		errs := ""
		for _, resType := range resTypes {
			res := p.kubeClient.Delete(project.ClusterId, resType, &resource.DeleteParams{
				Namespace:     project.Namespace,
				LabelSelector: kubetypes.ProjectLabelSelector,
			})
			if !res.IsSuccess() {
				errs += res.Msg + "\n"
			}
		}
		if errs != "" {
			return &utils.Response{Code: code.CreateError, Msg: "删除k8s以下资源失败：\n" + errs}
		}
	}
	err = p.models.ProjectManager.Delete(project)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: "删除工作空间失败: " + err.Error()}
	}
	return &utils.Response{Code: code.Success}
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

func (p *ProjectService) processServiceObj(obj *unstructured.Unstructured) error {
	kind := obj.GetKind()
	name := obj.GetName()
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

func (p *ProjectService) cloneK8sResource(oriProject, destProject *types.Project, resType string) error {
	var process = false
	queryParams := &resource.QueryParams{
		Namespace:     oriProject.Namespace,
		LabelSelector: kubetypes.ProjectLabelSelector,
		Process:       &process,
	}
	res := p.kubeClient.List(oriProject.ClusterId, resType, queryParams)
	errs := ""
	if !res.IsSuccess() {
		return fmt.Errorf("get kubernetes %s resource error: %s", resType, res.Msg)
	}
	if res.Data == nil {
		return nil
	}
	var objects []map[string]interface{}
	resList := res.Data
	if resType == kubetypes.IngressType {
		data := make(map[string]interface{})
		if err := utils.ConvertTypeByJson(res.Data, &data); err != nil {
			return fmt.Errorf("get %s data error\n", resType)
		}
		if ingresses, ok := data["ingresses"]; ok {
			resList = ingresses
		} else {
			return fmt.Errorf("not found ingress data")
		}
	}
	if err := utils.ConvertTypeByJson(resList, &objects); err != nil {
		return fmt.Errorf("get kubernetes %s resource error: %s", resType, err.Error())
	}
	for _, obj := range objects {
		unstructuredObj := unstructured.Unstructured{Object: obj}
		unstructuredObj.SetNamespace(destProject.Namespace)
		unstructuredObj.SetManagedFields(nil)
		unstructuredObj.SetUID("")
		unstructuredObj.SetResourceVersion("")
		unstructuredObj.SetCreationTimestamp(metav1.Time{})
		if resType == kubetypes.ServiceType {
			if err := p.processServiceObj(&unstructuredObj); err != nil {
				errs += err.Error() + "\n"
			}
		}
		yamlStr, _ := yaml.Marshal(unstructuredObj.Object)
		applyRes := p.kubeClient.Apply(destProject.ClusterId, &resource.ApplyParams{
			YamlStr: string(yamlStr),
		})
		if !applyRes.IsSuccess() {
			errs += fmt.Sprintf("apply object %s/%s error: %s\n", unstructuredObj.GetKind(), unstructuredObj.GetName(), applyRes.Msg)
		}
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
	resTypes := []string{
		kubetypes.ConfigMapType,
		kubetypes.SecretType,
		kubetypes.ServiceType,
		kubetypes.IngressType,
		kubetypes.PersistentVolumeClaimType,
	}
	errs := ""
	for _, resType := range resTypes {
		err = p.cloneK8sResource(oriProject, newProject, resType)
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
	typeKindMap := map[string]string{
		kubetypes.ConfigMapType:             "ConfigMap",
		kubetypes.SecretType:                "Secret",
		kubetypes.PersistentVolumeClaimType: "PersistentVolumeClaim",
	}
	process := false
	for resType, resKind := range typeKindMap {
		res := p.kubeClient.List(oriProject.ClusterId, resType, &resource.QueryParams{
			Namespace:     oriProject.Namespace,
			Process:       &process,
			LabelSelector: kubetypes.ProjectLabelSelector,
		})
		if !res.IsSuccess() {
			return res
		}
		data[resKind] = res.Data
	}
	return &utils.Response{Code: code.Success, Data: data}
}
