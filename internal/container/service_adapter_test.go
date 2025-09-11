package container

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultRetryOptions(t *testing.T) {
	opts := DefaultRetryOptions()

	assert.Equal(t, 5, opts.MaxAttempts)
	assert.Equal(t, 100*time.Millisecond, opts.InitialDelay)
	assert.Equal(t, 30*time.Second, opts.MaxDelay)
	assert.NotNil(t, opts.BackoffFunction)
}

func TestNewServiceAdapter(t *testing.T) {
	cfg := aws.Config{
		Region: "us-east-1",
	}

	adapter := NewServiceAdapter(cfg)

	assert.NotNil(t, adapter)
	assert.Equal(t, "us-east-1", adapter.config.Region)
	assert.Equal(t, 5, adapter.retryOptions.MaxAttempts)
}

func TestServiceAdapter_WithCustomRetryOptions(t *testing.T) {
	cfg := aws.Config{
		Region: "us-east-1",
	}

	customOpts := func(opts *RetryOptions) {
		opts.MaxAttempts = 10
		opts.MaxDelay = 60 * time.Second
	}

	adapter := NewServiceAdapter(cfg, customOpts)

	assert.NotNil(t, adapter)
	assert.Equal(t, 10, adapter.retryOptions.MaxAttempts)
	assert.Equal(t, 60*time.Second, adapter.retryOptions.MaxDelay)
}

func TestServiceAdapter_Clients(t *testing.T) {
	cfg := aws.Config{
		Region: "us-east-1",
	}

	adapter := NewServiceAdapter(cfg)

	t.Run("CloudWatchLogsClient", func(t *testing.T) {
		client := adapter.CloudWatchLogsClient()
		assert.NotNil(t, client)
	})

	t.Run("ConfigServiceClient", func(t *testing.T) {
		client := adapter.ConfigServiceClient()
		assert.NotNil(t, client)
	})

	t.Run("KMSClient", func(t *testing.T) {
		client := adapter.KMSClient()
		assert.NotNil(t, client)
	})
}

func TestExecuteWithRetry(t *testing.T) {
	cfg := aws.Config{
		Region: "us-east-1",
	}

	t.Run("Success on first attempt", func(t *testing.T) {
		adapter := NewServiceAdapter(cfg)
		ctx := context.Background()

		callCount := 0
		err := adapter.ExecuteWithRetry(ctx, func() error {
			callCount++
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)
	})

	t.Run("Success after retry", func(t *testing.T) {
		adapter := NewServiceAdapter(cfg, func(opts *RetryOptions) {
			opts.InitialDelay = 1 * time.Millisecond
			opts.MaxDelay = 10 * time.Millisecond
		})
		ctx := context.Background()

		callCount := 0
		err := adapter.ExecuteWithRetry(ctx, func() error {
			callCount++
			if callCount < 3 {
				return &smithy.GenericAPIError{
					Code: "ThrottlingException",
				}
			}
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 3, callCount)
	})

	t.Run("Non-retryable error", func(t *testing.T) {
		adapter := NewServiceAdapter(cfg)
		ctx := context.Background()

		callCount := 0
		expectedErr := errors.New("non-retryable error")
		err := adapter.ExecuteWithRetry(ctx, func() error {
			callCount++
			return expectedErr
		})

		assert.Error(t, err)
		assert.Equal(t, 1, callCount)
	})

	t.Run("Max attempts exceeded", func(t *testing.T) {
		adapter := NewServiceAdapter(cfg, func(opts *RetryOptions) {
			opts.MaxAttempts = 3
			opts.InitialDelay = 1 * time.Millisecond
			opts.MaxDelay = 10 * time.Millisecond
		})
		ctx := context.Background()

		callCount := 0
		err := adapter.ExecuteWithRetry(ctx, func() error {
			callCount++
			return &smithy.GenericAPIError{
				Code: "ThrottlingException",
			}
		})

		assert.Error(t, err)
		assert.Equal(t, 3, callCount)
		assert.Contains(t, err.Error(), "operation failed after 3 attempts")
	})

	t.Run("Context cancellation", func(t *testing.T) {
		adapter := NewServiceAdapter(cfg, func(opts *RetryOptions) {
			opts.InitialDelay = 100 * time.Millisecond
		})
		ctx, cancel := context.WithCancel(context.Background())

		callCount := 0
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		err := adapter.ExecuteWithRetry(ctx, func() error {
			callCount++
			return &smithy.GenericAPIError{
				Code: "ThrottlingException",
			}
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cancelled")
		assert.Equal(t, 1, callCount)
	})
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
		{
			name: "ThrottlingException",
			err: &smithy.GenericAPIError{
				Code: "ThrottlingException",
			},
			expected: true,
		},
		{
			name: "TooManyRequestsException",
			err: &smithy.GenericAPIError{
				Code: "TooManyRequestsException",
			},
			expected: true,
		},
		{
			name: "RequestLimitExceededException",
			err: &smithy.GenericAPIError{
				Code: "RequestLimitExceededException",
			},
			expected: true,
		},
		{
			name: "ServiceUnavailable",
			err: &smithy.GenericAPIError{
				Code: "ServiceUnavailable",
			},
			expected: true,
		},
		{
			name: "InternalServerError",
			err: &smithy.GenericAPIError{
				Code: "InternalServerError",
			},
			expected: true,
		},
		{
			name: "RequestTimeout",
			err: &smithy.GenericAPIError{
				Code: "RequestTimeout",
			},
			expected: true,
		},
		{
			name: "RequestTimeoutException",
			err: &smithy.GenericAPIError{
				Code: "RequestTimeoutException",
			},
			expected: true,
		},
		{
			name: "Non-retryable API error",
			err: &smithy.GenericAPIError{
				Code: "ValidationException",
			},
			expected: false,
		},
		{
			name:     "Generic error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsThrottlingError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
		{
			name: "ThrottlingException",
			err: &smithy.GenericAPIError{
				Code: "ThrottlingException",
			},
			expected: true,
		},
		{
			name: "TooManyRequestsException",
			err: &smithy.GenericAPIError{
				Code: "TooManyRequestsException",
			},
			expected: true,
		},
		{
			name: "RequestLimitExceededException",
			err: &smithy.GenericAPIError{
				Code: "RequestLimitExceededException",
			},
			expected: true,
		},
		{
			name: "Non-throttling error",
			err: &smithy.GenericAPIError{
				Code: "ValidationException",
			},
			expected: false,
		},
		{
			name:     "Generic error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isThrottlingError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExponentialBackoff(t *testing.T) {
	tests := []struct {
		name        string
		attempt     int
		err         error
		minExpected time.Duration
		maxExpected time.Duration
	}{
		{
			name:        "First attempt",
			attempt:     1,
			err:         errors.New("error"),
			minExpected: 75 * time.Millisecond,  // 100ms - 25%
			maxExpected: 125 * time.Millisecond, // 100ms + 25%
		},
		{
			name:        "Second attempt",
			attempt:     2,
			err:         errors.New("error"),
			minExpected: 150 * time.Millisecond, // 200ms - 25%
			maxExpected: 250 * time.Millisecond, // 200ms + 25%
		},
		{
			name:        "Third attempt",
			attempt:     3,
			err:         errors.New("error"),
			minExpected: 300 * time.Millisecond, // 400ms - 25%
			maxExpected: 500 * time.Millisecond, // 400ms + 25%
		},
		{
			name:    "Throttling error doubles delay",
			attempt: 1,
			err: &smithy.GenericAPIError{
				Code: "ThrottlingException",
			},
			minExpected: 150 * time.Millisecond, // (100ms * 2) - 25%
			maxExpected: 250 * time.Millisecond, // (100ms * 2) + 25%
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := exponentialBackoff(tt.attempt, tt.err)
			assert.GreaterOrEqual(t, delay, tt.minExpected)
			assert.LessOrEqual(t, delay, tt.maxExpected)
		})
	}
}

func TestRateLimiter(t *testing.T) {
	t.Run("Basic rate limiting", func(t *testing.T) {
		rl := NewRateLimiter(2) // 2 requests per second
		defer rl.Stop()

		ctx := context.Background()

		// Should allow first two requests immediately
		err1 := rl.Wait(ctx)
		err2 := rl.Wait(ctx)

		assert.NoError(t, err1)
		assert.NoError(t, err2)

		// Third request should wait
		start := time.Now()
		err3 := rl.Wait(ctx)
		elapsed := time.Since(start)

		assert.NoError(t, err3)
		assert.GreaterOrEqual(t, elapsed, 400*time.Millisecond)
	})

	t.Run("Context cancellation", func(t *testing.T) {
		rl := NewRateLimiter(1)
		defer rl.Stop()

		ctx, cancel := context.WithCancel(context.Background())

		// Use up the token
		err := rl.Wait(ctx)
		require.NoError(t, err)

		// Cancel context before next token is available
		cancel()

		err = rl.Wait(ctx)
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("Throttle behavior", func(t *testing.T) {
		rl := NewRateLimiter(10)
		defer rl.Stop()

		// Simulate throttling - first ThrottleThreshold calls should return 0 duration
		for i := 0; i < ThrottleThreshold; i++ {
			duration := rl.Throttle()
			assert.Equal(t, time.Duration(0), duration)
		}

		// Next throttle (after threshold) should return backoff duration
		duration := rl.Throttle()
		assert.Equal(t, ThrottleBackoffDuration, duration)
		assert.Equal(t, int32(ThrottleThreshold+1), rl.GetThrottleCount())

		// After many successes, throttle count should reset
		for i := 0; i < 11; i++ {
			rl.Success()
		}

		assert.Equal(t, int32(0), rl.GetThrottleCount())
		assert.Equal(t, int32(0), rl.GetSuccessCount()) // Should reset both counters
	})
}

func TestServiceMetrics(t *testing.T) {
	metrics := &ServiceMetrics{}

	t.Run("RecordSuccess", func(t *testing.T) {
		metrics.RecordSuccess()
		assert.Equal(t, int64(1), metrics.TotalCalls)
		assert.Equal(t, int64(1), metrics.SuccessfulCalls)
		assert.Equal(t, int64(0), metrics.FailedCalls)
	})

	t.Run("RecordFailure", func(t *testing.T) {
		metrics.RecordFailure()
		assert.Equal(t, int64(2), metrics.TotalCalls)
		assert.Equal(t, int64(1), metrics.SuccessfulCalls)
		assert.Equal(t, int64(1), metrics.FailedCalls)
	})

	t.Run("RecordRetry", func(t *testing.T) {
		metrics.RecordRetry()
		assert.Equal(t, int64(1), metrics.RetryCount)
	})

	t.Run("RecordThrottle", func(t *testing.T) {
		metrics.RecordThrottle()
		assert.Equal(t, int64(1), metrics.ThrottleCount)
	})
}

func TestExecuteWithRateLimit(t *testing.T) {
	cfg := aws.Config{
		Region: "us-east-1",
	}

	t.Run("Successful execution with rate limiting", func(t *testing.T) {
		adapter := NewServiceAdapter(cfg, func(opts *RetryOptions) {
			opts.InitialDelay = 1 * time.Millisecond
			opts.MaxDelay = 10 * time.Millisecond
		})

		rl := NewRateLimiter(10)
		defer rl.Stop()

		ctx := context.Background()
		callCount := 0

		err := adapter.ExecuteWithRateLimit(ctx, func() error {
			callCount++
			return nil
		}, rl)

		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)
		assert.Equal(t, int32(1), rl.GetSuccessCount())
	})

	t.Run("Throttling error updates rate limiter", func(t *testing.T) {
		adapter := NewServiceAdapter(cfg, func(opts *RetryOptions) {
			opts.MaxAttempts = 2
			opts.InitialDelay = 1 * time.Millisecond
			opts.MaxDelay = 10 * time.Millisecond
		})

		rl := NewRateLimiter(10)
		defer rl.Stop()

		ctx := context.Background()
		callCount := 0

		err := adapter.ExecuteWithRateLimit(ctx, func() error {
			callCount++
			if callCount == 1 {
				return &smithy.GenericAPIError{
					Code: "ThrottlingException",
				}
			}
			return nil
		}, rl)

		assert.NoError(t, err)
		assert.Equal(t, 2, callCount)
		assert.Equal(t, int32(1), rl.GetThrottleCount())
		assert.Equal(t, int32(1), rl.GetSuccessCount())
	})

	t.Run("Context cancellation during throttle backoff", func(t *testing.T) {
		adapter := NewServiceAdapter(cfg, func(opts *RetryOptions) {
			opts.MaxAttempts = 2
			opts.InitialDelay = 1 * time.Millisecond
		})

		rl := NewRateLimiter(10)
		defer rl.Stop()

		// Pre-fill throttle count to trigger backoff
		for i := 0; i < ThrottleThreshold; i++ {
			rl.Throttle()
		}

		ctx, cancel := context.WithCancel(context.Background())
		callCount := 0

		// Cancel context after a short delay
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		err := adapter.ExecuteWithRateLimit(ctx, func() error {
			callCount++
			// Return throttling error to trigger backoff
			return &smithy.GenericAPIError{
				Code: "ThrottlingException",
			}
		}, rl)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cancelled")
		assert.Equal(t, 1, callCount)
	})
}
