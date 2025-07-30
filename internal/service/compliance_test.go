package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	kmstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	logguardiantypes "github.com/zsoftly/logguardian/internal/types"
)

func TestComplianceService_RemediateLogGroup(t *testing.T) {
	tests := []struct {
		name             string
		compliance       logguardiantypes.ComplianceResult
		dryRun           bool
		expectEncryption bool
		expectRetention  bool
		kmsError         error
		logsError        error
		expectedSuccess  bool
	}{
		{
			name: "apply both encryption and retention",
			compliance: logguardiantypes.ComplianceResult{
				LogGroupName:      "/aws/lambda/test",
				Region:            "us-east-1",
				MissingEncryption: true,
				MissingRetention:  true,
			},
			expectEncryption: true,
			expectRetention:  true,
			expectedSuccess:  true,
		},
		{
			name: "apply only encryption",
			compliance: logguardiantypes.ComplianceResult{
				LogGroupName:      "/aws/lambda/test",
				Region:            "us-east-1",
				MissingEncryption: true,
				MissingRetention:  false,
			},
			expectEncryption: true,
			expectRetention:  false,
			expectedSuccess:  true,
		},
		{
			name: "apply only retention",
			compliance: logguardiantypes.ComplianceResult{
				LogGroupName:      "/aws/lambda/test",
				Region:            "us-east-1",
				MissingEncryption: false,
				MissingRetention:  true,
			},
			expectEncryption: false,
			expectRetention:  true,
			expectedSuccess:  true,
		},
		{
			name: "no remediation needed",
			compliance: logguardiantypes.ComplianceResult{
				LogGroupName:      "/aws/lambda/test",
				Region:            "us-east-1",
				MissingEncryption: false,
				MissingRetention:  false,
			},
			expectEncryption: false,
			expectRetention:  false,
			expectedSuccess:  true,
		},
		{
			name: "dry run mode",
			compliance: logguardiantypes.ComplianceResult{
				LogGroupName:      "/aws/lambda/test",
				Region:            "us-east-1",
				MissingEncryption: true,
				MissingRetention:  true,
			},
			dryRun:           true,
			expectEncryption: true,
			expectRetention:  true,
			expectedSuccess:  true,
		},
		{
			name: "kms key validation failure should fail",
			compliance: logguardiantypes.ComplianceResult{
				LogGroupName:      "/aws/lambda/test",
				Region:            "us-east-1",
				MissingEncryption: true,
				MissingRetention:  false,
			},
			kmsError:        errors.New("NotFoundException: Key 'alias/test-key' does not exist"),
			expectedSuccess: false,
		},
		{
			name: "kms key disabled should fail",
			compliance: logguardiantypes.ComplianceResult{
				LogGroupName:      "/aws/lambda/test",
				Region:            "us-east-1",
				MissingEncryption: true,
				MissingRetention:  false,
			},
			expectEncryption: false,
			expectedSuccess:  false,
		},
		{
			name: "logs error should fail",
			compliance: logguardiantypes.ComplianceResult{
				LogGroupName:      "/aws/lambda/test",
				Region:            "us-east-1",
				MissingEncryption: false,
				MissingRetention:  true,
			},
			logsError:       errors.New("CloudWatch Logs access denied"),
			expectedSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock clients
			mockLogsClient := &MockCloudWatchLogsClient{
				AssociateKmsKeyError:    tt.kmsError,
				PutRetentionPolicyError: tt.logsError,
			}

			keyState := kmstypes.KeyStateEnabled
			if tt.name == "kms key disabled should fail" {
				keyState = kmstypes.KeyStateDisabled
			}

			mockKmsClient := &MockKMSClient{
				DescribeKeyError: tt.kmsError,
				KeyState:         keyState,
			}

			// Create service
			service := &ComplianceService{
				logsClient: mockLogsClient,
				kmsClient:  mockKmsClient,
				config: ServiceConfig{
					DefaultKMSKeyAlias:   "alias/test-key",
					DefaultRetentionDays: 365,
					DryRun:               tt.dryRun,
					MaxKMSRetries:        3,
					RetryBaseDelay:       100,
				},
			}

			// Execute
			result, err := service.RemediateLogGroup(context.Background(), tt.compliance)

			// Check error expectation
			if tt.expectedSuccess && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectedSuccess && err == nil {
				t.Error("Expected error but got none")
			}

			if result == nil {
				if tt.expectedSuccess {
					t.Error("Expected result but got nil")
				}
				return
			}

			// Check result fields
			if result.Success != tt.expectedSuccess {
				t.Errorf("Expected Success=%v, got %v", tt.expectedSuccess, result.Success)
			}

			if result.EncryptionApplied != tt.expectEncryption {
				t.Errorf("Expected EncryptionApplied=%v, got %v", tt.expectEncryption, result.EncryptionApplied)
			}

			if result.RetentionApplied != tt.expectRetention {
				t.Errorf("Expected RetentionApplied=%v, got %v", tt.expectRetention, result.RetentionApplied)
			}

			// Check that the right API calls were made (unless dry run)
			if !tt.dryRun && tt.expectedSuccess {
				if tt.expectEncryption && !mockLogsClient.AssociateKmsKeyCalled {
					t.Error("Expected AssociateKmsKey to be called")
				}
				if tt.expectRetention && !mockLogsClient.PutRetentionPolicyCalled {
					t.Error("Expected PutRetentionPolicy to be called")
				}
				if tt.expectEncryption && !mockKmsClient.DescribeKeyCalled {
					t.Error("Expected DescribeKey to be called")
				}
			}
		})
	}
}

// MockCloudWatchLogsClient implements the CloudWatch Logs client interface for testing
type MockCloudWatchLogsClient struct {
	AssociateKmsKeyCalled    bool
	AssociateKmsKeyError     error
	PutRetentionPolicyCalled bool
	PutRetentionPolicyError  error
}

func (m *MockCloudWatchLogsClient) AssociateKmsKey(ctx context.Context, params *cloudwatchlogs.AssociateKmsKeyInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.AssociateKmsKeyOutput, error) {
	m.AssociateKmsKeyCalled = true
	if m.AssociateKmsKeyError != nil {
		return nil, m.AssociateKmsKeyError
	}
	return &cloudwatchlogs.AssociateKmsKeyOutput{}, nil
}

func (m *MockCloudWatchLogsClient) PutRetentionPolicy(ctx context.Context, params *cloudwatchlogs.PutRetentionPolicyInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.PutRetentionPolicyOutput, error) {
	m.PutRetentionPolicyCalled = true
	if m.PutRetentionPolicyError != nil {
		return nil, m.PutRetentionPolicyError
	}
	return &cloudwatchlogs.PutRetentionPolicyOutput{}, nil
}

func (m *MockCloudWatchLogsClient) DescribeLogGroups(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
	return &cloudwatchlogs.DescribeLogGroupsOutput{
		LogGroups: []types.LogGroup{},
	}, nil
}

// MockKMSClient implements the KMS client interface for testing
type MockKMSClient struct {
	DescribeKeyCalled  bool
	DescribeKeyError   error
	KeyId              string
	KeyState           kmstypes.KeyState
	GetKeyPolicyCalled bool
	GetKeyPolicyError  error
	KeyPolicy          string
	ListGrantsCalled   bool
	ListGrantsError    error
}

func (m *MockKMSClient) DescribeKey(ctx context.Context, params *kms.DescribeKeyInput, optFns ...func(*kms.Options)) (*kms.DescribeKeyOutput, error) {
	m.DescribeKeyCalled = true
	if m.DescribeKeyError != nil {
		return nil, m.DescribeKeyError
	}

	keyId := "12345678-1234-1234-1234-123456789012"
	if m.KeyId != "" {
		keyId = m.KeyId
	}

	keyState := kmstypes.KeyStateEnabled
	if m.KeyState != "" {
		keyState = m.KeyState
	}

	return &kms.DescribeKeyOutput{
		KeyMetadata: &kmstypes.KeyMetadata{
			KeyId:    aws.String(keyId),
			Arn:      aws.String(fmt.Sprintf("arn:aws:kms:us-east-1:123456789012:key/%s", keyId)),
			KeyState: keyState,
		},
	}, nil
}

func (m *MockKMSClient) GetKeyPolicy(ctx context.Context, params *kms.GetKeyPolicyInput, optFns ...func(*kms.Options)) (*kms.GetKeyPolicyOutput, error) {
	m.GetKeyPolicyCalled = true
	if m.GetKeyPolicyError != nil {
		return nil, m.GetKeyPolicyError
	}

	policy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
					"Service": "logs.amazonaws.com"
				},
				"Action": [
					"kms:Encrypt",
					"kms:Decrypt",
					"kms:ReEncrypt*",
					"kms:GenerateDataKey*",
					"kms:DescribeKey"
				],
				"Resource": "*"
			}
		]
	}`
	if m.KeyPolicy != "" {
		policy = m.KeyPolicy
	}

	return &kms.GetKeyPolicyOutput{
		Policy: aws.String(policy),
	}, nil
}

func (m *MockKMSClient) ListGrants(ctx context.Context, params *kms.ListGrantsInput, optFns ...func(*kms.Options)) (*kms.ListGrantsOutput, error) {
	m.ListGrantsCalled = true
	if m.ListGrantsError != nil {
		return nil, m.ListGrantsError
	}

	return &kms.ListGrantsOutput{
		Grants: []kmstypes.GrantListEntry{},
	}, nil
}

func TestNewComplianceService(t *testing.T) {
	// Create a basic AWS config (this won't actually make AWS calls in tests)
	cfg := aws.Config{
		Region: "us-east-1",
	}

	service := NewComplianceService(cfg)

	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}

	if service.config.DefaultKMSKeyAlias == "" {
		t.Error("Expected default KMS key alias to be set")
	}

	if service.config.DefaultRetentionDays == 0 {
		t.Error("Expected default retention days to be set")
	}
}

func TestEnvironmentVariableHandling(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue string
		expected     string
	}{
		{
			name:         "return environment value when set",
			key:          "TEST_VAR",
			value:        "test_value",
			defaultValue: "default_value",
			expected:     "test_value",
		},
		{
			name:         "return default when env var not set",
			key:          "NONEXISTENT_VAR",
			value:        "",
			defaultValue: "default_value",
			expected:     "default_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if specified
			if tt.value != "" {
				t.Setenv(tt.key, tt.value)
			}

			result := getEnvOrDefault(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
