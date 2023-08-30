package spacelet

import (
	"bytes"
	_ "embed"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
	"io"
	"text/template"
)

//go:embed install_spacelet.sh.tpl
var installAgentShellTpl string

type installHandler struct {
	models *model.Models
}

func InstallHandler(conf *config.ServerConfig) api.Handler {
	return &installHandler{models: conf.Models}
}

type InstallSpaceletForm struct {
	OS         string `form:"os,default=linux"`
	Arch       string `form:"arch,default=amd64"`
	Port       string `form:"port,default=7520"`
	ServerHost string `form:"server_host"`
	DataDir    string `form:"data_dir,default=/data"`
	HostIp     string `form:"host_ip"`
}

func (h *installHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return false, nil, nil
}

func (h *installHandler) Handle(c *api.Context) *utils.Response {
	var form InstallSpaceletForm
	if err := c.BindQuery(&form); err != nil {
		c.ResponseError(errors.New(code.ParamsError, err))
	}
	serverHost := form.ServerHost
	if serverHost == "" {
		serverHost = utils.RequestHost(c.Request)
	}
	placeholders := map[string]interface{}{
		"ServerHost": serverHost,
		"OS":         form.OS,
		"Arch":       form.Arch,
		"Port":       form.Port,
		"DataDir":    form.DataDir,
		"HostIp":     form.HostIp,
	}
	var buffer bytes.Buffer
	if err := template.Must(template.New("install_agent.sh").Parse(installAgentShellTpl)).Execute(&buffer, placeholders); err != nil {
		return c.ResponseError(errors.New(code.ParseError, err))
	}
	io.Copy(c.Writer, &buffer)
	return nil
}
