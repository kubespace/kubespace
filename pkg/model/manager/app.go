package manager

import (
	"encoding/base64"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/kubespace/kubespace/pkg/model/types"
	"helm.sh/helm/v3/pkg/chart/loader"
	"io/ioutil"
	"k8s.io/klog"
	"os"
	"path/filepath"
	"strings"
)

type AppManager struct {
	*CommonManager
}

func NewAppManager(redisClient *redis.Client) *AppManager {
	return &AppManager{
		CommonManager: NewCommonManager(redisClient, nil, "osp:app", true),
	}
}

func (a *AppManager) appKey(name, chartVersion string) string {
	return name + "-" + chartVersion
}

func (a *AppManager) Create(app *types.App) error {
	if err := a.CommonManager.Save(a.appKey(app.Name, app.ChartVersion), app, -1, true); err != nil {
		return err
	}

	return nil
}

func (a *AppManager) Get(name, chartVersion string) (*types.App, error) {
	app := &types.App{}
	if err := a.CommonManager.Get(a.appKey(name, chartVersion), app); err != nil {
		return nil, err
	}
	return app, nil
}

func (a *AppManager) List(filters map[string]interface{}) ([]*types.App, error) {
	dList, err := a.CommonManager.List(filters)
	if err != nil {
		return nil, err
	}
	jsonBody, err := json.Marshal(dList)
	if err != nil {
		return nil, err
	}
	var apps []*types.App

	if err := json.Unmarshal(jsonBody, &apps); err != nil {
		return nil, err
	}

	return apps, nil
}

func (a *AppManager) Init() {
	pwd, _ := os.Getwd()
	//获取文件或目录相关信息
	fileInfoList, err := ioutil.ReadDir(filepath.Join(pwd, "helm_apps"))
	if err != nil {
		klog.Errorf("read dir error: %v", err)
	}
	for i := range fileInfoList {
		klog.Infof("start load %s", fileInfoList[i].Name())
		if strings.HasSuffix(fileInfoList[i].Name(), ".tgz") {
			a.loadApp(filepath.Join(pwd, "helm_apps", fileInfoList[i].Name()))
		}
	}
}

func (a *AppManager) loadApp(filePath string) {
	charts, err := loader.LoadFile(filePath)
	if err != nil {
		klog.Errorf("load %s file error: %s", filePath, err)
		return
	}
	tarFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		klog.Errorf("read %s file error: %s", filePath, err)
		return
	}
	tarEncoded := base64.StdEncoding.EncodeToString(tarFile)

	app := &types.App{
		Name:         charts.Name(),
		ChartVersion: charts.Metadata.Version,
		AppVersion:   charts.AppVersion(),
		Icon:         charts.Metadata.Icon,
		Chart:        tarEncoded,
		Description:  charts.Metadata.Description,
	}
	err = a.Create(app)
	if err != nil {
		klog.Errorf("create app %s-%s error: %s", app.Name, app.ChartVersion, err)
	}
}
