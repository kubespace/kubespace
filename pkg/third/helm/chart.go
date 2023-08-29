package helm

import (
	"encoding/base64"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"io/ioutil"
	"os"
	"path"
)

type ChartGeneration struct {
	// 是否需要修改Chart.yaml版本
	NeedModifyVersion bool
	// Chart版本
	PackageVersion string
	AppVersion     string
	// Chart文件
	Files map[string]interface{}

	// 文件内容是否base64
	Base64Encoded bool
}

func (c *ChartGeneration) GenerateChart() (chartDir, chartPath string, err error) {
	if c.NeedModifyVersion {
		// 更新chart文件版本
		if err = c.ModifyChartVersion(); err != nil {
			return
		}
	}
	chartDir, err = os.MkdirTemp("/tmp", "")
	if err != nil {
		err = errors.New(code.OsError, err)
		return
	}
	tmpChartDir := path.Join(chartDir, "charts")
	if err = c.WriteChartFiles(tmpChartDir, c.Files); err != nil {
		return
	}
	pack := action.NewPackage()
	pack.Destination = chartDir
	chartPath, err = pack.Run(tmpChartDir, nil)
	if err != nil {
		err = errors.New(code.HelmError, err)
		return
	}
	return
}

func (c *ChartGeneration) WriteChartFiles(chartDir string, files map[string]interface{}) (err error) {
	if err = os.MkdirAll(chartDir, 0755); err != nil {
		return err
	}
	for name, obj := range files {
		switch obj.(type) {
		case string:
			var fileBytes []byte
			if c.Base64Encoded {
				fileBytes, err = base64.StdEncoding.DecodeString(obj.(string))
				if err != nil {
					return errors.New(code.DecodeError, err)
				}
			} else {
				fileBytes = []byte(obj.(string))
			}
			if err = ioutil.WriteFile(path.Join(chartDir, name), fileBytes, 0644); err != nil {
				return errors.New(code.OsError, err)
			}
		case map[string]interface{}:
			if err = c.WriteChartFiles(path.Join(chartDir, name), obj.(map[string]interface{})); err != nil {
				return err
			}
		default:
			continue
		}
	}
	return nil
}

func (c *ChartGeneration) ModifyChartVersion() error {
	meta, ok := c.Files["Chart.yaml"]
	if !ok {
		return errors.New(code.DataNotExists, "未找到 Chart.yaml")
	}
	metaStr, ok := meta.(string)
	if !ok {
		return errors.New(code.ParseError, "Chart.yaml内容错误")
	}
	if c.Base64Encoded {
		metaBytes, err := base64.StdEncoding.DecodeString(metaStr)
		if err != nil {
			return errors.New(code.DecodeError, err)
		}
		metaStr = string(metaBytes)
	}

	metaObj := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(metaStr), &metaObj); err != nil {
		return err
	}
	metaObj["appVersion"] = c.AppVersion
	metaObj["version"] = c.PackageVersion
	metaBytes, err := yaml.Marshal(metaObj)
	if err != nil {
		return errors.New(code.MarshalError, err)
	}
	if c.Base64Encoded {
		metaStr = base64.StdEncoding.EncodeToString(metaBytes)
	} else {
		metaStr = string(metaBytes)
	}
	c.Files["Chart.yaml"] = metaStr
	return nil
}
