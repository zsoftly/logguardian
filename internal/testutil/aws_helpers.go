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
func SetupTestEnvironment(t *testing.T) {
	t.Helper()
	// Clear any AWS-specific environment variables that might interfere
	// with test execution
	os.Unsetenv("AWS_REGION")
	os.Unsetenv("AWS_DEFAULT_REGION")
	os.Unsetenv("AWS_PROFILE")
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	os.Unsetenv("AWS_SESSION_TOKEN")
}

// CleanupTestEnvironment restores the test environment after test execution
// This function should be called in test cleanup (defer) if SetupTestEnvironment was used.
func CleanupTestEnvironment(t *testing.T) {
	t.Helper()
	// Environment cleanup is typically handled by test isolation
	// This function exists for symmetry and future extensibility
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
