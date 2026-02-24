package health

import (
	"context"
	"time"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

// ComponentHealth represents the health of a single component
type ComponentHealth struct {
	Status  HealthStatus `json:"status"`
	Latency int64        `json:"latency_ms,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// HealthReport represents the overall health report
type HealthReport struct {
	Status      HealthStatus                `json:"status"`
	Timestamp   time.Time                   `json:"timestamp"`
	Components  map[string]*ComponentHealth `json:"components"`
	Environment string                      `json:"environment"`
	Version     string                      `json:"version"`
}

// HealthChecker defines the interface for health check implementations
type HealthChecker interface {
	Name() string
	Check(ctx context.Context) (int64, error)
}

// Service provides health checking functionality
type Service struct {
	checkers []HealthChecker
	env      string
	version  string
}

// NewHealthService creates a new health service
func NewHealthService(env, version string) *Service {
	return &Service{
		checkers: make([]HealthChecker, 0),
		env:      env,
		version:  version,
	}
}

// RegisterChecker registers a health checker for a component
func (s *Service) RegisterChecker(checker HealthChecker) {
	s.checkers = append(s.checkers, checker)
}

// Check performs health checks on all registered components
func (s *Service) Check(ctx context.Context) *HealthReport {
	report := &HealthReport{
		Status:      HealthStatusHealthy,
		Timestamp:   time.Now(),
		Components:  make(map[string]*ComponentHealth),
		Environment: s.env,
		Version:     s.version,
	}

	hasUnhealthy := false
	hasDegraded := false

	for _, checker := range s.checkers {
		checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		latency, err := checker.Check(checkCtx)
		cancel()

		component := &ComponentHealth{
			Status: HealthStatusHealthy,
		}

		if err != nil {
			component.Status = HealthStatusUnhealthy
			component.Error = err.Error()
			hasUnhealthy = true
		} else {
			component.Latency = latency
		}

		report.Components[checker.Name()] = component
	}

	// Determine overall status
	if hasUnhealthy {
		report.Status = HealthStatusUnhealthy
	} else if hasDegraded {
		report.Status = HealthStatusDegraded
	}

	return report
}

// MongoHealthChecker implements health check for MongoDB
type MongoHealthChecker struct {
	pingFunc func(ctx context.Context) error
}

// NewMongoHealthChecker creates a MongoDB health checker
func NewMongoHealthChecker(pingFunc func(ctx context.Context) error) *MongoHealthChecker {
	return &MongoHealthChecker{
		pingFunc: pingFunc,
	}
}

func (m *MongoHealthChecker) Name() string {
	return "mongodb"
}

func (m *MongoHealthChecker) Check(ctx context.Context) (int64, error) {
	start := time.Now()
	err := m.pingFunc(ctx)
	latency := time.Since(start).Milliseconds()
	return latency, err
}

// RedisHealthChecker implements health check for Redis
type RedisHealthChecker struct {
	pingFunc func(ctx context.Context) error
}

// NewRedisHealthChecker creates a Redis health checker
func NewRedisHealthChecker(pingFunc func(ctx context.Context) error) *RedisHealthChecker {
	return &RedisHealthChecker{
		pingFunc: pingFunc,
	}
}

func (r *RedisHealthChecker) Name() string {
	return "redis"
}

func (r *RedisHealthChecker) Check(ctx context.Context) (int64, error) {
	start := time.Now()
	err := r.pingFunc(ctx)
	latency := time.Since(start).Milliseconds()
	return latency, err
}

// PaymentProviderHealthChecker implements health check for payment providers
type PaymentProviderHealthChecker struct {
	name     string
	checkFunc func(ctx context.Context) error
}

// NewPaymentProviderHealthChecker creates a payment provider health checker
func NewPaymentProviderHealthChecker(name string, checkFunc func(ctx context.Context) error) *PaymentProviderHealthChecker {
	return &PaymentProviderHealthChecker{
		name:      name,
		checkFunc: checkFunc,
	}
}

func (p *PaymentProviderHealthChecker) Name() string {
	return "payment_" + p.name
}

func (p *PaymentProviderHealthChecker) Check(ctx context.Context) (int64, error) {
	start := time.Now()
	err := p.checkFunc(ctx)
	latency := time.Since(start).Milliseconds()
	return latency, err
}
