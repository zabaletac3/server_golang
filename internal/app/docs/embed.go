package docs

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed swagger.json
var swaggerJSON []byte

func SwaggerJSONHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write(swaggerJSON)
	}
}
