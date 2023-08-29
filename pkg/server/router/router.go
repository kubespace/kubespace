package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/core/errors"
	"github.com/kubespace/kubespace/pkg/server/api"
	apictx "github.com/kubespace/kubespace/pkg/server/api/api"
	"github.com/kubespace/kubespace/pkg/server/api/cluster/agent"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/utils"
	"html/template"
	"io"
	"k8s.io/klog/v2"
	"net/http"
	"runtime"
)

type Router struct {
	*gin.Engine
	conf *config.ServerConfig
	auth *apictx.Auth
}

func NewRouter(conf *config.ServerConfig) *Router {
	return &Router{
		Engine: gin.Default(),
		conf:   conf,
		auth:   apictx.NewAuth(conf),
	}
}

func (r *Router) Init() error {
	engine := r.Engine

	engine.Use(LocalMiddleware())

	indexHtml, _ := io.ReadAll(Assets.Files["/index.html"])
	t, _ := template.New("").New("/index.html").Parse(string(indexHtml))
	engine.SetHTMLTemplate(t)
	engine.StaticFS("/static", Assets)
	engine.StaticFile("/favicon.png", "./favicon.png")

	engine.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "/index.html", nil)
	})

	engine.GET("/ui/*path", func(c *gin.Context) {
		c.HTML(http.StatusOK, "/index.html", nil)
	})

	// 统一认证的api接口
	apiGroup := engine.Group("/api/v1")

	for group, apis := range api.Apis(r.conf) {
		g := apiGroup.Group(group)

		for _, a := range apis.Apis() {
			g.Handle(a.Method, a.Path, r.apiWrapper(a.Handler))
		}
	}

	agentImportHandler := agent.ImportHandler(r.conf)
	// 集群导入agent
	engine.GET("/import/agent/:token", r.apiWrapper(agentImportHandler))

	// 静态资源，包括spacelet二进制
	apiGroup.StaticFS("/assets", gin.Dir("./assets", true))
	return nil
}

// 对所有api进行封装，进行认证以及鉴权
func (r *Router) apiWrapper(handler apictx.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		context := &apictx.Context{
			Context: c,
			Models:  r.conf.Models,
		}
		needAuth, perm, err := handler.Auth(context)
		if err != nil {
			c.JSON(http.StatusForbidden, context.ResponseError(errors.New(code.AuthError, err, errors.Overlap)))
			return
		}
		if needAuth {
			// 获取认证用户
			user, err := r.auth.Authenticate(context)
			if err != nil {
				c.JSON(http.StatusUnauthorized, context.ResponseError(errors.New(code.AuthError, err, errors.Overlap)))
				return
			}
			context.User = user
		}
		if perm != nil {
			// 对用户进行权限鉴权
			ok, err := r.auth.Authorize(context, perm)
			if err != nil {
				c.JSON(http.StatusForbidden, context.ResponseError(errors.New(code.AuthError, err, errors.Overlap)))
				return
			}
			if !ok {
				c.JSON(http.StatusForbidden, context.ResponseError(errors.New(code.AuthError, "无操作权限", errors.Overlap)))
				return
			}
		}
		if res := handler.Handle(context); res != nil {
			c.JSON(200, res)
		}
	}
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
