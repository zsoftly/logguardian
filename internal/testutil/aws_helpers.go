package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// SetupTestEnvironment configures the test environment for isolated testing
// This function should be called at the start of test functions that need
// a clean environment state.
//
// Using t.Setenv (Go 1.17+) automatically saves and restores environment
// variables after the test completes, ensuring proper cleanup and isolation.
func SetupTestEnvironment(t *testing.T) {
	t.Helper()
	// Clear any AWS-specific environment variables that might interfere
	// with test execution. t.Setenv automatically restores original values
	// after test completion, preventing side effects in parallel tests.
	t.Setenv("AWS_REGION", "")
	t.Setenv("AWS_DEFAULT_REGION", "")
	t.Setenv("AWS_PROFILE", "")
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "")
	t.Setenv("AWS_SESSION_TOKEN", "")
}

// CleanupTestEnvironment is deprecated - cleanup is now automatic via t.Setenv
// This function is kept for backward compatibility but does nothing.
//
// Deprecated: No longer needed as SetupTestEnvironment uses t.Setenv which
// automatically restores environment variables after test completion.
func CleanupTestEnvironment(t *testing.T) {
	t.Helper()
	// Cleanup is handled automatically by t.Setenv in SetupTestEnvironment
}

// NewMockAWSConfig creates a mock AWS config for testing
// This returns a config that won't work with real AWS but allows
// tests to verify configuration loading without credentials.
func NewMockAWSConfig() (aws.Config, error) {
	// Load default config but don't require credentials
	// This will fail on actual AWS calls but allows config validation
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("ca-central-1"),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
	)
	if err != nil {
		return aws.Config{}, err
	}
	return cfg, nil
}

// IsLocalStackAvailable checks if LocalStack is available for integration testing
func IsLocalStackAvailable() bool {
	endpoint := os.Getenv("LOCALSTACK_ENDPOINT")
	return endpoint != ""
}

// GetLocalStackEndpoint returns the LocalStack endpoint if available
func GetLocalStackEndpoint() string {
	return os.Getenv("LOCALSTACK_ENDPOINT")
}
