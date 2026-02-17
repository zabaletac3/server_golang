package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	tenantIDKey  = "tenant_id"
	tenantHeader = "X-Tenant-ID"
)

// TenantMiddleware extracts X-Tenant-ID from the request header,
// validates it as a valid ObjectID, and stores it in the Gin context.
// Must be applied after JWTMiddleware.
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader(tenantHeader)
		if raw == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "X-Tenant-ID header is required",
			})
			return
		}

		oid, err := primitive.ObjectIDFromHex(raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "invalid X-Tenant-ID",
			})
			return
		}

		c.Set(tenantIDKey, oid)
		c.Next()
	}
}

// GetTenantID retrieves the tenant ObjectID from the Gin context.
// Returns primitive.NilObjectID if not set.
func GetTenantID(c *gin.Context) primitive.ObjectID {
	val, exists := c.Get(tenantIDKey)
	if !exists {
		return primitive.NilObjectID
	}
	oid, ok := val.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID
	}
	return oid
}
