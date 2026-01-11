package middleware

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/config"
)

func Compression(cfg *config.Config) gin.HandlerFunc {
	if !cfg.CompressionEnabled {
		return func(c *gin.Context) { c.Next() }
	}

	return gzip.Gzip(cfg.CompressionLevel)
}
