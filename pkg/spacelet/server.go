package spacelet

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kubespace/kubespace/pkg/spacelet/pipeline_job"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"runtime"
)

type Server struct {
	config *Config
	engine *gin.Engine
}

func NewServer(config *Config) (*Server, error) {
	engine := gin.Default()
	s := &Server{config: config, engine: engine}
	// 配置token接口
	engine.POST("/v1/token", s.Token)

	// 统一token认证
	authGroup := engine.Group("/v1")
	authGroup.Use(s.AuthMiddleware())

	jobExecutor := pipeline_job.NewJobExecutor(config.DataDir, config.Client)
	authGroup.POST("/pipeline_job/execute", jobExecutor.Execute)
	authGroup.GET("/pipeline_job/status", jobExecutor.Status)
	authGroup.PUT("/pipeline_job/cleanup", jobExecutor.Cleanup)

	return s, nil
}

func (s *Server) Run() {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: s.engine,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}
	}()
	// 注册spacelet节点
	if err := s.Register(); err != nil {
		// 注册失败程序退出
		klog.Fatalf("register spacelet error: %s", err.Error())
	}
}

type RegisterRequest struct {
	Hostname string `json:"hostname"`
	HostIp   string `json:"hostip"`
	Port     int    `json:"port"`
}

// Register 启动spacelet后进行注册
func (s *Server) Register() error {
	hostname, _ := os.Hostname()
	var resp utils.Response
	// 调用spacelet注册接口
	if _, err := s.config.Client.Post("/api/v1/spacelet/register", &RegisterRequest{
		Hostname: hostname,
		HostIp:   s.config.HostIp,
		Port:     s.config.Port,
	}, &resp, utils.RequestOptions{}); err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("%s", resp.Msg)
	}
	return nil
}

type RegisterToken struct {
	Token string `json:"token"`
}

// Token 注册时kubespace server会调用该接口配置token
func (s *Server) Token(c *gin.Context) {
	var token RegisterToken
	if err := c.BindJSON(&token); err != nil {
		c.JSON(http.StatusBadRequest, &utils.Response{Code: code.ParamsError, Msg: err.Error()})
		return
	}
	if token.Token == "" {
		c.JSON(http.StatusBadRequest, &utils.Response{Code: code.ParamsError, Msg: "token is empty"})
		return
	}
	// 配置token，后续认证
	s.config.Token = token.Token
	klog.Infof("config token=%s", token.Token)
	c.JSON(http.StatusOK, &utils.Response{Code: code.Success})
}

// AuthMiddleware 调用spacelet接口需要通过token认证
func (s *Server) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				klog.Error("error: ", err)
				var buf [4096]byte
				n := runtime.Stack(buf[:], false)
				klog.Errorf("==> %s\n", string(buf[:n]))
				msg := fmt.Sprintf("%s", err)
				resp := &utils.Response{Code: code.UnknownError, Msg: msg}
				c.JSON(200, resp)
			}
		}()
		// 未配置token，返回报错
		if s.config.Token == "" {
			c.JSON(http.StatusUnauthorized, &utils.Response{Code: code.AuthError, Msg: "not register with token"})
			c.Abort()
			return
		}
		// 从header获取token
		token := c.Request.Header.Get("token")
		// 判断token跟配置是否相同
		if token != s.config.Token {
			c.JSON(http.StatusUnauthorized, &utils.Response{Code: code.AuthError, Msg: "token is incorrect"})
			c.Abort()
			return
		}
		c.Next()
	}
}
