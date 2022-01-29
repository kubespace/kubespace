package core

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/router"
	"net/http"
)

type Server struct {
	config *ServerConfig
	router *router.Router
}

func NewServer(config *ServerConfig) (*Server, error) {
	r, err := router.NewRouter(config.RedisOptions, config.MysqlOptions)
	if err != nil {
		return nil, err
	}
	return &Server{
		config: config,
		router: r,
	}, nil
}

func (s *Server) Run() {
	insecureServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.InsecurePort),
		Handler: s.router,
	}
	go insecureServer.ListenAndServe()

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: s.router,
	}
	server.ListenAndServeTLS(s.config.CertFilePath, s.config.KeyFilePath)
}
