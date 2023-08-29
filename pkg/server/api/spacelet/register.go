package spacelet

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/spacelet"
	"github.com/kubespace/kubespace/pkg/third/httpclient"
	"github.com/kubespace/kubespace/pkg/utils"
	"net/http"
	"time"
)

type registerHandler struct {
	models *model.Models
}

func RegisterHandler(conf *config.ServerConfig) api.Handler {
	return &registerHandler{models: conf.Models}
}

func (h *registerHandler) Auth(c *api.Context) (bool, *api.AuthPerm, error) {
	return false, nil, nil
}

func (h *registerHandler) Handle(c *api.Context) *utils.Response {
	var req spacelet.RegisterRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, c.ResponseError(errors.New(code.ParamsError, err)))
		return nil
	}
	if req.HostIp == "" {
		req.HostIp = c.ClientIP()
	}
	if req.Port == 0 {
		req.Port = 7521
	}
	return h.register(&req)
}

// register 对spacelet进行注册，调用spacelet token接口，配置认证，之后将该spacelet入库
func (h *registerHandler) register(req *spacelet.RegisterRequest) *utils.Response {
	if req.Hostname == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "param hostname is empty"}
	}
	httpcli, err := httpclient.NewHttpClient(fmt.Sprintf("http://%s:%d", req.HostIp, req.Port))
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	var tokenResp utils.Response
	spaceletObj, err := h.models.SpaceletManager.GetByIpPort(req.HostIp, req.Port)
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	token := &spacelet.RegisterToken{Token: utils.CreateUUID()}
	if spaceletObj != nil {
		// 如果已注册了，token不变
		token.Token = spaceletObj.Token
	}
	// 调用spacelet token接口，配置认证
	if _, err = httpcli.Post("/v1/token", token, &tokenResp, httpclient.RequestOptions{}); err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	if !tokenResp.IsSuccess() {
		return &utils.Response{Code: code.RequestError, Msg: fmt.Sprintf("set spacelet token error: %s", tokenResp.Msg)}
	}
	if spaceletObj == nil {
		// 如果不存在则入库
		if _, err = h.models.SpaceletManager.Create(&types.Spacelet{
			Hostname:   req.Hostname,
			HostIp:     req.HostIp,
			Port:       req.Port,
			Token:      token.Token,
			Status:     types.SpaceletStatusOnline,
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		}); err != nil {
			return &utils.Response{Code: code.DBError, Msg: err.Error()}
		}
	}
	return &utils.Response{Code: code.Success}
}
