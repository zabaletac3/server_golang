package httpx

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/platform/logger"
)

func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID, _ := logger.RequestIDFromContext(c.Request.Context())

		c.JSON(http.StatusNotFound, StandardResponse{
			Success:    false,
			StatusCode: http.StatusNotFound,
			Data: ErrorResponse{
				Code:    "NOT_FOUND",
				Message: "endpoint not found",
			},
			Path:      c.Request.URL.Path,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			RequestID: requestID,
		})
	}
}

func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID, _ := logger.RequestIDFromContext(c.Request.Context())

		c.JSON(http.StatusMethodNotAllowed, StandardResponse{
			Success:    false,
			StatusCode: http.StatusMethodNotAllowed,
			Data: ErrorResponse{
				Code:    "METHOD_NOT_ALLOWED",
				Message: "method not allowed",
			},
			Path:      c.Request.URL.Path,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			RequestID: requestID,
		})
	}
}
