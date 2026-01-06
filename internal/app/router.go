package app

import (
	"io"

	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/app/health"
	"github.com/eren_dev/go_server/internal/app/middleware"
)

func registerRoutes(router *gin.Engine) {
	// Power off internal Gin logs
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard

	// Custom middlewares
	router.Use(middleware.RequestID())
	router.Use(middleware.SlogLogger())
	router.Use(middleware.SlogRecovery())

	router.GET("/health", health.Health)
	router.GET("/ready", health.Ready)


	// api := router.Group("/api")
	// {
	// }
}
