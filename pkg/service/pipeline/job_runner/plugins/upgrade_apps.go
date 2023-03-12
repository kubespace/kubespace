package plugins

import (
	"fmt"
	kubetypes "github.com/kubespace/kubespace/pkg/kubernetes/types"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/service/cluster"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
	"strings"
)

type UpgradeAppPlugin struct {
	*model.Models
	KubeClient *cluster.KubeClient
}

func (p UpgradeAppPlugin) Executor(params *ExecutorParams) (Executor, error) {
	return newUpgradeApp(params, p.Models, p.KubeClient)
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
	Logger
	models     *model.Models
	kubeClient *cluster.KubeClient
	params     *upgradeAppParams
	images     map[string]string
	result     *upgradeAppResult
	project    *types.Project
}

func newUpgradeApp(params *ExecutorParams, models *model.Models, kubeClient *cluster.KubeClient) (*upgradeApp, error) {
	var upgradeParams upgradeAppParams
	if err := utils.ConvertTypeByJson(params.Params, &upgradeParams); err != nil {
		params.Logger.Log("插件参数：%v", params.Params)
		return nil, fmt.Errorf("插件参数错误: %s", err.Error())
	}
	return &upgradeApp{
		models:     models,
		kubeClient: kubeClient,
		params:     &upgradeParams,
		result:     &upgradeAppResult{},
		Logger:     params.Logger,
	}, nil
}

func (u *upgradeApp) Execute() (interface{}, error) {
	if err := u.execute(); err != nil {
		return nil, err
	}
	return u.result, nil
}

func (u *upgradeApp) Cancel() error {
	return nil
}

func (u *upgradeApp) execute() error {
	//u.Log("插件执行参数：%+v", *u.params)
	project, err := u.models.ProjectManager.Get(u.params.Project)
	if err != nil {
		u.Log("获取工作空间 id=%s error: %s", u.params.Project, err.Error())
		return fmt.Errorf("get workspace %v error: %s", u.params.Project, err.Error())
	}
	u.Log("工作空间id=%d, name=%s", u.params.Project, project.Name)
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
	u.images = stringToImage(u.params.Images)
	u.Log("升级的镜像列表：%v", u.params.Images)
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
	replaced, upgradeValues, err := u.upgradeAppValues(app.AppVersion.Values)
	if err != nil {
		u.Log("替换升级镜像失败: %s", err.Error())
		return err
	}
	if !replaced {
		return nil
	}
	if upgradeValues != "" {
		app.AppVersion.Values = upgradeValues
		if err = u.models.ProjectAppVersionManager.UpdateAppVersion(app.AppVersion, "values"); err != nil {
			u.Log("更新应用「%s」版本values失败：%s", app.Name, err.Error())
			return err
		}
		u.Log("更新应用「%s」values成功", app.Name)
		if withInstall {
			appChart, err := u.models.ProjectAppVersionManager.GetAppVersionChart(app.AppVersion.ChartPath)
			if err != nil {
				u.Log("not found chart path=%s", app.AppVersion.ChartPath)
				return err
			}
			installParams := map[string]interface{}{
				"name":        app.Name,
				"namespace":   u.project.Namespace,
				"chart_bytes": appChart.Content,
				"values":      upgradeValues,
			}
			var resp *utils.Response
			if app.Status != types.AppStatusUninstall {
				u.Log("开始对应用进行升级")
				resp = u.kubeClient.Update(u.project.ClusterId, kubetypes.HelmType, installParams)
			} else {
				u.Log("开始对应用进行安装")
				resp = u.kubeClient.Create(u.project.ClusterId, kubetypes.HelmType, installParams)
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

func (u *upgradeApp) upgradeAppValues(values string) (replaced bool, upgradeValues string, err error) {
	var valuesDict = map[string]interface{}{}
	if err = yaml.Unmarshal([]byte(values), &valuesDict); err != nil {
		return
	}
	u.Log("升级前values:\n%s", values)
	if replaced, err = u.replaceValuesImage(valuesDict); err != nil {
		return
	}
	if replaced {
		var valBytes []byte
		if valBytes, err = yaml.Marshal(valuesDict); err != nil {
			return
		}
		upgradeValues = string(valBytes)
		u.Log("升级后values:\n%s", upgradeValues)
	} else {
		u.Log("未匹配到升级的镜像")
	}
	return
}

func (u *upgradeApp) replaceValuesImage(values map[string]interface{}) (bool, error) {
	klog.Infof("visiting: %v", values)
	replaced := false
	for valKey, val := range values {
		switch v := val.(type) {
		case map[string]interface{}:
			flag, err := u.replaceValuesImage(v)
			if err != nil {
				return false, err
			}
			if flag {
				replaced = true
			}
		case string:
			if valKey == "image" {
				if u.matchImageReplace(values) {
					replaced = true
				}
			} else if valKey == "repository" {
				if u.matchRepositoryReplace(values) {
					replaced = true
				}
			}
		}
	}
	return replaced, nil
}

func (u *upgradeApp) matchImageReplace(value map[string]interface{}) bool {
	imageObj, ok := value["image"]
	if !ok {
		return false
	}
	tagObj, hasTag := value["tag"]
	oriTag := ""
	if hasTag {
		oriTag, ok = tagObj.(string)
	}
	oriImage, ok := imageObj.(string)
	if !ok {
		return false
	}
	image := oriImage
	if hasTag {
		oriImage += ":" + oriTag
	} else {
		image = strings.Split(image, ":")[0]
	}

	if strings.Contains(strings.Split(image, "/")[0], ".") {
		image = strings.Join(strings.Split(image, "/")[1:], "/")
	}
	for _, destImage := range u.images {
		if strings.Contains(destImage, image+":") {
			if hasTag {
				destNoTag := strings.Split(destImage, ":")[0]
				tag := "latest"
				if len(strings.Split(destImage, ":")) == 2 {
					tag = strings.Split(destImage, ":")[1]
				}
				value["image"] = destNoTag
				value["tag"] = tag
			} else {
				value["image"] = destImage
			}
			u.Log("应用原镜像「%s」，替换升级为「%s」", oriImage, destImage)
			return true
		}
	}
	return false
}

func (u *upgradeApp) matchRepositoryReplace(value map[string]interface{}) bool {
	repositoryObj, ok := value["repository"]
	if !ok {
		return false
	}
	repository, ok := repositoryObj.(string)
	if !ok {
		return false
	}
	if _, ok = value["tag"]; !ok {
		return false
	}
	oriImage := repository
	if _, ok = value["registry"]; !ok {
		if strings.Contains(strings.Split(repository, "/")[0], ".") {
			oriImage = strings.Join(strings.Split(repository, "/")[1:], "/")
		}
	}
	for _, image := range u.images {
		destImage := image
		destTag := "latest"
		if len(strings.Split(destImage, ":")) == 2 {
			destTag = strings.Split(destImage, ":")[1]
		} else {
			destImage = destImage + ":latest"
		}
		if strings.Contains(destImage, oriImage+":") {
			u.Log("应用原镜像「%v/%v:%v」，替换升级为「%s」", value["registry"], value["repository"], value["tag"], destImage)

			value["tag"] = destTag
			destImage = strings.Split(destImage, ":")[0]

			if _, ok = value["registry"]; ok {
				destRegistry := "docker.io"
				destImageSplit := strings.Split(destImage, "/")
				if strings.Contains(destImageSplit[0], ".") {
					destRegistry = destImageSplit[0]
					destImage = strings.Join(destImageSplit[1:], "/")
				}
				value["registry"] = destRegistry
			}
			value["repository"] = destImage

			return true
		}
	}
	return false
}
