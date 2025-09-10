package container

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticationStrategy_Priority(t *testing.T) {
	strategies := []AuthStrategy{
		&ExplicitCredentialsStrategy{},
		&ProfileStrategy{},
		&AssumeRoleStrategy{},
		&TaskRoleStrategy{},
		&EnvironmentStrategy{},
		&InstanceProfileStrategy{},
		&DefaultStrategy{},
	}

	expectedPriorities := []int{1, 2, 3, 4, 5, 6, 99}

	for i, strategy := range strategies {
		assert.Equal(t, expectedPriorities[i], strategy.Priority(), "Strategy %s has wrong priority", strategy.Name())
	}
}

func TestExplicitCredentialsStrategy(t *testing.T) {
	strategy := &ExplicitCredentialsStrategy{}
	ctx := context.Background()

	tests := []struct {
		name        string
		envVars     map[string]string
		isAvailable bool
	}{
		{
			name: "with explicit credentials",
			envVars: map[string]string{
				"AWS_ACCESS_KEY_ID":     "test-key",
				"AWS_SECRET_ACCESS_KEY": "test-secret",
			},
			isAvailable: true,
		},
		{
			name: "missing access key",
			envVars: map[string]string{
				"AWS_SECRET_ACCESS_KEY": "test-secret",
			},
			isAvailable: false,
		},
		{
			name: "missing secret key",
			envVars: map[string]string{
				"AWS_ACCESS_KEY_ID": "test-key",
			},
			isAvailable: false,
		},
		{
			name:        "no credentials",
			envVars:     map[string]string{},
			isAvailable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore environment
			originalEnv := saveEnvironment([]string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"})
			defer restoreEnvironment(originalEnv)

			// Set test environment
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			options := AuthOptions{Region: "us-east-1"}
			assert.Equal(t, tt.isAvailable, strategy.IsAvailable(ctx, options))
		})
	}
}

func TestProfileStrategy(t *testing.T) {
	strategy := &ProfileStrategy{}
	ctx := context.Background()

	tests := []struct {
		name        string
		options     AuthOptions
		isAvailable bool
	}{
		{
			name: "with profile",
			options: AuthOptions{
				Profile: "test-profile",
				Region:  "us-east-1",
			},
			isAvailable: true,
		},
		{
			name: "without profile",
			options: AuthOptions{
				Region: "us-east-1",
			},
			isAvailable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isAvailable, strategy.IsAvailable(ctx, tt.options))
		})
	}
}

func TestAssumeRoleStrategy(t *testing.T) {
	strategy := &AssumeRoleStrategy{}
	ctx := context.Background()

	tests := []struct {
		name        string
		options     AuthOptions
		isAvailable bool
	}{
		{
			name: "with assume role",
			options: AuthOptions{
				AssumeRole: "arn:aws:iam::123456789012:role/test-role",
				Region:     "us-east-1",
			},
			isAvailable: true,
		},
		{
			name: "without assume role",
			options: AuthOptions{
				Region: "us-east-1",
			},
			isAvailable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isAvailable, strategy.IsAvailable(ctx, tt.options))
		})
	}
}

func TestTaskRoleStrategy(t *testing.T) {
	strategy := &TaskRoleStrategy{}
	ctx := context.Background()

	tests := []struct {
		name        string
		envVars     map[string]string
		isAvailable bool
	}{
		{
			name: "with ECS relative URI",
			envVars: map[string]string{
				"AWS_CONTAINER_CREDENTIALS_RELATIVE_URI": "/v2/credentials/task-id",
			},
			isAvailable: true,
		},
		{
			name: "with ECS full URI",
			envVars: map[string]string{
				"AWS_CONTAINER_CREDENTIALS_FULL_URI": "http://169.254.170.2/v2/credentials/task-id",
			},
			isAvailable: true,
		},
		{
			name:        "without ECS metadata",
			envVars:     map[string]string{},
			isAvailable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore environment
			originalEnv := saveEnvironment([]string{
				"AWS_CONTAINER_CREDENTIALS_RELATIVE_URI",
				"AWS_CONTAINER_CREDENTIALS_FULL_URI",
			})
			defer restoreEnvironment(originalEnv)

			// Clear environment
			os.Unsetenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI")
			os.Unsetenv("AWS_CONTAINER_CREDENTIALS_FULL_URI")

			// Set test environment
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			options := AuthOptions{Region: "us-east-1"}
			assert.Equal(t, tt.isAvailable, strategy.IsAvailable(ctx, options))
		})
	}
}

func TestEnvironmentStrategy(t *testing.T) {
	strategy := &EnvironmentStrategy{}
	ctx := context.Background()

	tests := []struct {
		name        string
		envVars     map[string]string
		options     AuthOptions
		isAvailable bool
	}{
		{
			name: "with AWS_PROFILE env var and no profile option",
			envVars: map[string]string{
				"AWS_PROFILE": "env-profile",
			},
			options:     AuthOptions{Region: "us-east-1"},
			isAvailable: true,
		},
		{
			name: "with AWS_PROFILE env var but profile option set",
			envVars: map[string]string{
				"AWS_PROFILE": "env-profile",
			},
			options: AuthOptions{
				Profile: "cli-profile",
				Region:  "us-east-1",
			},
			isAvailable: false,
		},
		{
			name:        "without AWS_PROFILE",
			envVars:     map[string]string{},
			options:     AuthOptions{Region: "us-east-1"},
			isAvailable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore environment
			originalEnv := saveEnvironment([]string{"AWS_PROFILE"})
			defer restoreEnvironment(originalEnv)

			// Clear and set test environment
			os.Unsetenv("AWS_PROFILE")
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			assert.Equal(t, tt.isAvailable, strategy.IsAvailable(ctx, tt.options))
		})
	}
}

func TestDefaultStrategy(t *testing.T) {
	strategy := &DefaultStrategy{}
	ctx := context.Background()

	// Default strategy should always be available
	options := AuthOptions{Region: "us-east-1"}
	assert.True(t, strategy.IsAvailable(ctx, options))
	assert.Equal(t, "default", strategy.Name())
	assert.Equal(t, 99, strategy.Priority())
}

func TestNewAuthenticationStrategy(t *testing.T) {
	authStrategy := NewAuthenticationStrategy()

	assert.NotNil(t, authStrategy)
	assert.Len(t, authStrategy.strategies, 7)

	// Verify strategies are in correct order
	expectedNames := []string{
		"explicit-credentials",
		"profile",
		"assume-role",
		"ecs-task-role",
		"environment",
		"ec2-instance-profile",
		"default",
	}

	for i, expectedName := range expectedNames {
		assert.Equal(t, expectedName, authStrategy.strategies[i].Name())
	}
}

// Helper functions for environment management
func saveEnvironment(keys []string) map[string]string {
	saved := make(map[string]string)
	for _, key := range keys {
		saved[key] = os.Getenv(key)
	}
	return saved
}

func restoreEnvironment(saved map[string]string) {
	for key, value := range saved {
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}
}
