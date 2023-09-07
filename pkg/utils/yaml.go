package utils

import (
	"bytes"
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
	"strconv"
	"strings"
	"text/template"
)

const (
	replaceYamlLeftDelim  = "<ReplacePath>"
	replaceYamlRightDelim = "</ReplacePath>"
)

// ReplaceYamlPathValue 替换YAML内容path的值，且不改变原yaml其余内容
// 如path=a.b，即将
// a:
//
//	b: xx
//
// xx值进行替换
func ReplaceYamlPathValue(yamlBytes []byte, pathVal map[string]string) ([]byte, error) {
	var err error
	var yamlNode yaml.Node
	if err = yaml.Unmarshal(yamlBytes, &yamlNode); err != nil {
		return nil, errors.New(code.UnMarshalError, err)
	}
	tplMap := make(map[string]string)
	pathIdx := 0
	for path, val := range pathVal {
		pathIdx += 1
		tplPathKey := fmt.Sprintf("key_%d", pathIdx)
		tplMap[tplPathKey] = val
		pathHolder := fmt.Sprintf("%s .%s %s", replaceYamlLeftDelim, tplPathKey, replaceYamlRightDelim)
		pathNode := yaml.Node{Kind: yaml.ScalarNode, Value: pathHolder}
		if err := setPathNode(&yamlNode, strings.Split(path, "."), pathNode); err != nil {
			klog.Warningf("set path=%s with value=%s error: %s", path, val, err.Error())
			continue
		}
	}

	// 根据原yaml缩进数，生成新的yaml
	indent := parseIndent(yamlNode.Content[0])
	var tpl bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&tpl)
	yamlEncoder.SetIndent(indent)
	if err = yamlEncoder.Encode(&yamlNode); err != nil {
		return nil, err
	}

	// 替换模板变量
	buf := new(bytes.Buffer)
	t := template.Must(template.New("test").Delims(replaceYamlLeftDelim, replaceYamlRightDelim).Parse(tpl.String()))
	if err := t.Execute(buf, tplMap); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func setPathNode(node *yaml.Node, path []string, value yaml.Node) error {
	if len(path) == 0 {
		*node = value
		return nil
	}
	key := path[0]
	rest := path[1:]

	switch node.Kind {
	case yaml.DocumentNode:
		return setPathNode(node.Content[0], path, value)
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			if node.Content[i].Value == key {
				return setPathNode(node.Content[i+1], rest, value)
			}
		}
		return fmt.Errorf("not find the path")
	case yaml.SequenceNode:
		index, err := strconv.Atoi(key)
		if err != nil {
			return err
		}
		if len(node.Content) <= index {
			return fmt.Errorf("not find the path")
		}
		return setPathNode(node.Content[index], rest, value)
	}

	return nil
}

// 解析原 yaml indent 空格数
func parseIndent(root *yaml.Node) int {
	for _, sub := range root.Content {
		if sub.Kind != yaml.MappingNode {
			continue
		}

		if len(sub.Content) == 0 {
			continue
		}

		// 取第一个二级节点的 Column - 1 作为 indent
		return sub.Column - 1
	}

	return 4
}
