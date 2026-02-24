package circuitbreaker

import (
	"context"
	"time"

	"github.com/eren_dev/go_server/internal/platform/retry"
	"github.com/sony/gobreaker"
)

var (
	// PaymentBreaker is the circuit breaker for payment provider calls
	// Opens after 5 consecutive failures, closes after 60 seconds
	PaymentBreaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "payment",
		MaxRequests: 1,
		Interval:    30 * time.Second,
		Timeout:     60 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// State change logging can be added here
		},
	})

	// FCMBreaker is the circuit breaker for Firebase Cloud Messaging calls
	// Opens after 3 consecutive failures, closes after 30 seconds
	FCMBreaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "fcm",
		MaxRequests: 1,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	})

	// ExternalAPIBreaker is a generic circuit breaker for external API calls
	// Opens after 5 consecutive failures, closes after 60 seconds
	ExternalAPIBreaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "external_api",
		MaxRequests: 1,
		Interval:    30 * time.Second,
		Timeout:     60 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
	})
)

// ExecuteWithPaymentBreaker executes a function with the payment circuit breaker
func ExecuteWithPaymentBreaker(fn func() (interface{}, error)) (interface{}, error) {
	return PaymentBreaker.Execute(fn)
}

// ExecuteWithFCMBreaker executes a function with the FCM circuit breaker
func ExecuteWithFCMBreaker(fn func() (interface{}, error)) (interface{}, error) {
	return FCMBreaker.Execute(fn)
}

// ExecuteWithExternalAPIBreaker executes a function with the external API circuit breaker
func ExecuteWithExternalAPIBreaker(fn func() (interface{}, error)) (interface{}, error) {
	return ExternalAPIBreaker.Execute(fn)
}

// GetState returns the current state of a circuit breaker by name
func GetState(name string) gobreaker.State {
	switch name {
	case "payment":
		return PaymentBreaker.State()
	case "fcm":
		return FCMBreaker.State()
	case "external_api":
		return ExternalAPIBreaker.State()
	default:
		return gobreaker.StateClosed
	}
}

// StateToString converts a circuit breaker state to string
func StateToString(state gobreaker.State) string {
	switch state {
	case gobreaker.StateClosed:
		return "closed"
	case gobreaker.StateHalfOpen:
		return "half-open"
	case gobreaker.StateOpen:
		return "open"
	default:
		return "unknown"
	}
}

// ExecuteWithRetryAndBreaker executes a function with both retry logic and circuit breaker
// This provides resilience against transient failures while preventing cascade failures
func ExecuteWithRetryAndBreaker(ctx context.Context, breaker *gobreaker.CircuitBreaker, fn func() (interface{}, error), retryCfg *retry.Config) (interface{}, error) {
	var result interface{}
	var lastErr error

	// Wrap the function with retry logic
	retryFn := func() error {
		var err error
		result, err = fn()
		if err != nil {
			lastErr = err
		}
		return err
	}

	// Execute with circuit breaker
	_, cbErr := breaker.Execute(func() (interface{}, error) {
		// Execute with retry inside the circuit breaker
		retryErr := retry.WithExponentialBackoff(ctx, retryFn, retryCfg)
		return result, retryErr
	})

	if cbErr != nil {
		return nil, cbErr
	}

	return result, lastErr
}

// ExecuteWithPaymentBreakerAndRetry executes with payment circuit breaker and retry
func ExecuteWithPaymentBreakerAndRetry(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	return ExecuteWithRetryAndBreaker(ctx, PaymentBreaker, fn, nil)
}

// ExecuteWithFCMBreakerAndRetry executes with FCM circuit breaker and retry
func ExecuteWithFCMBreakerAndRetry(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	return ExecuteWithRetryAndBreaker(ctx, FCMBreaker, fn, nil)
}
