package agent

import (
	"bytes"
	_ "embed"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
	"text/template"
)

//go:embed import.yaml.tpl
var importAgentYaml string

// 导入agent yaml
type importHandler struct {
	agentRepository string
	agentVersion    string
}

func ImportHandler(conf *config.ServerConfig) api.Handler {
	return &importHandler{
		agentRepository: conf.AgentRepository,
		agentVersion:    conf.AgentVersion,
	}
}

func (h *importHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	// 不需要认证鉴权
	return false, nil, nil
}

func (h *importHandler) Handle(c *api.Context) *utils.Response {
	token := c.Param("token")
	serverUrl := utils.RequestHost(c.Request)

	placeholders := map[string]interface{}{
		"AgentRepository": h.agentRepository,
		"AgentVersion":    h.agentVersion,
		"Token":           token,
		"ServerUrl":       serverUrl,
	}
	var buffer bytes.Buffer
	if err := template.Must(template.New("import_agent.yaml").Parse(importAgentYaml)).Execute(&buffer, placeholders); err != nil {
		return c.ResponseError(errors.New(code.ParseError, err))
	}
	c.String(200, buffer.String())
	return nil
}
