package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/platform/ratelimit"
)

// TenantRateLimitMiddleware creates a hierarchical rate limiter middleware
// that applies rate limits per tenant and globally
func TenantRateLimitMiddleware(limiter *ratelimit.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get tenant ID from context (set by TenantMiddleware)
		tenantID := GetTenantID(c)
		if tenantID.IsZero() {
			// No tenant ID, skip tenant rate limiting
			c.Next()
			return
		}

		allowed, retryAfter := limiter.Allow(tenantID.Hex())
		if !allowed {
			c.Header("Retry-After", retryAfter.String())
			c.Header("X-RateLimit-Limit", "50")
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", time.Now().Add(retryAfter).Format(time.RFC3339))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "rate limit exceeded",
				"message": "Too many requests from this tenant. Please try again later.",
				"retry_after": retryAfter.String(),
			})
			return
		}

		c.Next()
	}
}

// IPRateLimitMiddleware creates a rate limiter middleware based on IP address
func IPRateLimitMiddleware(limiter *ratelimit.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP
		ip := c.ClientIP()
		if ip == "" {
			c.Next()
			return
		}

		allowed, retryAfter := limiter.Allow(ip)
		if !allowed {
			c.Header("Retry-After", retryAfter.String())
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "rate limit exceeded",
				"message": "Too many requests from your IP. Please try again later.",
				"retry_after": retryAfter.String(),
			})
			return
		}

		c.Next()
	}
}
