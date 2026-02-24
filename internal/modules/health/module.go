package health

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers health check routes
func RegisterRoutes(engine *gin.Engine) {
	engine.GET("/health", Health)
	engine.GET("/ready", Ready)
}

// RegisterRoutesGroup registers health check routes under a group
func RegisterRoutesGroup(group *gin.RouterGroup) {
	group.GET("/health", Health)
	group.GET("/ready", Ready)
}
