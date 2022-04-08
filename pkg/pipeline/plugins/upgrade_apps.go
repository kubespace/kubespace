package plugins

import (
	"encoding/json"
	"fmt"
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog"
	"sigs.k8s.io/yaml"
	"strings"
)

type UpgradeAppPlugin struct {
	*model.Models
	*kube_resource.KubeResources
}

func (p UpgradeAppPlugin) Execute(params *PluginParams) (interface{}, error) {
	upgrade, err := NewUpgradeApp(params, p.Models, p.KubeResources)
	if err != nil {
		return nil, err
	}
	err = upgrade.execute()
	if err != nil {
		return nil, err
	}
	return upgrade.result, nil
}

type upgradeAppParams struct {
	Project     uint   `json:"project"`
	Apps        []uint `json:"apps"`
	WithInstall bool   `json:"with_install"`
	Images      string `json:"images"`
}

type upgradeAppResultApps struct {
	Id   uint   `json:"id"`
	Name string `json:"name"`
}

type upgradeAppResult struct {
	Project string                  `json:"project"`
	Apps    []*upgradeAppResultApps `json:"apps"`
}

type upgradeApp struct {
	models        *model.Models
	kubeResources *kube_resource.KubeResources
	params        *upgradeAppParams
	images        []string
	result        *upgradeAppResult
	project       *types.Project
	*PluginLogger
}

func NewUpgradeApp(params *PluginParams, models *model.Models, kr *kube_resource.KubeResources) (*upgradeApp, error) {
	var upgradeParams upgradeAppParams
	marshalParams, err := json.Marshal(params.Params)
	if err != nil {
		return nil, fmt.Errorf("marshal params error: %s", err.Error())
	}
	err = json.Unmarshal(marshalParams, &upgradeParams)
	if err != nil {
		return nil, fmt.Errorf("marshal upgrade app params error: %s", err.Error())
	}
	return &upgradeApp{
		models:        models,
		kubeResources: kr,
		params:        &upgradeParams,
		result:        &upgradeAppResult{},
		PluginLogger:  params.Logger,
	}, nil
}

func (u *upgradeApp) execute() error {
	project, err := u.models.ProjectManager.Get(u.params.Project)
	if err != nil {
		u.Log("获取工作空间 id=%s error: %s", u.params.Project, err.Error())
		return fmt.Errorf("get workspace %v error: %s", u.params.Project, err.Error())
	}
	u.Log("工作空间id=%s, name=%s", u.params.Project, project.Name)
	u.result.Project = project.Name
	u.project = project
	if u.params.Images == "" {
		u.Log("要升级的镜像列表参数为空")
		return nil
	}
	if len(u.params.Apps) == 0 {
		u.Log("要升级的应用列表参数为空")
		return nil
	}
	var images []string
	imageStrs := strings.Split(u.params.Images, ",")
	for _, imageStr := range imageStrs {
		var imgs []string
		err = json.Unmarshal([]byte(imageStr), &imgs)
		if err != nil {
			u.Log("解析镜像列表（%s）失败: %s", imageStrs, err.Error())
			return err
		}
		for _, img := range imgs {
			images = append(images, img)
		}
	}
	u.Log("升级的镜像列表：%v", images)
	u.images = images
	for _, appId := range u.params.Apps {
		if err = u.upgrade(appId, u.params.WithInstall); err != nil {
			return err
		}
	}
	return nil
}

func (u *upgradeApp) upgrade(appId uint, withInstall bool) error {
	app, err := u.models.ProjectAppManager.GetProjectApp(appId)
	if err != nil {
		u.Log("获取空间应用（id=%s)失败：%s", appId, err.Error())
		return err
	}
	u.Log("开始对应用「%s」进行镜像升级", app.Name)
	upgradeValues, err := u.upgradeAppValues(app.AppVersion.Values)
	if err != nil {
		u.Log("")
	}
	if upgradeValues != "" {
		app.AppVersion.Values = upgradeValues
		if err = u.models.ProjectAppVersionManager.UpdateAppVersion(app.AppVersion, "values"); err != nil {
			u.Log("更新应用「%s」版本values失败：%s", app.Name, err.Error())
			return err
		}
		u.Log("更新应用「%s」values成功", app.Name)
		if withInstall {
			installParams := map[string]interface{}{
				"name":       app.Name,
				"namespace":  u.project.Namespace,
				"chart_path": app.AppVersion.ChartPath,
				"values":     upgradeValues,
			}
			var resp *utils.Response
			if app.Status != types.AppStatusUninstall {
				u.Log("开始对应用进行升级")
				resp = u.kubeResources.Helm.UpdateObj(u.project.ClusterId, installParams)
			} else {
				u.Log("开始对应用进行安装")
				resp = u.kubeResources.Helm.Create(u.project.ClusterId, installParams)
			}
			if resp.IsSuccess() {
				u.Log("安装/升级成功")
			} else {
				u.Log("安装/升级失败：%v", resp)
				return fmt.Errorf("升级应用失败：%s", resp.Msg)
			}
		}
		u.result.Apps = append(u.result.Apps, &upgradeAppResultApps{
			Id:   appId,
			Name: app.Name,
		})
	}
	return nil
}

func (u *upgradeApp) upgradeAppValues(values string) (upgradeValues string, err error) {
	var valuesDict = map[string]interface{}{}
	if err = yaml.Unmarshal([]byte(values), valuesDict); err != nil {
		return
	}
	if err = u.replaceValuesImage(valuesDict); err != nil {
		return
	}
	var valBytes []byte
	if valBytes, err = yaml.Marshal(valuesDict); err != nil {
		return
	}
	upgradeValues = string(valBytes)
	return
}

func (u *upgradeApp) replaceValuesImage(values map[string]interface{}) error {
	klog.Infof("visiting: %v", values)
	for valKey, val := range values {
		switch v := val.(type) {
		case map[string]interface{}:
			err := u.replaceValuesImage(v)
			if err != nil {
				return err
			}
		case string:
			if valKey == "image" {
				u.Log("应用原镜像「%s」，升级为「%s」", v, u.images[0])
				values[valKey] = u.images[0]
			}
		}
	}
	return nil
}
