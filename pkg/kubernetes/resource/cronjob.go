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
	"strconv"
)

var CronJobGVR = &schema.GroupVersionResource{
	Group:    "batch",
	Version:  "v1beta1",
	Resource: "cronjobs",
}

var CronJobV1GVR = &schema.GroupVersionResource{
	Group:    "batch",
	Version:  "v1",
	Resource: "cronjobs",
}

type CronJob struct {
	*Resource
}

func NewCronJob(config *config.KubeConfig) *CronJob {
	p := &CronJob{}
	gvr := CronJobGVR
	if config.Client.VersionGreaterThan(types.ServerVersion21) {
		gvr = CronJobV1GVR
	}
	p.Resource = NewResource(config, types.CronjobType, gvr, p.listObjectProcess)
	p.actions = map[string]ActionHandle{
		types.ListAction:   p.List,
		types.GetAction:    p.Get,
		types.PatchAction:  p.Patch,
		types.UpdateAction: p.Update,
		types.DeleteAction: p.Delete,
	}
	return p
}

type BuildCronJob struct {
	UID               string                   `json:"uid"`
	Name              string                   `json:"name"`
	Namespace         string                   `json:"namespace"`
	Active            []corev1.ObjectReference `json:"active"`
	LastScheduleTime  *metav1.Time             `json:"last_schedule_time"`
	Schedule          string                   `json:"schedule"`
	ConcurrencyPolicy string                   `json:"concurrency_policy"`
	ResourceVersion   string                   `json:"resource_version"`
	Suspend           string                   `json:"suspend"`
	Created           metav1.Time              `json:"created"`
}

func (c *CronJob) ToBuildCronJob(cronjob *batchv1.CronJob) *BuildCronJob {
	if cronjob == nil {
		return nil
	}
	data := &BuildCronJob{
		UID:               string(cronjob.UID),
		Name:              cronjob.Name,
		Namespace:         cronjob.Namespace,
		Active:            cronjob.Status.Active,
		LastScheduleTime:  cronjob.Status.LastScheduleTime,
		Schedule:          cronjob.Spec.Schedule,
		ConcurrencyPolicy: string(cronjob.Spec.ConcurrencyPolicy),
		Suspend:           strconv.FormatBool(*cronjob.Spec.Suspend),
		Created:           cronjob.CreationTimestamp,
		ResourceVersion:   cronjob.ResourceVersion,
	}

	return data
}

func (c *CronJob) listObjectProcess(query *QueryParams, obj *unstructured.Unstructured) (interface{}, error) {
	ds := &batchv1.CronJob{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, ds); err != nil {
		return nil, err
	}
	return c.ToBuildCronJob(ds), nil
}
