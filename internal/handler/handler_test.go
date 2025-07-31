package handler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/zsoftly/logguardian/internal/types"
)

func TestComplianceHandler_HandleConfigEvent(t *testing.T) {
	tests := []struct {
		name        string
		event       types.ConfigEvent
		expectError bool
		expectCall  bool
	}{
		{
			name: "valid non-compliant log group",
			event: types.ConfigEvent{
				ConfigRuleName: "cloudwatch-log-group-encrypted",
				AccountId:      "123456789012",
				ConfigRuleInvokingEvent: types.ConfigRuleInvokingEvent{
					ConfigurationItem: types.ConfigurationItem{
						ResourceType:                 "AWS::Logs::LogGroup",
						ResourceName:                 "/aws/lambda/test-function",
						AwsRegion:                    "ca-central-1",
						AwsAccountId:                 "123456789012",
						ConfigurationItemStatus:      "ResourceDiscovered",
						ConfigurationItemCaptureTime: time.Now(),
						Configuration: types.LogGroupConfiguration{
							LogGroupName:    "/aws/lambda/test-function",
							RetentionInDays: nil, // Missing retention
							KmsKeyId:        "",  // Missing encryption
						},
					},
				},
			},
			expectError: false,
			expectCall:  true,
		},
		{
			name: "deleted resource should be skipped",
			event: types.ConfigEvent{
				ConfigRuleName: "cloudwatch-log-group-encrypted",
				AccountId:      "123456789012",
				ConfigRuleInvokingEvent: types.ConfigRuleInvokingEvent{
					ConfigurationItem: types.ConfigurationItem{
						ResourceType:            "AWS::Logs::LogGroup",
						ResourceName:            "/aws/lambda/test-function",
						AwsRegion:               "ca-central-1",
						ConfigurationItemStatus: "ResourceDeleted",
					},
				},
			},
			expectError: false,
			expectCall:  false,
		},
		{
			name: "non-log-group resource should be skipped",
			event: types.ConfigEvent{
				ConfigRuleName: "some-other-rule",
				AccountId:      "123456789012",
				ConfigRuleInvokingEvent: types.ConfigRuleInvokingEvent{
					ConfigurationItem: types.ConfigurationItem{
						ResourceType:            "AWS::S3::Bucket",
						ResourceName:            "my-bucket",
						AwsRegion:               "ca-central-1",
						ConfigurationItemStatus: "ResourceDiscovered",
					},
				},
			},
			expectError: false,
			expectCall:  false,
		},
		{
			name: "compliant log group should not trigger remediation",
			event: types.ConfigEvent{
				ConfigRuleName: "cloudwatch-log-group-encrypted",
				AccountId:      "123456789012",
				ConfigRuleInvokingEvent: types.ConfigRuleInvokingEvent{
					ConfigurationItem: types.ConfigurationItem{
						ResourceType:                 "AWS::Logs::LogGroup",
						ResourceName:                 "/aws/lambda/compliant-function",
						AwsRegion:                    "ca-central-1",
						AwsAccountId:                 "123456789012",
						ConfigurationItemStatus:      "ResourceDiscovered",
						ConfigurationItemCaptureTime: time.Now(),
						Configuration: types.LogGroupConfiguration{
							LogGroupName:    "/aws/lambda/compliant-function",
							RetentionInDays: intPtr(365),
							KmsKeyId:        "arn:aws:kms:ca-central-1:123456789012:key/12345678-1234-1234-1234-123456789012",
						},
					},
				},
			},
			expectError: false,
			expectCall:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := &MockComplianceService{}
			handler := NewComplianceHandler(mockService)

			// Marshal event to JSON
			eventBytes, err := json.Marshal(tt.event)
			if err != nil {
				t.Fatalf("Failed to marshal event: %v", err)
			}

			// Execute handler
			err = handler.HandleConfigEvent(context.Background(), eventBytes)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check if remediation was called
			if tt.expectCall && !mockService.RemediateLogGroupCalled {
				t.Error("Expected RemediateLogGroup to be called but it wasn't")
			}
			if !tt.expectCall && mockService.RemediateLogGroupCalled {
				t.Error("Expected RemediateLogGroup not to be called but it was")
			}
		})
	}
}

func TestComplianceHandler_analyzeCompliance(t *testing.T) {
	handler := &ComplianceHandler{}

	tests := []struct {
		name                      string
		configItem                types.ConfigurationItem
		expectedMissingEncryption bool
		expectedMissingRetention  bool
	}{
		{
			name: "missing both encryption and retention",
			configItem: types.ConfigurationItem{
				AwsRegion:    "ca-central-1",
				AwsAccountId: "123456789012",
				Configuration: types.LogGroupConfiguration{
					LogGroupName:    "/aws/lambda/test",
					RetentionInDays: nil,
					KmsKeyId:        "",
				},
			},
			expectedMissingEncryption: true,
			expectedMissingRetention:  true,
		},
		{
			name: "missing only encryption",
			configItem: types.ConfigurationItem{
				AwsRegion:    "ca-central-1",
				AwsAccountId: "123456789012",
				Configuration: types.LogGroupConfiguration{
					LogGroupName:    "/aws/lambda/test",
					RetentionInDays: intPtr(365),
					KmsKeyId:        "",
				},
			},
			expectedMissingEncryption: true,
			expectedMissingRetention:  false,
		},
		{
			name: "missing only retention",
			configItem: types.ConfigurationItem{
				AwsRegion:    "ca-central-1",
				AwsAccountId: "123456789012",
				Configuration: types.LogGroupConfiguration{
					LogGroupName:    "/aws/lambda/test",
					RetentionInDays: nil,
					KmsKeyId:        "arn:aws:kms:ca-central-1:123456789012:key/12345678-1234-1234-1234-123456789012",
				},
			},
			expectedMissingEncryption: false,
			expectedMissingRetention:  true,
		},
		{
			name: "fully compliant",
			configItem: types.ConfigurationItem{
				AwsRegion:    "ca-central-1",
				AwsAccountId: "123456789012",
				Configuration: types.LogGroupConfiguration{
					LogGroupName:    "/aws/lambda/test",
					RetentionInDays: intPtr(365),
					KmsKeyId:        "arn:aws:kms:ca-central-1:123456789012:key/12345678-1234-1234-1234-123456789012",
				},
			},
			expectedMissingEncryption: false,
			expectedMissingRetention:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.analyzeCompliance(tt.configItem)

			if result.MissingEncryption != tt.expectedMissingEncryption {
				t.Errorf("Expected MissingEncryption=%v, got %v", tt.expectedMissingEncryption, result.MissingEncryption)
			}

			if result.MissingRetention != tt.expectedMissingRetention {
				t.Errorf("Expected MissingRetention=%v, got %v", tt.expectedMissingRetention, result.MissingRetention)
			}

			if result.LogGroupName != tt.configItem.Configuration.LogGroupName {
				t.Errorf("Expected LogGroupName=%s, got %s", tt.configItem.Configuration.LogGroupName, result.LogGroupName)
			}
		})
	}
}

// MockComplianceService is a mock implementation for testing
type MockComplianceService struct {
	RemediateLogGroupCalled bool
	RemediateLogGroupError  error
	RemediateLogGroupResult *types.RemediationResult
}

func (m *MockComplianceService) RemediateLogGroup(ctx context.Context, compliance types.ComplianceResult) (*types.RemediationResult, error) {
	m.RemediateLogGroupCalled = true

	if m.RemediateLogGroupError != nil {
		return nil, m.RemediateLogGroupError
	}

	if m.RemediateLogGroupResult != nil {
		return m.RemediateLogGroupResult, nil
	}

	// Default success result
	return &types.RemediationResult{
		LogGroupName:      compliance.LogGroupName,
		Region:            compliance.Region,
		EncryptionApplied: compliance.MissingEncryption,
		RetentionApplied:  compliance.MissingRetention,
		Success:           true,
	}, nil
}

func (m *MockComplianceService) ProcessNonCompliantResources(ctx context.Context, request types.BatchComplianceRequest) (*types.BatchRemediationResult, error) {
	// Mock implementation for batch processing
	return &types.BatchRemediationResult{
		TotalProcessed: len(request.NonCompliantResults),
		SuccessCount:   len(request.NonCompliantResults),
		FailureCount:   0,
		Results:        []types.RemediationResult{},
	}, nil
}

func (m *MockComplianceService) GetNonCompliantResources(ctx context.Context, configRuleName string, region string) ([]types.NonCompliantResource, error) {
	// Mock implementation - return empty list for testing
	return []types.NonCompliantResource{}, nil
}

func (m *MockComplianceService) ValidateResourceExistence(ctx context.Context, resources []types.NonCompliantResource) ([]types.NonCompliantResource, error) {
	// Mock implementation - return all resources as valid
	return resources, nil
}

// Helper function to create int32 pointer
func intPtr(i int32) *int32 {
	return &i
}
