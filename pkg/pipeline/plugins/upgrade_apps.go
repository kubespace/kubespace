package plugins

import (
	"encoding/json"
	"fmt"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"strings"
)

type UpgradeAppPlugin struct{}

func (p UpgradeAppPlugin) Execute(params *PluginParams) (interface{}, error) {
	upgrade, err := NewUpgradeApp(params)
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
	models  *model.Models
	params  *upgradeAppParams
	result  *upgradeAppResult
	project *types.Project
	*PluginLogger
}

func NewUpgradeApp(params *PluginParams) (*upgradeApp, error) {
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
		models:       params.Models,
		params:       &upgradeParams,
		result:       &upgradeAppResult{},
		PluginLogger: params.Logger,
	}, nil
}

func (u *upgradeApp) execute() error {
	project, err := u.models.ProjectManager.Get(u.params.Project)
	if err != nil {
		u.Log("获取工作空间 id=%s error: %s", u.params.Project, err.Error())
		return fmt.Errorf("get workspace %v error: %s", u.params.Project, err.Error())
	}
	u.Log("获取工作空间id=%s, name=%s", u.params.Project, project.Name)
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
	for _, appId := range u.params.Apps {
		if err = u.upgrade(images, appId, u.params.WithInstall); err != nil {
			return err
		}
	}
	return nil
}

func (u *upgradeApp) upgrade(images []string, appId uint, withInstall bool) error {
	return nil
}
