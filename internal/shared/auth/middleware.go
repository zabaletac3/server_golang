package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/config"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	EmailKey    contextKey = "email"
	UserTypeKey contextKey = "user_type"
)

func JWTMiddleware(cfg *config.Config) gin.HandlerFunc {
	jwtService := NewJWTService(cfg)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "unauthorized",
				"status":  http.StatusUnauthorized,
			})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid authorization format",
			})
			return
		}

		claims, err := jwtService.ValidateToken(parts[1], AccessToken)
		if err != nil {
			status := http.StatusUnauthorized
			message := "invalid token"

			if err == ErrExpiredToken {
				message = "token expired"
			}

			c.AbortWithStatusJSON(status, gin.H{
				"success": false,
				"error":   message,
			})
			return
		}

		c.Set(string(UserIDKey), claims.UserID)
		c.Set(string(EmailKey), claims.Email)
		c.Set(string(UserTypeKey), string(claims.UserType))

		c.Next()
	}
}

func GetUserID(c *gin.Context) string {
	return c.GetString(string(UserIDKey))
}

func GetEmail(c *gin.Context) string {
	return c.GetString(string(EmailKey))
}

func GetUserType(c *gin.Context) string {
	return c.GetString(string(UserTypeKey))
}
