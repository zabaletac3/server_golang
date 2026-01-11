package health

import "github.com/gin-gonic/gin"

func RegisterRoutes(engine *gin.Engine) {
	engine.GET("/health", Health)
	engine.GET("/ready", Ready)
}
