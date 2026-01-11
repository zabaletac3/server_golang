package httpx

import "github.com/gin-gonic/gin"

type AppHandler func(c *gin.Context) (any, error)
