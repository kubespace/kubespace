package plugins

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
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

type upgradeAppResultImage struct {
	Path   map[string]string `json:"path"`
	Before string            `json:"before"`
	After  string            `json:"after"`
}

type upgradeAppResultApps struct {
	Id            uint                     `json:"id"`
	Name          string                   `json:"name"`
	UpgradeImages []*upgradeAppResultImage `json:"upgrade_images"`
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
		u.Log("获取工作空间id=%s失败: %s", u.params.Project, err.Error())
		return errors.New(code.DataNotExists, fmt.Sprintf("获取工作空间失败: %v", err))
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
		if err = u.upgrade(appId); err != nil {
			return err
		}
	}
	return nil
}

func (u *upgradeApp) upgrade(appId uint) error {
	app, err := u.models.AppManager.GetAppWithVersion(appId)
	if err != nil {
		u.Log("获取空间应用（id=%s)失败：%s", appId, err.Error())
		return errors.New(code.DataNotExists, fmt.Sprintf("获取应用id=%d失败：%v", appId, err))
	}
	u.Log("开始对应用「%s」进行镜像升级", app.Name)
	upgradeValues, upgradeImages, err := u.upgradeAppValues(app.AppVersion.Values)
	if err != nil {
		u.Log("升级应用镜像失败: %s", err.Error())
		return err
	}
	if len(upgradeImages) == 0 {
		return nil
	}
	if upgradeValues == "" {
		u.Log("升级应用镜像失败，values为空")
		return errors.New(code.PluginError, "upgrade values is empty")
	}
	app.AppVersion.Values = upgradeValues
	if err = u.models.AppVersionManager.UpdateAppVersion(app.AppVersion, "values"); err != nil {
		u.Log("更新应用「%s」版本values失败：%s", app.Name, err.Error())
		return err
	}
	u.Log("更新应用「%s」values成功:", app.Name)
	u.Log(upgradeValues)
	if u.params.WithInstall {
		appChart, err := u.models.AppVersionManager.GetChart(app.AppVersion.ChartPath)
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
		Id:            appId,
		Name:          app.Name,
		UpgradeImages: upgradeImages,
	})

	return nil
}

func (u *upgradeApp) upgradeAppValues(values string) (upgradeValues string, upgradeImages []*upgradeAppResultImage, err error) {
	var valuesDict = map[string]interface{}{}
	if err = yaml.Unmarshal([]byte(values), &valuesDict); err != nil {
		err = errors.New(code.MarshalError, err)
		return
	}
	u.Log("升级前values:\n%s", values)
	var path string
	upgradeImages = u.replaceValuesImage(valuesDict, path)
	if len(upgradeImages) == 0 {
		u.Log("未匹配到升级的镜像")
		return
	}

	replaceValuesPath := make(map[string]string)
	for _, img := range upgradeImages {
		for k, v := range img.Path {
			replaceValuesPath[k] = v
		}
	}

	replacedValuesBytes, err := utils.ReplaceYamlPathValue([]byte(values), replaceValuesPath)
	if err != nil {
		return
	}
	upgradeValues = string(replacedValuesBytes)

	return
}

func (u *upgradeApp) appendPath(path, append string) string {
	if path == "" {
		return append
	}
	return path + "." + append
}

func (u *upgradeApp) replaceValuesImage(values map[string]interface{}, path string) []*upgradeAppResultImage {
	klog.Infof("visiting: %s", path)
	imageValueKeyFuncMap := map[string]matchImageFunc{
		"image":      u.matchImageUpgrade,
		"repository": u.matchRepositoryUpgrade,
	}
	var upgradeImages []*upgradeAppResultImage
	for k, v := range values {
		switch val := v.(type) {
		case map[string]interface{}:
			upgradeImages = append(upgradeImages, u.replaceValuesImage(val, u.appendPath(path, k))...)
		case string:
			matchFunc, ok := imageValueKeyFuncMap[k]
			if !ok {
				continue
			}
			if upgradeImage := matchFunc(values, path); upgradeImage != nil {
				upgradeImages = append(upgradeImages, upgradeImage)
			}
		}
	}
	return upgradeImages
}

type matchImageFunc func(value map[string]interface{}, path string) *upgradeAppResultImage

func (u *upgradeApp) matchImageUpgrade(value map[string]interface{}, path string) (upgradeImage *upgradeAppResultImage) {
	oriImage := value["image"].(string)
	var ok bool
	tagObj, hasTag := value["tag"]
	oriTag := ""
	if hasTag {
		oriTag, ok = tagObj.(string)
		if !ok {
			return
		}
		oriImage += ":" + oriTag
	}
	oriImageName := utils.GetImageName(oriImage)

	for _, destImage := range u.images {
		if !strings.Contains(destImage, oriImageName+":") {
			continue
		}
		replacePath := make(map[string]string)
		if !hasTag {
			replacePath[u.appendPath(path, "image")] = destImage
		} else {
			registry, imgName, tag := utils.ParseImageName(destImage, true)
			replacePath[u.appendPath(path, "image")] = registry + "/" + imgName
			replacePath[u.appendPath(path, "tag")] = fmt.Sprintf(`"%s"`, tag)
		}
		u.Log("应用原镜像「%s」，升级为「%s」", oriImage, destImage)
		return &upgradeAppResultImage{
			Path:   replacePath,
			Before: oriImage,
			After:  destImage,
		}
	}
	return
}

func (u *upgradeApp) matchRepositoryUpgrade(value map[string]interface{}, path string) (upgradeImage *upgradeAppResultImage) {
	var registry, repository, tag, oriImage, oriImageName string
	var ok, hasRegistry bool
	if repository, ok = utils.GetMapStringValue(value, "repository"); !ok {
		return
	}
	if tag, ok = utils.GetMapStringValue(value, "tag"); !ok {
		return
	}
	registry, hasRegistry = utils.GetMapStringValue(value, "registry")
	if hasRegistry {
		oriImage = fmt.Sprintf("%s/%s:%s", registry, repository, tag)
	} else {
		oriImage = fmt.Sprintf("%s:%s", repository, tag)
	}
	oriImageName = utils.GetImageName(oriImage)
	for _, destImage := range u.images {
		if !strings.Contains(destImage, oriImageName+":") {
			continue
		}

		destRegistry, destImageName, destTag := utils.ParseImageName(destImage, true)

		replacePath := map[string]string{
			u.appendPath(path, "tag"): fmt.Sprintf(`"%s"`, destTag),
		}
		if hasRegistry {
			replacePath[u.appendPath(path, "registry")] = destRegistry
			replacePath[u.appendPath(path, "repository")] = destImageName
		} else {
			replacePath[u.appendPath(path, "repository")] = destRegistry + "/" + destImageName
		}

		u.Log("应用原镜像「%s」，替换升级为「%s」", oriImage, destImage)
		return &upgradeAppResultImage{
			Path:   replacePath,
			Before: oriImage,
			After:  destImage,
		}
	}
	return
}
