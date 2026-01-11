package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

func SlogRecovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		slog.Error("panic_recovered",
			"error", recovered,
			"path", c.Request.URL.Path,
		)
		c.AbortWithStatus(500)
	})
}
