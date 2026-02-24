package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/eren_dev/go_server/internal/platform/metrics"
)

// MetricsMiddleware creates a middleware that collects HTTP metrics
func MetricsMiddleware(m *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Increment in-flight requests
		m.HTTPRequestsInFlight.Inc()
		defer m.HTTPRequestsInFlight.Dec()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get path (remove parameters for cardinality)
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Record metrics
		m.HTTPRequestsTotal.WithLabelValues(
			c.Request.Method,
			path,
			strconv.Itoa(c.Writer.Status()),
		).Inc()

		m.HTTPRequestDuration.WithLabelValues(
			c.Request.Method,
			path,
		).Observe(duration)
	}
}

// PrometheusHandler returns the Prometheus metrics HTTP handler
func PrometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// RegisterMetricsRoutes registers the Prometheus metrics endpoint
func RegisterMetricsRoutes(engine *gin.Engine, m *metrics.Metrics) {
	// Use the middleware for all routes
	engine.Use(MetricsMiddleware(m))

	// Expose metrics endpoint
	metricsGroup := engine.Group("/metrics")
	metricsGroup.GET("", PrometheusHandler())
}
