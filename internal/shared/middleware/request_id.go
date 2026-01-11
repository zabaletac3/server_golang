package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/eren_dev/go_server/internal/platform/logger"
)

const headerRequestID = "X-Request-ID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(headerRequestID)
		if rid == "" {
			rid = uuid.NewString()
		}

		ctx := logger.WithRequestID(c.Request.Context(), rid)
		c.Request = c.Request.WithContext(ctx)

		c.Header(headerRequestID, rid)

		c.Next()
	}
}
