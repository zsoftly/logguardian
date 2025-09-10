package main

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCommandLineArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		envVars  map[string]string
		expected CommandInput
	}{
		{
			name: "default values",
			args: []string{"cmd"},
			expected: CommandInput{
				Type:         "config-rule-evaluation",
				BatchSize:    10,
				OutputFormat: "json",
			},
		},
		{
			name: "with command line args",
			args: []string{"cmd", "--config-rule", "test-rule", "--region", "us-west-2", "--batch-size", "20", "--dry-run"},
			expected: CommandInput{
				Type:           "config-rule-evaluation",
				ConfigRuleName: "test-rule",
				Region:         "us-west-2",
				BatchSize:      20,
				DryRun:         true,
				OutputFormat:   "json",
			},
		},
		{
			name: "with environment variables",
			args: []string{"cmd"},
			envVars: map[string]string{
				"CONFIG_RULE_NAME": "env-rule",
				"AWS_REGION":       "eu-west-1",
				"BATCH_SIZE":       "30",
				"DRY_RUN":          "true",
			},
			expected: CommandInput{
				Type:           "config-rule-evaluation",
				ConfigRuleName: "env-rule",
				Region:         "eu-west-1",
				BatchSize:      30,
				DryRun:         true,
				OutputFormat:   "json",
			},
		},
		{
			name: "command line overrides environment",
			args: []string{"cmd", "--config-rule", "cli-rule", "--region", "ap-south-1"},
			envVars: map[string]string{
				"CONFIG_RULE_NAME": "env-rule",
				"AWS_REGION":       "eu-west-1",
			},
			expected: CommandInput{
				Type:           "config-rule-evaluation",
				ConfigRuleName: "cli-rule",
				Region:         "ap-south-1",
				BatchSize:      10,
				OutputFormat:   "json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			originalArgs := os.Args
			originalEnv := map[string]string{}

			// Set environment variables
			for key, value := range tt.envVars {
				originalEnv[key] = os.Getenv(key)
				os.Setenv(key, value)
			}

			// Set command line args
			os.Args = tt.args

			// Reset flag.CommandLine for testing
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			// Parse args
			result := parseCommandLineArgs()

			// Verify results
			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.ConfigRuleName, result.ConfigRuleName)
			assert.Equal(t, tt.expected.Region, result.Region)
			assert.Equal(t, tt.expected.BatchSize, result.BatchSize)
			assert.Equal(t, tt.expected.DryRun, result.DryRun)
			assert.Equal(t, tt.expected.OutputFormat, result.OutputFormat)

			// Restore original values
			os.Args = originalArgs
			for key, value := range originalEnv {
				if value == "" {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, value)
				}
			}
			for key := range tt.envVars {
				if _, exists := originalEnv[key]; !exists {
					os.Unsetenv(key)
				}
			}
		})
	}
}

func TestValidateInput(t *testing.T) {
	tests := []struct {
		name    string
		input   CommandInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid input",
			input: CommandInput{
				Type:           "config-rule-evaluation",
				ConfigRuleName: "test-rule",
				Region:         "us-east-1",
				BatchSize:      10,
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			input: CommandInput{
				Type:           "invalid-type",
				ConfigRuleName: "test-rule",
				Region:         "us-east-1",
				BatchSize:      10,
			},
			wantErr: true,
			errMsg:  "unsupported request type",
		},
		{
			name: "missing config rule name",
			input: CommandInput{
				Type:      "config-rule-evaluation",
				Region:    "us-east-1",
				BatchSize: 10,
			},
			wantErr: true,
			errMsg:  "config rule name is required",
		},
		{
			name: "missing region",
			input: CommandInput{
				Type:           "config-rule-evaluation",
				ConfigRuleName: "test-rule",
				BatchSize:      10,
			},
			wantErr: true,
			errMsg:  "region is required",
		},
		{
			name: "invalid batch size - zero",
			input: CommandInput{
				Type:           "config-rule-evaluation",
				ConfigRuleName: "test-rule",
				Region:         "us-east-1",
				BatchSize:      0,
			},
			wantErr: true,
			errMsg:  "batch size must be between 1 and 100",
		},
		{
			name: "invalid batch size - too large",
			input: CommandInput{
				Type:           "config-rule-evaluation",
				ConfigRuleName: "test-rule",
				Region:         "us-east-1",
				BatchSize:      101,
			},
			wantErr: true,
			errMsg:  "batch size must be between 1 and 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInput(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "with version env var",
			envValue: "1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "without version env var",
			envValue: "",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalValue := os.Getenv("APP_VERSION")
			defer func() {
				if originalValue == "" {
					os.Unsetenv("APP_VERSION")
				} else {
					os.Setenv("APP_VERSION", originalValue)
				}
			}()

			if tt.envValue != "" {
				os.Setenv("APP_VERSION", tt.envValue)
			} else {
				os.Unsetenv("APP_VERSION")
			}

			assert.Equal(t, tt.expected, getVersion())
		})
	}
}

func TestGetExecutionMode(t *testing.T) {
	tests := []struct {
		name     string
		dryRun   bool
		expected string
	}{
		{
			name:     "dry-run mode",
			dryRun:   true,
			expected: "dry-run",
		},
		{
			name:     "apply mode",
			dryRun:   false,
			expected: "apply",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, getExecutionMode(tt.dryRun))
		})
	}
}
