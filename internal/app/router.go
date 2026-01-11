package app

import (
	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/shared/errors"
	"github.com/eren_dev/go_server/internal/shared/httpx"
)

func registerRoutes(engine *gin.Engine) {
	r := httpx.NewRouter(engine)

	api := r.Group("/api")

	api.GET("/test-endpoint", func(c *gin.Context) (any, error) {
		return nil, errors.ErrNotFound
	})
}
