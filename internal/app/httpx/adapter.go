package httpx

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/app/httperror"
	"github.com/eren_dev/go_server/internal/platform/logger"
)

func Adapt(handler AppHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		data, err := handler(c)

		if err != nil {
			status, payload := httperror.FromError(err)

			logger.Default().Error(ctx,
				"http_request_failed",
				"error", err,
				"status", status,
			)

			writeResponse(c, status, false, payload)
			return
		}

		writeResponse(c, http.StatusOK, true, data)
	}
}

func writeResponse(
	c *gin.Context,
	status int,
	success bool,
	data any,
) {
	requestID, _ := logger.RequestIDFromContext(c.Request.Context())

	resp := StandardResponse{
		Success:    success,
		Data:       data,
		StatusCode: status,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Path:        c.Request.URL.Path,
		RequestID:   requestID,
	}

	c.JSON(status, resp)
}
