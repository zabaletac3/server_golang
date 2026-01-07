package app

import (
	"io"

	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/app/health"
	"github.com/eren_dev/go_server/internal/app/httpx"
	"github.com/eren_dev/go_server/internal/app/middleware"
	"github.com/eren_dev/go_server/internal/domain/errors"
)

func registerRoutes(engine *gin.Engine) {
	// Power off internal Gin logs
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard

	// Global middlewares (infra)
	engine.Use(middleware.RequestID())
	engine.Use(middleware.SlogLogger())
	engine.Use(middleware.SlogRecovery())

	// Create wrapped router (Interceptor enabled)
	r := httpx.NewRouter(engine)

	// -----------------------------
	// Health (infra, no interceptor)
	// -----------------------------
	engine.GET("/health", health.Health)
	engine.GET("/ready", health.Ready)

	// -----------------------------
	// API routes (business)
	// -----------------------------
	api := r.Group("/api")

	// Register modules routes here
	// users.RegisterRoutes(api)
	// orders.RegisterRoutes(api)
	// auth.RegisterRoutes(api)

	api.GET("/test-endpoint", httpx.AppHandler(func(c *gin.Context) (any, error) {

		var isTrue bool = true

		if(isTrue) {
			return nil, errors.ErrNotFound
		}

		return gin.H{
			"message": "Hello, World!",
		}, nil
	}))

}
