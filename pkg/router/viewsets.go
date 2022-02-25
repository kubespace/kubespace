package router

import (
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/views"
	"github.com/kubespace/kubespace/pkg/views/kube_views"
	"github.com/kubespace/kubespace/pkg/views/pipeline_views"
	"github.com/kubespace/kubespace/pkg/views/project_views"
	"github.com/kubespace/kubespace/pkg/views/settings_views"
)

type ViewSets map[string][]*views.View

func NewViewSets(kr *kube_resource.KubeResources, models *model.Models) *ViewSets {
	cluster := views.NewCluster(models, kr)
	user := views.NewUser(models)
	settingsRole := views.NewRole(models)

	pods := kube_views.NewPod(kr)
	event := kube_views.NewEvent(kr)
	namespace := kube_views.NewNamespace(kr)
	deployment := kube_views.NewDeployment(kr)
	node := kube_views.NewNode(kr)
	statefulset := kube_views.NewStatefulset(kr)
	daemonset := kube_views.NewDaemonset(kr)
	cronjob := kube_views.NewCronjob(kr)
	job := kube_views.NewJob(kr)
	service := kube_views.NewService(kr)
	endpoints := kube_views.NewEndpoint(kr)
	ingress := kube_views.NewIngress(kr)
	networkpolicy := kube_views.NewNetworkPolicy(kr)
	serviceaccount := kube_views.NewServiceAccount(kr)
	rolebinding := kube_views.NewRolebinding(kr)
	role := kube_views.NewRole(kr)
	configmap := kube_views.NewConfigMap(kr)
	secret := kube_views.NewSecret(kr)
	hpa := kube_views.NewHpa(kr)
	pvc := kube_views.NewPvc(kr)
	pv := kube_views.NewPV(kr)
	storageclass := kube_views.NewStorageClass(kr)
	helm := kube_views.NewHelm(kr, models)
	crd := kube_views.NewCrd(kr)

	pipelineWorkspace := pipeline_views.NewPipelineWorkspace(models)
	pipeline := pipeline_views.NewPipeline(models)
	pipelineRun := pipeline_views.NewPipelineRun(models)

	settingsSecret := settings_views.NewSettingsSecret(models)
	imageRegistry := settings_views.NewImageRegistry(models)

	projectWorkspace := project_views.NewProject(models)
	projectApps := project_views.NewProjectApp(kr, models)

	viewsets := &ViewSets{
		"cluster":        cluster.Views,
		"user":           user.Views,
		"settings_role":  settingsRole.Views,
		"pods":           pods.Views,
		"event":          event.Views,
		"namespace":      namespace.Views,
		"deployment":     deployment.Views,
		"nodes":          node.Views,
		"statefulset":    statefulset.Views,
		"daemonset":      daemonset.Views,
		"cronjob":        cronjob.Views,
		"job":            job.Views,
		"service":        service.Views,
		"endpoints":      endpoints.Views,
		"ingress":        ingress.Views,
		"networkpolicy":  networkpolicy.Views,
		"serviceaccount": serviceaccount.Views,
		"rolebinding":    rolebinding.Views,
		"role":           role.Views,
		"configmap":      configmap.Views,
		"secret":         secret.Views,
		"hpa":            hpa.Views,
		"pvc":            pvc.Views,
		"pv":             pv.Views,
		"storageclass":   storageclass.Views,
		"helm":           helm.Views,
		"crd":            crd.Views,

		"pipeline/workspace": pipelineWorkspace.Views,
		"pipeline/pipeline":  pipeline.Views,
		"pipeline/build":     pipelineRun.Views,

		"settings/secret":         settingsSecret.Views,
		"settings/image_registry": imageRegistry.Views,

		"project/workspace": projectWorkspace.Views,
		"project/apps":      projectApps.Views,
	}

	return viewsets
}
