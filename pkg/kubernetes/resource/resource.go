package resource

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kubespace/kubespace/pkg/kubernetes/config"
	"github.com/kubespace/kubespace/pkg/kubernetes/kubeclient"
	kubetypes "github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"io"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
	"strings"
)

type ActionHandle func(params interface{}) *utils.Response

type listObjectProcess func(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error)

type OutWriter interface {
	Write(interface{}) error
	StopCh() <-chan struct{}
	Close()
}

type Resource struct {
	resType           string
	config            *config.KubeConfig
	client            kubeclient.Client
	actions           map[string]ActionHandle
	gvr               *schema.GroupVersionResource
	listObjectProcess listObjectProcess
}

func NewResource(
	config *config.KubeConfig,
	resType string,
	gvr *schema.GroupVersionResource,
	listObjectProcess listObjectProcess) *Resource {
	r := &Resource{
		resType:           resType,
		config:            config,
		client:            config.Client,
		actions:           make(map[string]ActionHandle),
		gvr:               gvr,
		listObjectProcess: listObjectProcess,
	}
	r.actions[kubetypes.ApplyAction] = r.Apply
	return r
}

func (r *Resource) Handle(action string, params interface{}) *utils.Response {
	if actionHandle, ok := r.actions[action]; ok {
		return actionHandle(params)
	} else {
		return &utils.Response{Code: code.GetError, Msg: fmt.Sprintf("not found %s %s action", r.resType, action)}
	}
}

type QueryParams struct {
	Kind               string                `json:"kind" form:"kind"`
	Name               string                `json:"name" form:"name"`
	Namespace          string                `json:"namespace" form:"namespace"`
	Output             string                `json:"output" form:"output"`
	LabelSelector      *metav1.LabelSelector `json:"label_selector" form:"label_selector"`
	Names              []string              `json:"names" form:"names"`
	ResourceVersion    string                `json:"resourceVersion" form:"resource_version"`
	OwnerReferenceKind string                `json:"owner_reference_kind" form:"owner_reference_kind"`
	OwnerReferenceName string                `json:"owner_reference_name" form:"owner_reference_name"`
	Process            bool                  `json:"process" form:"process"` // 是否对查询结果进行处理
}

func (r *Resource) listOptionsFromQuery(query *QueryParams) (options *metav1.ListOptions, err error) {
	labelSelector := labels.Everything()
	options = &metav1.ListOptions{}
	if query.LabelSelector != nil {
		labelSelector, err = metav1.LabelSelectorAsSelector(query.LabelSelector)
		if err != nil {
			return nil, err
		}
		options.LabelSelector = labelSelector.String()
	}
	if query.ResourceVersion != "" {
		options.ResourceVersion = query.ResourceVersion
	}
	if query.Name != "" {
		options.FieldSelector = "metadata.name=" + query.Name
	}
	return options, nil
}

func (r *Resource) List(params interface{}) *utils.Response {
	query := &QueryParams{}
	if err := utils.ConvertTypeByJson(params, query); err != nil {
		return &utils.Response{Code: code.UnMarshalError, Msg: err.Error()}
	}
	listOptions, err := r.listOptionsFromQuery(query)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	objects, err := r.client.Dynamic().Resource(*r.gvr).Namespace(query.Namespace).List(context.Background(), *listOptions)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	var data []interface{}
	for i := range objects.Items {
		if query.OwnerReferenceKind != "" && query.OwnerReferenceName != "" {
			find := false
			for _, ref := range objects.Items[i].GetOwnerReferences() {
				if ref.Kind == query.OwnerReferenceKind && ref.Name == query.OwnerReferenceName {
					find = true
					break
				}
			}
			if !find {
				continue
			}
		}
		if obj, err := r.listObjectProcess(query, &objects.Items[i]); err != nil {
			return &utils.Response{Code: code.RequestError, Msg: err.Error()}
		} else if obj != nil {
			data = append(data, obj)
		}
	}
	return &utils.Response{Code: code.Success, Msg: "Success", Data: data}
}

func (r *Resource) Get(params interface{}) *utils.Response {
	var query QueryParams
	if err := utils.ConvertTypeByJson(params, &query); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	var obj *unstructured.Unstructured
	var err error
	if query.Namespace != "" {
		obj, err = r.client.Dynamic().Resource(*r.gvr).Namespace(query.Namespace).Get(
			context.Background(), query.Name, metav1.GetOptions{})
	} else {
		obj, err = r.client.Dynamic().Resource(*r.gvr).Get(context.Background(), query.Name, metav1.GetOptions{})
	}
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	if query.Output == "yaml" {
		if yamlStr, err := yaml.Marshal(obj); err != nil {
			return &utils.Response{Code: code.MarshalError, Msg: err.Error()}
		} else {
			return &utils.Response{Code: code.Success, Data: string(yamlStr)}
		}
	}
	return &utils.Response{Code: code.Success, Data: obj}
}

func (r *Resource) Watch(params interface{}, writer OutWriter) *utils.Response {
	query := &QueryParams{}
	if err := utils.ConvertTypeByJson(params, query); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	listOptions, err := r.listOptionsFromQuery(query)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	if listOptions.ResourceVersion == "" {
		listObjs, err := r.client.Dynamic().Resource(*r.gvr).Namespace(query.Namespace).List(context.Background(), *listOptions)
		if err != nil {
			return &utils.Response{Code: code.RequestError, Msg: err.Error()}
		}
		listOptions.ResourceVersion = listObjs.GetResourceVersion()
	}
	watcher, err := r.client.Dynamic().Resource(*r.gvr).Namespace(query.Namespace).Watch(context.Background(), *listOptions)
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	go func() {
		defer writer.Close()
		defer watcher.Stop()
		for {
			select {
			case res, ok := <-watcher.ResultChan():
				if !ok {
					klog.Infof("watcher stopped")
					return
				}
				if query.Process {
					if object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(res.Object); err != nil {
						klog.Errorf("watch result to unstructured error: %s", err.Error())
						continue
					} else {
						if data, err := r.listObjectProcess(query, &unstructured.Unstructured{Object: object}); err != nil {
							klog.Errorf("watch resource to process error: %s", err.Error())
						} else {
							writer.Write(map[string]interface{}{
								"Type":   res.Type,
								"Object": data,
							})
						}
					}
				} else {
					writer.Write(res)
				}
			case <-writer.StopCh():
				klog.Infof("watch writer stopped")
				return
			}
		}
	}()
	return &utils.Response{Code: code.Success}
}

type DeleteParamResource struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type DeleteParams struct {
	Resources []*DeleteParamResource `json:"resources"`
}

func (r *Resource) Delete(params interface{}) *utils.Response {
	delParams := &DeleteParams{}
	if err := utils.ConvertTypeByJson(params, delParams); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	for _, res := range delParams.Resources {
		err := r.client.Dynamic().Resource(*r.gvr).Namespace(res.Namespace).Delete(
			context.Background(), res.Name, metav1.DeleteOptions{})
		if err != nil {
			klog.Errorf("delete group %v namespace %s name %s error: %v", r.gvr, res.Namespace, res.Name, err)
			return &utils.Response{Code: code.DeleteError, Msg: fmt.Sprintf("delete %s error: %s", res.Name, err.Error())}
		}
	}
	return &utils.Response{Code: code.Success}
}

type UpdateParams struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	YamlStr   string `json:"yaml"`
}

func (r *Resource) Update(params interface{}) *utils.Response {
	updateParams := &UpdateParams{}
	if err := utils.ConvertTypeByJson(params, &updateParams); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	mapObj := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(updateParams.YamlStr), &mapObj); err != nil {
		klog.Error("Parse yaml error: ", err)
		return &utils.Response{Code: code.ParamsError, Msg: fmt.Sprintf("Parse yaml error: %s", err.Error())}
	}
	obj := &unstructured.Unstructured{Object: mapObj}
	if obj.GetKind() != "Service" {
		obj.SetResourceVersion("")
	}
	var err error
	if updateParams.Namespace != "" {
		_, err = r.client.Dynamic().Resource(*r.gvr).Namespace(updateParams.Namespace).Update(context.Background(), obj, metav1.UpdateOptions{})
	} else {
		_, err = r.client.Dynamic().Resource(*r.gvr).Update(context.Background(), obj, metav1.UpdateOptions{})
	}
	if err != nil {
		klog.Error("Update error: ", err.Error())
		return &utils.Response{Code: code.UpdateError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success}
}

type PatchParams struct {
	Name      string      `json:"name"`
	Namespace string      `json:"namespace"`
	Data      interface{} `json:"data"`
}

func (r *Resource) Patch(params interface{}) *utils.Response {
	patchParams := &PatchParams{}
	if err := utils.ConvertTypeByJson(params, &patchParams); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	patchData, err := json.Marshal(patchParams.Data)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	if patchParams.Namespace != "" {
		_, err = r.client.Dynamic().Resource(*r.gvr).Namespace(patchParams.Namespace).Patch(context.Background(), patchParams.Name, types.MergePatchType, patchData, metav1.PatchOptions{})
	} else {
		_, err = r.client.Dynamic().Resource(*r.gvr).Patch(context.Background(), patchParams.Name, types.MergePatchType, patchData, metav1.PatchOptions{})
	}
	if err != nil {
		klog.Error("Patch error: ", err.Error())
		return &utils.Response{Code: code.UpdateError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success}
}

type ApplyParams struct {
	Create  bool   `json:"create"`
	YamlStr string `json:"yaml"`
}

func (r *Resource) Apply(params interface{}) *utils.Response {
	applyParams := &ApplyParams{}
	if err := utils.ConvertTypeByJson(params, applyParams); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	multidocReader := utilyaml.NewYAMLReader(bufio.NewReader(bytes.NewReader([]byte(applyParams.YamlStr))))
	var res []string
	applyErr := false
	for {
		buf, err := multidocReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return &utils.Response{Code: code.ParamsError, Msg: "read yaml error: " + err.Error()}
		}
		obj, dr, err := r.buildDynamicResourceClient(buf)
		if err != nil {
			applyErr = true
			res = append(res, err.Error())
			continue
		}

		// Create or Update
		if applyParams.Create {
			_, err = dr.Create(context.Background(), obj, metav1.CreateOptions{FieldManager: "kubespace-create"})
		} else {
			_, err = dr.Patch(context.Background(), obj.GetName(), types.ApplyPatchType, buf, metav1.PatchOptions{
				FieldManager: "kubespace-create",
			})
		}
		if err != nil {
			applyErr = true
			res = append(res, obj.GetKind()+"/"+obj.GetName()+" error : "+err.Error())
		} else {
			res = append(res, obj.GetKind()+"/"+obj.GetName()+" applied successful.")
		}
	}
	if applyErr {
		return &utils.Response{Code: code.RequestError, Msg: strings.Join(res, "\n")}
	}
	return &utils.Response{Code: code.Success, Msg: strings.Join(res, "\n")}
}

func (r *Resource) buildDynamicResourceClient(data []byte) (obj *unstructured.Unstructured, dr dynamic.ResourceInterface, err error) {
	// Decode YAML manifest into unstructured.Unstructured
	obj = &unstructured.Unstructured{}
	_, gvk, err := r.config.DecUnstructured.Decode(data, nil, obj)
	if err != nil {
		return obj, dr, fmt.Errorf("decode yaml failed. %s", err.Error())
	}

	// Find GVR
	mapping, err := r.config.RestMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return obj, dr, fmt.Errorf("mapping kind with version failed, %s", err.Error())
	}

	// Obtain REST interface for the GVR
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		if obj.GetNamespace() == "" {
			obj.SetNamespace("default")
		}
		// namespaced resources should specify the namespace
		dr = r.client.Dynamic().Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		// for cluster-wide resources
		dr = r.client.Dynamic().Resource(mapping.Resource)
	}
	return obj, dr, nil
}
