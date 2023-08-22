package resource

import (
	"context"
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/kubernetes/config"
	"github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/utils"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

var CustomResourceDefinitionGVR = &schema.GroupVersionResource{
	Group:    "apiextensions.k8s.io",
	Version:  "v1",
	Resource: "customresourcedefinitions",
}

var CustomResourceDefinitionV1beta1GVR = &schema.GroupVersionResource{
	Group:    "apiextensions.k8s.io",
	Version:  "v1beta1",
	Resource: "customresourcedefinitions",
}

type CustomResourceDefinition struct {
	*Resource
}

func NewCustomResourceDefinition(config *config.KubeConfig) *CustomResourceDefinition {
	p := &CustomResourceDefinition{}
	gvr := CustomResourceDefinitionV1beta1GVR
	if config.Client.VersionGreaterThan(types.ServerVersion16) {
		gvr = CustomResourceDefinitionGVR
	}
	p.Resource = NewResource(config, types.CustomResourceDefinitionType, gvr, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.DeleteAction: p.Delete,
		types.UpdateAction: p.Update,
	}
	return p
}

func (s *CustomResourceDefinition) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	if s.gvr.Version == "v1beta1" {
		crd := &apiextv1beta1.CustomResourceDefinition{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, crd); err != nil {
			return nil, err
		}
		var version = crd.Spec.Version
		for _, v := range crd.Spec.Versions {
			if v.Storage {
				version = v.Name
				break
			}
		}
		return map[string]interface{}{
			"name":        crd.Name,
			"resource":    crd.Spec.Names.Plural,
			"scope":       crd.Spec.Scope,
			"version":     version,
			"group":       crd.Spec.Group,
			"create_time": crd.CreationTimestamp,
		}, nil
	} else {
		crd := &apiextv1.CustomResourceDefinition{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, crd); err != nil {
			return nil, err
		}
		var version = ""
		for _, v := range crd.Spec.Versions {
			if v.Storage {
				version = v.Name
				break
			}
		}
		return map[string]interface{}{
			"name":        crd.Name,
			"resource":    crd.Spec.Names.Plural,
			"scope":       crd.Spec.Scope,
			"version":     version,
			"group":       crd.Spec.Group,
			"create_time": crd.CreationTimestamp,
		}, nil
	}
}

type CustomResource struct {
	*Resource
}

func NewCustomResource(config *config.KubeConfig) *CustomResource {
	p := &CustomResource{}
	p.Resource = NewResource(config, types.CustomResourceType, nil, nil)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.DeleteAction: p.Delete,
		types.UpdateAction: p.Update,
	}
	return p
}

type CustomResourceParam struct {
	Group    string `json:"group" form:"group"`
	Resource string `json:"resource" form:"resource"`
	Version  string `json:"version" form:"version"`
}

type CustomResourceQueryParams struct {
	CustomResourceParam
	Namespace string `json:"namespace" form:"namespace"`
	Name      string `json:"name" form:"name"`
	Output    string `json:"output" form:"output"`
}

func (c *CustomResource) gvr(param *CustomResourceParam) (*schema.GroupVersionResource, error) {
	if param.Group == "" {
		return nil, fmt.Errorf("custom resource group is blank")
	}
	if param.Resource == "" {
		return nil, fmt.Errorf("custom resource resource is blank")
	}
	if param.Version == "" {
		return nil, fmt.Errorf("custom resource version is blank")
	}
	gvr := &schema.GroupVersionResource{
		Group:    param.Group,
		Version:  param.Version,
		Resource: param.Resource,
	}
	return gvr, nil
}

func (c *CustomResource) List(params interface{}) *utils.Response {
	queryParams := &CustomResourceQueryParams{}
	if err := utils.ConvertTypeByJson(params, queryParams); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	gvr, err := c.gvr(&queryParams.CustomResourceParam)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	crs, err := c.client.Dynamic().Resource(*gvr).Namespace("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	var crList []map[string]interface{}
	for _, cr := range crs.Items {
		crList = append(crList, map[string]interface{}{
			"name":        cr.GetName(),
			"namespace":   cr.GetNamespace(),
			"create_time": cr.GetCreationTimestamp(),
		})
	}
	return &utils.Response{Code: code.Success, Msg: "Success", Data: crList}
}

func (c *CustomResource) Get(params interface{}) *utils.Response {
	queryParams := &CustomResourceQueryParams{}
	if err := utils.ConvertTypeByJson(params, queryParams); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	gvr, err := c.gvr(&queryParams.CustomResourceParam)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	var obj *unstructured.Unstructured
	if queryParams.Namespace != "" {
		obj, err = c.client.Dynamic().Resource(*gvr).Namespace(queryParams.Namespace).Get(
			context.Background(), queryParams.Name, metav1.GetOptions{})
	} else {
		obj, err = c.client.Dynamic().Resource(*gvr).Get(context.Background(), queryParams.Name, metav1.GetOptions{})
	}
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	if queryParams.Output == "yaml" {
		if yamlStr, err := yaml.Marshal(obj); err != nil {
			return &utils.Response{Code: code.MarshalError, Msg: err.Error()}
		} else {
			return &utils.Response{Code: code.Success, Data: string(yamlStr)}
		}
	}
	return &utils.Response{Code: code.Success, Data: obj}
}

type DeleteCustomResource struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type CustomResourceDeleteParams struct {
	CustomResourceParam
	Resources []*DeleteCustomResource `json:"resources"`
}

func (c *CustomResource) Delete(params interface{}) *utils.Response {
	delParams := &CustomResourceDeleteParams{}
	if err := utils.ConvertTypeByJson(params, delParams); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	gvr, err := c.gvr(&delParams.CustomResourceParam)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	for _, res := range delParams.Resources {
		err := c.client.Dynamic().Resource(*gvr).Namespace(res.Namespace).Delete(
			context.Background(), res.Name, metav1.DeleteOptions{})
		if err != nil {
			klog.Errorf("delete group %v namespace %s name %s error: %v", gvr, res.Namespace, res.Name, err)
			return &utils.Response{Code: code.DeleteError, Msg: fmt.Sprintf("delete %s error: %s", res.Name, err.Error())}
		}
	}
	return &utils.Response{Code: code.Success}
}

type UpdateCustomResourceParams struct {
	CustomResourceParam
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	YamlStr   string `json:"yaml"`
}

func (c *CustomResource) Update(params interface{}) *utils.Response {
	updateParams := &UpdateCustomResourceParams{}
	if err := utils.ConvertTypeByJson(params, &updateParams); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	gvr, err := c.gvr(&updateParams.CustomResourceParam)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	mapObj := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(updateParams.YamlStr), &mapObj); err != nil {
		klog.Error("Parse yaml error: ", err)
		return &utils.Response{Code: code.ParamsError, Msg: fmt.Sprintf("Parse yaml error: %s", err.Error())}
	}
	obj := &unstructured.Unstructured{Object: mapObj}
	obj.SetResourceVersion("")
	if updateParams.Namespace != "" {
		_, err = c.client.Dynamic().Resource(*gvr).Namespace(updateParams.Namespace).Update(
			context.Background(), obj, metav1.UpdateOptions{})
	} else {
		_, err = c.client.Dynamic().Resource(*gvr).Update(context.Background(), obj, metav1.UpdateOptions{})
	}
	if err != nil {
		klog.Error("Update error: ", err.Error())
		return &utils.Response{Code: code.UpdateError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success}
}
