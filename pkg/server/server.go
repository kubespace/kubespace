package server

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/server/router"
	"net/http"
)

type Server struct {
	config *config.ServerConfig
	router *router.Router
}

func NewServer(config *config.ServerConfig) (*Server, error) {
	r := router.NewRouter(config)
	if err := r.Init(); err != nil {
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
