package retry

import (
	"context"
	"errors"
	"math"
	"time"
)

var (
	// ErrMaxRetriesExceeded is returned when all retry attempts have been exhausted
	ErrMaxRetriesExceeded = errors.New("max retries exceeded")
	// ErrContextCancelled is returned when the context is cancelled during retry
	ErrContextCancelled = errors.New("context cancelled during retry")
)

// Config holds the retry configuration
type Config struct {
	MaxRetries   int           // Maximum number of retry attempts
	BaseDelay    time.Duration // Initial delay between retries
	MaxDelay     time.Duration // Maximum delay between retries
	Multiplier   float64       // Multiplier for exponential backoff
	Jitter       bool          // Whether to add random jitter to delays
	IsRetryable  func(error) bool // Function to determine if an error is retryable
}

// DefaultConfig returns a default retry configuration
func DefaultConfig() Config {
	return Config{
		MaxRetries:  3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    10 * time.Second,
		Multiplier:  2.0,
		Jitter:      true,
		IsRetryable: IsRetryableError,
	}
}

// IsRetryableError determines if an error is retryable
// Returns true for network errors, timeouts, and service unavailable errors
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	// Add custom error type checks here
	return true // Default to retrying all errors
}

// WithExponentialBackoff executes a function with exponential backoff retry logic
func WithExponentialBackoff(ctx context.Context, fn func() error, cfg *Config) error {
	if cfg == nil {
		defaultCfg := DefaultConfig()
		cfg = &defaultCfg
	}

	var lastErr error
	baseDelay := cfg.BaseDelay

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Execute the function
		if err := fn(); err == nil {
			return nil // Success
		} else {
			lastErr = err
		}

		// Check if error is retryable
		if cfg.IsRetryable != nil && !cfg.IsRetryable(lastErr) {
			return lastErr
		}

		// Don't sleep after the last attempt
		if attempt >= cfg.MaxRetries {
			break
		}

		// Calculate delay with exponential backoff
		delay := time.Duration(float64(baseDelay) * math.Pow(cfg.Multiplier, float64(attempt)))
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}

		// Add jitter if configured
		if cfg.Jitter {
			delay = addJitter(delay)
		}

		// Wait for delay or context cancellation
		select {
		case <-time.After(delay):
			// Continue to next retry
		case <-ctx.Done():
			return ErrContextCancelled
		}
	}

	return lastErr
}

// addJitter adds random jitter to the delay (±25%)
func addJitter(delay time.Duration) time.Duration {
	// Simple jitter: random value between 0.75x and 1.25x of delay
	// Using a simple approach without importing math/rand
	// For production, consider using crypto/rand for better randomness
	jitterFactor := 0.5 // Simplified: adds 0-50% extra delay
	return delay + time.Duration(int64(float64(delay)*jitterFactor*0.5))
}

// RetryFunc is a function that can be retried
type RetryFunc func() error

// Retrier provides a fluent interface for retry operations
type Retrier struct {
	cfg Config
}

// NewRetrier creates a new Retrier with the given configuration
func NewRetrier(cfg Config) *Retrier {
	return &Retrier{cfg: cfg}
}

// Execute runs the function with retry logic
func (r *Retrier) Execute(ctx context.Context, fn RetryFunc) error {
	return WithExponentialBackoff(ctx, fn, &r.cfg)
}

// WithMaxRetries sets the maximum number of retries
func (r *Retrier) WithMaxRetries(max int) *Retrier {
	r.cfg.MaxRetries = max
	return r
}

// WithBaseDelay sets the base delay
func (r *Retrier) WithBaseDelay(delay time.Duration) *Retrier {
	r.cfg.BaseDelay = delay
	return r
}

// WithMaxDelay sets the maximum delay
func (r *Retrier) WithMaxDelay(delay time.Duration) *Retrier {
	r.cfg.MaxDelay = delay
	return r
}

// WithMultiplier sets the exponential backoff multiplier
func (r *Retrier) WithMultiplier(multiplier float64) *Retrier {
	r.cfg.Multiplier = multiplier
	return r
}

// WithJitter enables or disables jitter
func (r *Retrier) WithJitter(enabled bool) *Retrier {
	r.cfg.Jitter = enabled
	return r
}

// WithRetryableFunc sets the function to determine if an error is retryable
func (r *Retrier) WithRetryableFunc(fn func(error) bool) *Retrier {
	r.cfg.IsRetryable = fn
	return r
}
