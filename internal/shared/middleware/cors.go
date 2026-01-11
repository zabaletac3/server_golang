package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/config"
)

func CORS(cfg *config.Config) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowOrigins,
		AllowMethods:     cfg.CORSAllowMethods,
		AllowHeaders:     cfg.CORSAllowHeaders,
		ExposeHeaders:    cfg.CORSExposeHeaders,
		AllowCredentials: cfg.CORSAllowCredentials,
		MaxAge:           time.Duration(cfg.CORSMaxAge) * time.Second,
	})
}
