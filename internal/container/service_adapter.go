package container

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/smithy-go"
)

const (
	// Throttling behavior constants
	ThrottleThreshold       = 3           // Number of throttles before backing off more aggressively
	ThrottleBackoffDuration = time.Second // Additional backoff duration when throttled

	// Jitter constants
	JitterPercentage = 0.25 // ±25% jitter range
)

// ServiceAdapter provides an abstraction layer for AWS service interactions
// with built-in retry logic, error handling, and circuit breaking capabilities
type ServiceAdapter struct {
	config       aws.Config
	retryOptions RetryOptions
}

// RetryOptions configures retry behavior for AWS service calls
type RetryOptions struct {
	MaxAttempts     int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFunction func(attempt int, err error) time.Duration
}

// DefaultRetryOptions provides sensible defaults for retry behavior
func DefaultRetryOptions() RetryOptions {
	return RetryOptions{
		MaxAttempts:     5,
		InitialDelay:    100 * time.Millisecond,
		MaxDelay:        30 * time.Second,
		BackoffFunction: exponentialBackoff,
	}
}

// NewServiceAdapter creates a new service adapter with the given configuration
func NewServiceAdapter(config aws.Config, opts ...func(*RetryOptions)) *ServiceAdapter {
	retryOpts := DefaultRetryOptions()
	for _, opt := range opts {
		opt(&retryOpts)
	}

	// Configure AWS SDK retry behavior
	config.Retryer = func() aws.Retryer {
		return retry.NewStandard(func(o *retry.StandardOptions) {
			o.MaxAttempts = retryOpts.MaxAttempts
			o.MaxBackoff = retryOpts.MaxDelay
		})
	}

	return &ServiceAdapter{
		config:       config,
		retryOptions: retryOpts,
	}
}

// CloudWatchLogsClient returns a CloudWatch Logs client with retry configuration
func (s *ServiceAdapter) CloudWatchLogsClient() *cloudwatchlogs.Client {
	return cloudwatchlogs.NewFromConfig(s.config)
}

// ConfigServiceClient returns a Config Service client with retry configuration
func (s *ServiceAdapter) ConfigServiceClient() *configservice.Client {
	return configservice.NewFromConfig(s.config)
}

// KMSClient returns a KMS client with retry configuration
func (s *ServiceAdapter) KMSClient() *kms.Client {
	return kms.NewFromConfig(s.config)
}

// ExecuteWithRetry performs an operation with retry logic and exponential backoff
func (s *ServiceAdapter) ExecuteWithRetry(ctx context.Context, operation func() error) error {
	var lastErr error

	for attempt := 1; attempt <= s.retryOptions.MaxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err) {
			slog.Error("Non-retryable error encountered",
				"attempt", attempt,
				"error", err)
			return err
		}

		// Don't retry on last attempt
		if attempt == s.retryOptions.MaxAttempts {
			break
		}

		// Calculate backoff delay
		delay := s.retryOptions.BackoffFunction(attempt, err)
		if delay > s.retryOptions.MaxDelay {
			delay = s.retryOptions.MaxDelay
		}

		slog.Warn("Operation failed, retrying",
			"attempt", attempt,
			"max_attempts", s.retryOptions.MaxAttempts,
			"delay", delay,
			"error", err)

		// Wait with context cancellation support
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", s.retryOptions.MaxAttempts, lastErr)
}

// ExecuteWithRateLimit performs an operation with rate limit handling
func (s *ServiceAdapter) ExecuteWithRateLimit(ctx context.Context, operation func() error, rateLimit *RateLimiter) error {
	return s.ExecuteWithRetry(ctx, func() error {
		// Wait for rate limit token
		if err := rateLimit.Wait(ctx); err != nil {
			return fmt.Errorf("rate limit wait failed: %w", err)
		}

		// Execute the operation
		err := operation()

		// Update rate limiter based on error
		if isThrottlingError(err) {
			// Get backoff duration from rate limiter
			if backoffDuration := rateLimit.Throttle(); backoffDuration > 0 {
				// Context-aware sleep
				select {
				case <-time.After(backoffDuration):
					// Backoff completed
				case <-ctx.Done():
					return fmt.Errorf("context cancelled during throttle backoff: %w", ctx.Err())
				}
			}
		} else if err == nil {
			rateLimit.Success()
		}

		return err
	})
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for AWS API errors
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "ThrottlingException",
			"TooManyRequestsException",
			"RequestLimitExceededException",
			"ServiceUnavailable",
			"InternalServerError",
			"RequestTimeout",
			"RequestTimeoutException":
			return true
		}
	}

	// Check for specific error types
	if isThrottlingError(err) {
		return true
	}

	return false
}

// isThrottlingError checks if an error is due to API throttling
func isThrottlingError(err error) bool {
	if err == nil {
		return false
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "ThrottlingException",
			"TooManyRequestsException",
			"RequestLimitExceededException":
			return true
		}
	}

	return false
}

// calculateJitter adds randomized jitter to a duration to prevent thundering herd
// The jitter is ±25% of the base duration, providing a random variation
// that helps distribute retry attempts across time.
func calculateJitter(baseDuration time.Duration) time.Duration {
	// Generate a random value between -0.25 and +0.25
	// rand.Float64() returns [0.0, 1.0), so (rand.Float64() - 0.5) gives [-0.5, 0.5)
	// Multiplying by 2 * JitterPercentage gives us [-0.25, 0.25)
	jitterMultiplier := (rand.Float64() - 0.5) * 2 * JitterPercentage

	// Apply jitter to the base duration
	jitterAmount := time.Duration(float64(baseDuration) * jitterMultiplier)
	return baseDuration + jitterAmount
}

// exponentialBackoff calculates exponential backoff with jitter
func exponentialBackoff(attempt int, err error) time.Duration {
	// Base delay with exponential increase
	baseDelay := 100 * time.Millisecond
	multiplier := math.Pow(2, float64(attempt-1))
	delay := time.Duration(float64(baseDelay) * multiplier)

	// Add jitter to prevent thundering herd
	delay = calculateJitter(delay)

	// Apply throttling-specific backoff if needed
	if isThrottlingError(err) {
		delay *= 2 // Double the delay for throttling errors
	}

	return delay
}

// RateLimiter provides rate limiting functionality with thread-safe operations
type RateLimiter struct {
	tokens        chan struct{}
	refillTicker  *time.Ticker
	throttleCount atomic.Int32
	successCount  atomic.Int32
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(ratePerSecond int) *RateLimiter {
	rl := &RateLimiter{
		tokens:       make(chan struct{}, ratePerSecond),
		refillTicker: time.NewTicker(time.Second / time.Duration(ratePerSecond)),
	}

	// Fill initial tokens
	for i := 0; i < ratePerSecond; i++ {
		rl.tokens <- struct{}{}
	}

	// Start refill goroutine
	go rl.refill()

	return rl
}

// Wait blocks until a rate limit token is available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Throttle indicates that a throttling error occurred
// Returns a duration to wait if backoff is needed, or 0 if no backoff required
func (rl *RateLimiter) Throttle() time.Duration {
	count := rl.throttleCount.Add(1)
	// Return backoff duration if threshold exceeded
	if count > ThrottleThreshold {
		return ThrottleBackoffDuration
	}
	return 0
}

// Success indicates a successful operation
func (rl *RateLimiter) Success() {
	successCount := rl.successCount.Add(1)
	// Reset throttle count after consecutive successes
	if successCount > 10 {
		rl.throttleCount.Store(0)
		rl.successCount.Store(0) // Reset success count to start fresh
	}
}

// refill adds tokens to the rate limiter
func (rl *RateLimiter) refill() {
	for range rl.refillTicker.C {
		select {
		case rl.tokens <- struct{}{}:
		default:
			// Bucket is full
		}
	}
}

// Stop cleanly stops the rate limiter
func (rl *RateLimiter) Stop() {
	rl.refillTicker.Stop()
}

// GetThrottleCount returns the current throttle count (thread-safe)
func (rl *RateLimiter) GetThrottleCount() int32 {
	return rl.throttleCount.Load()
}

// GetSuccessCount returns the current success count (thread-safe)
func (rl *RateLimiter) GetSuccessCount() int32 {
	return rl.successCount.Load()
}

// ServiceMetrics tracks service call metrics
type ServiceMetrics struct {
	TotalCalls      int64
	SuccessfulCalls int64
	FailedCalls     int64
	RetryCount      int64
	ThrottleCount   int64
}

// RecordSuccess records a successful service call
func (m *ServiceMetrics) RecordSuccess() {
	m.TotalCalls++
	m.SuccessfulCalls++
}

// RecordFailure records a failed service call
func (m *ServiceMetrics) RecordFailure() {
	m.TotalCalls++
	m.FailedCalls++
}

// RecordRetry records a retry attempt
func (m *ServiceMetrics) RecordRetry() {
	m.RetryCount++
}

// RecordThrottle records a throttling event
func (m *ServiceMetrics) RecordThrottle() {
	m.ThrottleCount++
}
