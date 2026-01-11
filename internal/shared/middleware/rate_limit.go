package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/eren_dev/go_server/internal/config"
)

type ipLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rps      rate.Limit
	burst    int
}

func newIPLimiter(rps float64, burst int) *ipLimiter {
	return &ipLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rate.Limit(rps),
		burst:    burst,
	}
}

func (l *ipLimiter) getLimiter(ip string) *rate.Limiter {
	l.mu.RLock()
	limiter, exists := l.limiters[ip]
	l.mu.RUnlock()

	if exists {
		return limiter
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	limiter = rate.NewLimiter(l.rps, l.burst)
	l.limiters[ip] = limiter
	return limiter
}

func RateLimit(cfg *config.Config) gin.HandlerFunc {
	if !cfg.RateLimitEnabled {
		return func(c *gin.Context) { c.Next() }
	}

	limiter := newIPLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.getLimiter(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate_limit_exceeded",
			})
			return
		}
		c.Next()
	}
}
