package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/config"
)

func BodyLimit(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, cfg.MaxBodySize)
		c.Next()
	}
}

func Timeout(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), cfg.RequestTimeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		finished := make(chan struct{}, 1)
		go func() {
			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
			return
		case <-ctx.Done():
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
				"error": "request_timeout",
			})
		}
	}
}
