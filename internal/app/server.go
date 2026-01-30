package app

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/app/docs"
	"github.com/eren_dev/go_server/internal/config"
	"github.com/eren_dev/go_server/internal/modules/health"
	"github.com/eren_dev/go_server/internal/platform/payment"
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
	"github.com/eren_dev/go_server/internal/shared/middleware"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(cfg *config.Config, db *database.MongoDB, paymentManager *payment.PaymentManager) (*Server, error) {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	if len(cfg.TrustedProxies) > 0 {
		router.SetTrustedProxies(cfg.TrustedProxies)
	}

	router.Use(middleware.SlogRecovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.SecurityHeaders(cfg))
	router.Use(middleware.CORS(cfg))
	router.Use(middleware.RateLimit(cfg))
	router.Use(middleware.BodyLimit(cfg))
	router.Use(middleware.Compression(cfg))
	router.Use(middleware.SlogLogger())

	router.NoRoute(httpx.NotFoundHandler())
	router.NoMethod(httpx.MethodNotAllowedHandler())

	// Documentation
	router.GET("/docs", docs.ScalarHandler())
	router.GET("/docs/openapi.json", docs.SwaggerJSONHandler())

	health.RegisterRoutes(router)
	registerRoutes(router, db, cfg, paymentManager)

	return &Server{
		httpServer: &http.Server{
			Addr:              ":" + cfg.Port,
			Handler:           router,
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
