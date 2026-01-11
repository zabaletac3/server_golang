package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/platform/logger"
)

func SlogLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		logger.Default().Info(c.Request.Context(), "http_request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", duration.Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}
