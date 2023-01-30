package resource

import (
	"github.com/kubespace/kubespace/pkg/kubernetes/config"
	"github.com/kubespace/kubespace/pkg/kubernetes/types"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var JobGVR = &schema.GroupVersionResource{
	Group:    "batch",
	Version:  "v1",
	Resource: "jobs",
}

type Job struct {
	*Resource
}

func NewJob(config *config.KubeConfig) *Job {
	p := &Job{}
	p.Resource = NewResource(config, types.JobType, JobGVR, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.PatchAction:  p.Patch,
		types.UpdateAction: p.Update,
		types.DeleteAction: p.Delete,
	}
	return p
}

type BuildJob struct {
	UID             string            `json:"uid"`
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	Completions     *int32            `json:"completions"`
	Active          int32             `json:"active"`
	Succeeded       int32             `json:"succeeded"`
	Failed          int32             `json:"failed"`
	ResourceVersion string            `json:"resource_version"`
	Conditions      []string          `json:"conditions"`
	NodeSelector    map[string]string `json:"node_selector"`
	Created         metav1.Time       `json:"created"`
}

func (j *Job) ToBuildJob(job *batchv1.Job) *BuildJob {
	if job == nil {
		return nil
	}
	var conditions []string
	for _, c := range job.Status.Conditions {
		if c.Status == corev1.ConditionTrue {
			conditions = append(conditions, string(c.Type))
		}
	}
	data := &BuildJob{
		UID:             string(job.UID),
		Name:            job.Name,
		Namespace:       job.Namespace,
		Completions:     job.Spec.Completions,
		Active:          job.Status.Active,
		Succeeded:       job.Status.Succeeded,
		Failed:          job.Status.Failed,
		Conditions:      conditions,
		NodeSelector:    job.Spec.Template.Spec.NodeSelector,
		Created:         job.CreationTimestamp,
		ResourceVersion: job.ResourceVersion,
	}

	return data
}

func (j *Job) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	ds := &batchv1.Job{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, ds); err != nil {
		return nil, err
	}
	return j.ToBuildJob(ds), nil
}
