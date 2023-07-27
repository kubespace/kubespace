package spacelet

import (
	"bytes"
	_ "embed"
	"github.com/gin-gonic/gin"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"io"
	"net/http"
	"text/template"
)

//go:embed install_spacelet.sh.tpl
var installAgentShellTpl string

type InstallSpaceletRequest struct {
	OS         string `form:"os,default=linux"`
	Arch       string `form:"arch,default=amd64"`
	Port       string `form:"port,default=7520"`
	ServerHost string `form:"server_host"`
	DataDir    string `form:"data_dir,default=/data"`
	HostIp     string `form:"host_ip"`
}

func (s *SpaceletViews) InstallSpacelet(c *gin.Context) {
	var req InstallSpaceletRequest
	if err := c.BindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, &utils.Response{Code: code.ParamsError, Msg: err.Error()})
		return
	}
	serverHost := req.ServerHost
	if serverHost == "" {
		serverHost = utils.RequestHost(c.Request)
	}
	placeholders := map[string]interface{}{
		"ServerHost": serverHost,
		"OS":         req.OS,
		"Arch":       req.Arch,
		"Port":       req.Port,
		"DataDir":    req.DataDir,
		"HostIp":     req.HostIp,
	}
	var buffer bytes.Buffer
	if err := template.Must(template.New("install_agent.sh").Parse(installAgentShellTpl)).Execute(&buffer, placeholders); err != nil {
		c.JSON(200, gin.H{
			"code": "ParseError",
			"msg":  err.Error(),
		})
		return
	}
	io.Copy(c.Writer, &buffer)
}
