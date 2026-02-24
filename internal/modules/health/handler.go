package health

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/shared/database"
)

var (
	db            *database.MongoDB
	healthService *Service
)

// SetDatabase sets the database instance for health checks
func SetDatabase(d *database.MongoDB) {
	db = d
}

// SetHealthService sets the health service instance
func SetHealthService(s *Service) {
	healthService = s
}

// Health godoc
// @Summary Health check
// @Description Check the health status of all components
// @Tags health
// @Produce json
// @Success 200 {object} HealthReport
// @Failure 503 {object} HealthReport
// @Router /health [get]
func Health(c *gin.Context) {
	// Use extended health service if available
	if healthService != nil {
		report := healthService.Check(c.Request.Context())
		status := http.StatusOK
		if report.Status == HealthStatusUnhealthy {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, report)
		return
	}

	// Fallback to simple health check
	status := gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if db != nil {
		dbStatus := "healthy"
		if err := db.Health(c.Request.Context()); err != nil {
			dbStatus = "unhealthy"
			status["database_error"] = err.Error()
		}
		status["database"] = dbStatus
	}

	c.JSON(http.StatusOK, status)
}

// Ready godoc
// @Summary Readiness check
// @Description Check if the service is ready to accept traffic
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /ready [get]
func Ready(c *gin.Context) {
	if !IsReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "not_ready",
			"message": "Service is not ready to accept traffic",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"message": "Service is ready to accept traffic",
	})
}
