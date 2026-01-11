package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/config"
)

func SecurityHeaders(cfg *config.Config) gin.HandlerFunc {
	if !cfg.SecurityHeadersEnabled {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		if cfg.XFrameOptions != "" {
			c.Header("X-Frame-Options", cfg.XFrameOptions)
		}

		if cfg.ContentTypeNosniff {
			c.Header("X-Content-Type-Options", "nosniff")
		}

		if cfg.XSSProtection {
			c.Header("X-XSS-Protection", "1; mode=block")
		}

		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}
