package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Ready(c *gin.Context) {
	if !IsReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not_ready",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}
