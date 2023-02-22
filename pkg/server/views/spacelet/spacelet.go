package spacelet

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/spacelet"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"net/http"
	"time"
)

type SpaceletViews struct {
	models *model.Models
}

func NewSpaceletViews(config *config.ServerConfig) *SpaceletViews {
	return &SpaceletViews{models: config.Models}
}

// Register spacelet调用该接口进行注册入库
func (s *SpaceletViews) Register(c *gin.Context) {
	var req spacelet.RegisterRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, &utils.Response{Code: code.ParamsError, Msg: err.Error()})
		return
	}
	if req.HostIp == "" {
		req.HostIp = c.ClientIP()
	}
	if req.Port == 0 {
		req.Port = 7521
	}
	c.JSON(http.StatusOK, s.register(&req))
}

// register 对spacelet进行注册，调用spacelet token接口，配置认证，之后将该spacelet入库
func (s *SpaceletViews) register(req *spacelet.RegisterRequest) *utils.Response {
	if req.Hostname == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "param hostname is empty"}
	}
	httpcli, err := utils.NewHttpClient(fmt.Sprintf("http://%s:%d", req.HostIp, req.Port))
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	var tokenResp utils.Response
	spaceletObj, err := s.models.SpaceletManager.GetByIpPort(req.HostIp, req.Port)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	token := &spacelet.RegisterToken{Token: utils.CreateUUID()}
	if spaceletObj != nil {
		// 如果已注册了，token不变
		token.Token = spaceletObj.Token
	}
	// 调用spacelet token接口，配置认证
	if _, err = httpcli.Post("/v1/token", token, &tokenResp, utils.RequestOptions{}); err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	if !tokenResp.IsSuccess() {
		return &utils.Response{Code: code.RequestError, Msg: fmt.Sprintf("set spacelet token error: %s", tokenResp.Msg)}
	}
	if spaceletObj == nil {
		// 如果不存在则入库
		if _, err = s.models.SpaceletManager.Create(&types.Spacelet{
			Hostname:   req.Hostname,
			HostIp:     req.HostIp,
			Port:       req.Port,
			Token:      token.Token,
			Status:     "online",
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		}); err != nil {
			return &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
	}
	return &utils.Response{Code: code.Success}
}
