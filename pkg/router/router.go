package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kubespace/kubespace/pkg/kube_resource"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/mysql"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/redis"
	"github.com/kubespace/kubespace/pkg/sse"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	views2 "github.com/kubespace/kubespace/pkg/views"
	"github.com/kubespace/kubespace/pkg/views/kube_views"
	"github.com/kubespace/kubespace/pkg/views/pipeline_views"
	ws_views2 "github.com/kubespace/kubespace/pkg/views/ws_views"
	"html/template"
	"io/ioutil"
	"k8s.io/klog"
	"net/http"
	"runtime"
	"strings"
)

type Router struct {
	*gin.Engine
}

func NewRouter(redisOptions *redis.Options, mysqlOptions *mysql.Options) (*Router, error) {
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

	kubeMessage := kube_resource.NewMiddleMessage(redisOptions)
	models, err := model.NewModels(redisOptions, mysqlOptions)
	if err != nil {
		return nil, err
	}
	kubeResources := kube_resource.NewKubeResources(kubeMessage)
	sse.Stream = sse.NewStream(redisOptions)

	// 统一认证的api接口
	apiGroup := engine.Group("/api/v1")
	viewsets := NewViewSets(kubeResources, models)
	for group, vs := range *viewsets {
		g := apiGroup.Group(group)
		for _, v := range vs {
			g.Handle(v.Method, v.Path, apiWrapper(models, v.Handler))
		}
	}

	pipelineCallbackView := pipeline_views.NewPipelineCallback(models)
	apiGroup.POST("/pipeline/callback", pipelineCallbackView.Callback)

	clusterAgent := views2.NewClusterAgent(models)
	engine.GET("/v1/import/:token", clusterAgent.AgentYaml)

	// 登录登出接口
	loginView := views2.NewLogin(models)
	apiGroup.POST("/login", loginView.Login)
	apiGroup.GET("/has_admin", loginView.HasAdmin)
	apiGroup.POST("/admin", loginView.CreateAdmin)
	apiGroup.POST("/logout", loginView.Logout)

	// 连接k8s agent的websocket接口
	kubeWs := ws_views2.NewKubeWs(redisOptions, models)
	apiGroup.GET("/kube/connect", kubeWs.Connect)

	// 连接k8s agent的websocket接口，用来并发传输返回数据
	kubeResp := ws_views2.NewKubeResp(redisOptions, models)
	apiGroup.GET("/kube/response", kubeResp.Connect)

	// 连接api websocket接口
	apiWs := ws_views2.NewApiWs(redisOptions, models, kubeResources)
	engine.GET("/ws/web/connect", apiWs.Connect)

	// 连接exec websocket接口
	execWs := ws_views2.NewExecWs(redisOptions, models, kubeResources)
	engine.GET("/ws/exec/:cluster/:namespace/:pod", execWs.Connect)

	// 连接log websocket接口
	logWs := ws_views2.NewLogWs(redisOptions, models, kubeResources)
	engine.GET("/ws/log/:cluster/:namespace/:pod", logWs.Connect)

	helmView := kube_views.NewHelm(kubeResources, models)
	engine.GET("/app/charts/*path", helmView.GetAppChart)

	return &Router{
		Engine: engine,
	}, nil
}

func apiWrapper(m *model.Models, handler views2.ViewHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		authRes := auth(m, c)
		if !authRes.IsSuccess() {
			c.JSON(401, authRes)
		} else {
			context := &views2.Context{Context: c, User: authRes.Data.(*types.User)}
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
