package httpx

import (
	"github.com/gin-gonic/gin"
)

// AppHandler is a standard handler function type that returns data and error
type AppHandler func(c *gin.Context) (any, error)

// WriteError writes a standardized error response with request ID
func WriteError(c *gin.Context, status int, code, message string) {
	requestID, _ := c.Get("request_id")
	if requestID == nil {
		requestID = ""
	}

	response := NewErrorResponse(
		requestID.(string),
		c.Request.URL.Path,
		code,
		message,
	)

	c.JSON(status, response)
}

// WriteErrorFromErr converts an error to HTTP status and writes error response
func WriteErrorFromErr(c *gin.Context, err error) {
	if err == nil {
		return
	}

	requestID, _ := c.Get("request_id")
	if requestID == nil {
		requestID = ""
	}

	status, errResp := FromError(err)
	errResp.RequestID = requestID.(string)
	errResp.Path = c.Request.URL.Path

	c.JSON(status, errResp)
}

// WriteSuccess writes a standardized success response
func WriteSuccess(c *gin.Context, status int, data any) {
	requestID, _ := c.Get("request_id")
	if requestID == nil {
		requestID = ""
	}

	response := StandardResponse{
		Success:    true,
		Data:       data,
		StatusCode: status,
		Timestamp:  "", // Will be set by adapter
		Path:       c.Request.URL.Path,
		RequestID:  requestID.(string),
	}

	c.JSON(status, response)
}
