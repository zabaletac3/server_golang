package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	sharedAuth "github.com/eren_dev/go_server/internal/shared/auth"
)

// OwnerGuardMiddleware verifica que el JWT pertenece a un owner (mobile app).
// Debe usarse después de JWTMiddleware.
func OwnerGuardMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType := sharedAuth.GetUserType(c)
		if userType != string(sharedAuth.UserTypeOwner) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "access restricted to owners",
				"status":  http.StatusForbidden,
			})
			return
		}
		c.Next()
	}
}
