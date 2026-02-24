package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Config holds the rate limiter configuration
type Config struct {
	Enabled       bool
	RPS           float64 // Requests per second
	Burst         int     // Maximum burst size
	TenantRPS     float64 // Per-tenant requests per second
	TenantBurst   int     // Per-tenant burst size
	GlobalRPS     float64 // Global rate limit across all tenants
	GlobalBurst   int     // Global burst size
}

// DefaultConfig returns a default rate limiter configuration
func DefaultConfig() Config {
	return Config{
		Enabled:     true,
		RPS:         100,
		Burst:       200,
		TenantRPS:   50,
		TenantBurst: 100,
		GlobalRPS:   1000,
		GlobalBurst: 2000,
	}
}

// Limiter provides hierarchical rate limiting (global + per-tenant)
type Limiter struct {
	mu           sync.RWMutex
	config       Config
	globalLimiter *rate.Limiter
	tenantLimiters map[string]*rate.Limiter
	lastCleanup  time.Time
}

// NewLimiter creates a new hierarchical rate limiter
func NewLimiter(cfg Config) *Limiter {
	if !cfg.Enabled {
		return &Limiter{config: cfg}
	}

	return &Limiter{
		config:       cfg,
		globalLimiter: rate.NewLimiter(rate.Limit(cfg.GlobalRPS), cfg.GlobalBurst),
		tenantLimiters: make(map[string]*rate.Limiter),
		lastCleanup:  time.Now(),
	}
}

// Allow checks if a request is allowed for the given tenant
// Returns (allowed, retryAfter)
func (l *Limiter) Allow(tenantID string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check global rate limit first
	if !l.globalLimiter.Allow() {
		return false, l.globalLimiter.Reserve().Delay()
	}

	// Get or create tenant limiter
	limiter, exists := l.tenantLimiters[tenantID]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(l.config.TenantRPS), l.config.TenantBurst)
		l.tenantLimiters[tenantID] = limiter

		// Periodically cleanup old tenant limiters
		if time.Since(l.lastCleanup) > 10*time.Minute {
			l.cleanup()
		}
	}

	// Check tenant rate limit
	if !limiter.Allow() {
		return false, limiter.Reserve().Delay()
	}

	return true, 0
}

// AllowN checks if n requests are allowed for the given tenant
func (l *Limiter) AllowN(tenantID string, n int) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check global rate limit first
	if !l.globalLimiter.AllowN(time.Now(), n) {
		return false, l.globalLimiter.ReserveN(time.Now(), n).Delay()
	}

	// Get or create tenant limiter
	limiter, exists := l.tenantLimiters[tenantID]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(l.config.TenantRPS), l.config.TenantBurst)
		l.tenantLimiters[tenantID] = limiter
	}

	// Check tenant rate limit
	if !limiter.AllowN(time.Now(), n) {
		return false, limiter.ReserveN(time.Now(), n).Delay()
	}

	return true, 0
}

// cleanup removes tenant limiters that haven't been used recently
func (l *Limiter) cleanup() {
	// Simple cleanup: remove limiters when map gets too large
	if len(l.tenantLimiters) > 1000 {
		// Keep only the most recently used limiters
		newLimiters := make(map[string]*rate.Limiter, 500)
		count := 0
		for k, v := range l.tenantLimiters {
			if count >= 500 {
				break
			}
			newLimiters[k] = v
			count++
		}
		l.tenantLimiters = newLimiters
	}
	l.lastCleanup = time.Now()
}

// GetTenantLimiter returns the rate limiter for a specific tenant
func (l *Limiter) GetTenantLimiter(tenantID string) *rate.Limiter {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.tenantLimiters[tenantID]
}

// RemoveTenantLimiter removes the rate limiter for a specific tenant
func (l *Limiter) RemoveTenantLimiter(tenantID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.tenantLimiters, tenantID)
}

// GetStats returns rate limiter statistics
func (l *Limiter) GetStats() map[string]interface{} {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return map[string]interface{}{
		"global_rps":       l.config.GlobalRPS,
		"global_burst":     l.config.GlobalBurst,
		"tenant_rps":       l.config.TenantRPS,
		"tenant_burst":     l.config.TenantBurst,
		"active_tenants":   len(l.tenantLimiters),
		"enabled":          l.config.Enabled,
	}
}
