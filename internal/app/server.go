package app

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/config"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(cfg *config.Config) (*Server, error) {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	registerRoutes(router)

	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + cfg.Port,
			Handler:      router,
			ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeoutSecs) * time.Second,
			ReadTimeout:       time.Duration(cfg.ReadTimeoutSecs) * time.Second,
			WriteTimeout:      time.Duration(cfg.WriteTimeoutSecs) * time.Second,
			IdleTimeout:       time.Duration(cfg.IdleTimeoutSecs) * time.Second,
			MaxHeaderBytes:    cfg.MaxHeaderBytes,
		},
	}, nil
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

