package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/cluster"
	"github.com/kubespace/kubespace/pkg/server/views/user"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"html/template"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net/http"
	"runtime"
	"strings"
)

type Router struct {
	*gin.Engine
}

func NewRouter(conf *config.ServerConfig) (*Router, error) {
	//redisOptions := serverConfig.RedisOptions
	models := conf.Models

	engine := gin.Default()

	engine.Use(LocalMiddleware())

	indexHtml, _ := ioutil.ReadAll(Assets.Files["/index.html"])
	t, _ := template.New("").New("/index.html").Parse(string(indexHtml))
	engine.SetHTMLTemplate(t)
	engine.StaticFS("/static", Assets)
	engine.StaticFile("/favicon.ico", "./favicon.ico")

	engine.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "/index.html", nil)
	})

	engine.GET("/ui/*path", func(c *gin.Context) {
		c.HTML(http.StatusOK, "/index.html", nil)
	})

	// 统一认证的api接口
	apiGroup := engine.Group("/api/v1")
	viewsets := NewViewSets(conf)
	for group, vs := range *viewsets {
		g := apiGroup.Group(group)
		for _, v := range vs {
			g.Handle(v.Method, v.Path, apiWrapper(models, v.Handler))
		}
	}

	// 登录登出接口
	loginView := user.NewLogin(models)
	apiGroup.POST("/login", loginView.Login)
	apiGroup.GET("/has_admin", loginView.HasAdmin)
	apiGroup.POST("/admin", loginView.CreateAdmin)
	apiGroup.POST("/logout", loginView.Logout)

	// kube-agent访问接口
	agentView := cluster.NewAgentViews(conf)
	apiGroup.GET("/agent/connect", agentView.Connect)
	apiGroup.GET("/agent/response", agentView.Response)
	apiGroup.GET("/agent/yaml", agentView.AgentYaml)
	return &Router{
		Engine: engine,
	}, nil
}

func apiWrapper(m *model.Models, handler views.ViewHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		authRes := auth(m, c)
		if !authRes.IsSuccess() {
			c.JSON(401, authRes)
		} else {
			context := &views.Context{Context: c, User: authRes.Data.(*types.User)}
			res := handler(context)
			if res != nil {
				c.JSON(200, res)
			}
		}
	}
}

func auth(m *model.Models, c *gin.Context) *utils.Response {
	resp := utils.Response{Code: code.Success}
	token := c.DefaultQuery("token", "")
	if token == "" {
		token = c.Request.Header.Get("Authorization")
		if s := strings.Split(token, " "); len(s) == 2 {
			token = s[1]
		}
	}
	if token == "" {
		tokenCookie, err := c.Request.Cookie("osp-token")
		if err == nil {
			token = tokenCookie.Value
		}
	}
	if token == "" {
		resp.Code = code.ParamsError
		resp.Msg = "not found token"
		return &resp
	}

	tk, err := m.TokenManager.Get(token)
	if err != nil {
		resp.Code = code.GetError
		resp.Msg = err.Error()
		return &resp
	}

	u, err := m.UserManager.Get(tk.UserName)
	if err != nil {
		resp.Code = code.GetError
		resp.Msg = err.Error()
		return &resp
	}
	resp.Data = u
	//resp.Data = &types.User{}
	return &resp
}

func LocalMiddleware() gin.HandlerFunc {
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
		c.Next()
	}
}
