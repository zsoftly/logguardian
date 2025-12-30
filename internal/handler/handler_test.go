package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// MockComplianceService provides a mock implementation for testing
// This mock supports both function-based behavior (for new tests) and field-based behavior (for existing tests)
type MockComplianceService struct {
	// Function-based behavior (takes precedence)
	RemediateLogGroupFunc                     func(ctx context.Context, compliance types.ComplianceResult) (*types.RemediationResult, error)
	GetNonCompliantResourcesFunc              func(ctx context.Context, configRuleName string, region string) ([]types.NonCompliantResource, error)
	ValidateResourceExistenceFunc             func(ctx context.Context, resources []types.NonCompliantResource) ([]types.NonCompliantResource, error)
	ProcessNonCompliantResourcesOptimizedFunc func(ctx context.Context, request types.BatchComplianceRequest) (*types.BatchRemediationResult, error)

	// Field-based behavior (for backward compatibility with existing tests)
	RemediateLogGroupCalled bool
	RemediateLogGroupError  error
	RemediateLogGroupResult *types.RemediationResult
}

func (m *MockComplianceService) RemediateLogGroup(ctx context.Context, compliance types.ComplianceResult) (*types.RemediationResult, error) {
	if m.RemediateLogGroupFunc != nil {
		return m.RemediateLogGroupFunc(ctx, compliance)
	}

	// Fall back to field-based behavior for backward compatibility
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

func (m *MockComplianceService) GetNonCompliantResources(ctx context.Context, configRuleName string, region string) ([]types.NonCompliantResource, error) {
	if m.GetNonCompliantResourcesFunc != nil {
		return m.GetNonCompliantResourcesFunc(ctx, configRuleName, region)
	}
	// Default behavior - return empty list for testing
	return []types.NonCompliantResource{}, nil
}

func (m *MockComplianceService) ValidateResourceExistence(ctx context.Context, resources []types.NonCompliantResource) ([]types.NonCompliantResource, error) {
	if m.ValidateResourceExistenceFunc != nil {
		return m.ValidateResourceExistenceFunc(ctx, resources)
	}
	// Default behavior - return all resources as valid
	return resources, nil
}

func (m *MockComplianceService) ProcessNonCompliantResourcesOptimized(ctx context.Context, request types.BatchComplianceRequest) (*types.BatchRemediationResult, error) {
	if m.ProcessNonCompliantResourcesOptimizedFunc != nil {
		return m.ProcessNonCompliantResourcesOptimizedFunc(ctx, request)
	}
	// Default behavior for optimized batch processing
	return &types.BatchRemediationResult{
		TotalProcessed: len(request.NonCompliantResults),
		SuccessCount:   len(request.NonCompliantResults),
		FailureCount:   0,
		Results:        []types.RemediationResult{},
	}, nil
}

// Helper function to create int32 pointer
func intPtr(i int32) *int32 {
	return &i
}

func TestComplianceHandler_HandleConfigRuleEvaluationRequest(t *testing.T) {
	tests := []struct {
		name                  string
		configRuleName        string
		region                string
		batchSize             int
		mockGetResources      []types.NonCompliantResource
		mockGetResourcesError error
		mockValidateResources []types.NonCompliantResource
		mockValidateError     error
		mockBatchResult       *types.BatchRemediationResult
		mockBatchError        error
		expectedError         bool
		errorContains         string
	}{
		{
			name:           "successful_batch_evaluation",
			configRuleName: "test-encryption-rule",
			region:         "ca-central-1",
			batchSize:      10,
			mockGetResources: []types.NonCompliantResource{
				{
					ResourceId:     "/aws/lambda/test-1",
					ResourceType:   "AWS::Logs::LogGroup",
					ResourceName:   "/aws/lambda/test-1",
					Region:         "ca-central-1",
					AccountId:      "123456789012",
					ComplianceType: "NON_COMPLIANT",
					LastEvaluated:  time.Now(),
				},
				{
					ResourceId:     "/aws/lambda/test-2",
					ResourceType:   "AWS::Logs::LogGroup",
					ResourceName:   "/aws/lambda/test-2",
					Region:         "ca-central-1",
					AccountId:      "123456789012",
					ComplianceType: "NON_COMPLIANT",
					LastEvaluated:  time.Now(),
				},
			},
			mockValidateResources: []types.NonCompliantResource{
				{
					ResourceId:     "/aws/lambda/test-1",
					ResourceType:   "AWS::Logs::LogGroup",
					ResourceName:   "/aws/lambda/test-1",
					Region:         "ca-central-1",
					AccountId:      "123456789012",
					ComplianceType: "NON_COMPLIANT",
					LastEvaluated:  time.Now(),
				},
				{
					ResourceId:     "/aws/lambda/test-2",
					ResourceType:   "AWS::Logs::LogGroup",
					ResourceName:   "/aws/lambda/test-2",
					Region:         "ca-central-1",
					AccountId:      "123456789012",
					ComplianceType: "NON_COMPLIANT",
					LastEvaluated:  time.Now(),
				},
			},
			mockBatchResult: &types.BatchRemediationResult{
				TotalProcessed: 2,
				SuccessCount:   2,
				FailureCount:   0,
			},
			expectedError: false,
		},
		{
			name:             "no_non_compliant_resources",
			configRuleName:   "test-encryption-rule",
			region:           "ca-central-1",
			batchSize:        10,
			mockGetResources: []types.NonCompliantResource{},
			expectedError:    false,
		},
		{
			name:           "all_resources_filtered_out",
			configRuleName: "test-encryption-rule",
			region:         "ca-central-1",
			batchSize:      10,
			mockGetResources: []types.NonCompliantResource{
				{
					ResourceId:     "/aws/lambda/test-1",
					ResourceType:   "AWS::Logs::LogGroup",
					ResourceName:   "/aws/lambda/test-1",
					Region:         "ca-central-1",
					AccountId:      "123456789012",
					ComplianceType: "NON_COMPLIANT",
					LastEvaluated:  time.Now(),
				},
			},
			mockValidateResources: []types.NonCompliantResource{}, // All filtered out
			expectedError:         false,
		},
		{
			name:                  "get_resources_error",
			configRuleName:        "test-encryption-rule",
			region:                "ca-central-1",
			batchSize:             10,
			mockGetResourcesError: fmt.Errorf("failed to get resources"),
			expectedError:         true,
			errorContains:         "failed to retrieve non-compliant resources",
		},
		{
			name:           "validate_resources_error",
			configRuleName: "test-encryption-rule",
			region:         "ca-central-1",
			batchSize:      10,
			mockGetResources: []types.NonCompliantResource{
				{
					ResourceId:     "/aws/lambda/test-1",
					ResourceType:   "AWS::Logs::LogGroup",
					ResourceName:   "/aws/lambda/test-1",
					Region:         "ca-central-1",
					AccountId:      "123456789012",
					ComplianceType: "NON_COMPLIANT",
					LastEvaluated:  time.Now(),
				},
			},
			mockValidateError: fmt.Errorf("failed to validate resources"),
			expectedError:     true,
			errorContains:     "failed to validate resource existence",
		},
		{
			name:           "batch_processing_error",
			configRuleName: "test-encryption-rule",
			region:         "ca-central-1",
			batchSize:      10,
			mockGetResources: []types.NonCompliantResource{
				{
					ResourceId:     "/aws/lambda/test-1",
					ResourceType:   "AWS::Logs::LogGroup",
					ResourceName:   "/aws/lambda/test-1",
					Region:         "ca-central-1",
					AccountId:      "123456789012",
					ComplianceType: "NON_COMPLIANT",
					LastEvaluated:  time.Now(),
				},
			},
			mockValidateResources: []types.NonCompliantResource{
				{
					ResourceId:     "/aws/lambda/test-1",
					ResourceType:   "AWS::Logs::LogGroup",
					ResourceName:   "/aws/lambda/test-1",
					Region:         "ca-central-1",
					AccountId:      "123456789012",
					ComplianceType: "NON_COMPLIANT",
					LastEvaluated:  time.Now(),
				},
			},
			mockBatchError: fmt.Errorf("batch processing failed"),
			expectedError:  true,
			errorContains:  "optimized batch processing failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockComplianceService{}
			handler := NewComplianceHandler(mockService)

			// Configure mock behavior
			mockService.GetNonCompliantResourcesFunc = func(ctx context.Context, configRuleName string, region string) ([]types.NonCompliantResource, error) {
				assert.Equal(t, tt.configRuleName, configRuleName)
				assert.Equal(t, tt.region, region)
				if tt.mockGetResourcesError != nil {
					return nil, tt.mockGetResourcesError
				}
				return tt.mockGetResources, nil
			}

			mockService.ValidateResourceExistenceFunc = func(ctx context.Context, resources []types.NonCompliantResource) ([]types.NonCompliantResource, error) {
				if tt.mockValidateError != nil {
					return nil, tt.mockValidateError
				}
				return tt.mockValidateResources, nil
			}

			mockService.ProcessNonCompliantResourcesOptimizedFunc = func(ctx context.Context, request types.BatchComplianceRequest) (*types.BatchRemediationResult, error) {
				assert.Equal(t, tt.configRuleName, request.ConfigRuleName)
				assert.Equal(t, tt.region, request.Region)
				assert.Equal(t, tt.batchSize, request.BatchSize)
				if tt.mockBatchError != nil {
					return nil, tt.mockBatchError
				}
				return tt.mockBatchResult, nil
			}

			err := handler.HandleConfigRuleEvaluationRequest(context.Background(), tt.configRuleName, tt.region, tt.batchSize)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestComplianceHandler_AnalyzeComplianceForRule(t *testing.T) {
	tests := []struct {
		name           string
		configRuleName string
		configItem     types.ConfigurationItem
		expectedResult types.ComplianceResult
	}{
		{
			name:           "encryption_rule_missing_encryption",
			configRuleName: "test-encryption-rule",
			configItem: types.ConfigurationItem{
				ResourceType: "AWS::Logs::LogGroup",
				ResourceName: "/aws/lambda/test",
				AwsRegion:    "ca-central-1",
				AwsAccountId: "123456789012",
				Configuration: types.LogGroupConfiguration{
					LogGroupName:    "/aws/lambda/test",
					KmsKeyId:        "", // Missing encryption
					RetentionInDays: intPtr(30),
				},
			},
			expectedResult: types.ComplianceResult{
				LogGroupName:      "/aws/lambda/test",
				Region:            "ca-central-1",
				AccountId:         "123456789012",
				MissingEncryption: true,
				MissingRetention:  false,
				CurrentRetention:  intPtr(30),
				CurrentKmsKeyId:   "",
			},
		},
		{
			name:           "encryption_rule_has_encryption",
			configRuleName: "test-encryption-rule",
			configItem: types.ConfigurationItem{
				ResourceType: "AWS::Logs::LogGroup",
				ResourceName: "/aws/lambda/test",
				AwsRegion:    "ca-central-1",
				AwsAccountId: "123456789012",
				Configuration: types.LogGroupConfiguration{
					LogGroupName:    "/aws/lambda/test",
					KmsKeyId:        "arn:aws:kms:ca-central-1:123456789012:key/test-key",
					RetentionInDays: intPtr(30),
				},
			},
			expectedResult: types.ComplianceResult{
				LogGroupName:      "/aws/lambda/test",
				Region:            "ca-central-1",
				AccountId:         "123456789012",
				MissingEncryption: false,
				MissingRetention:  false,
				CurrentRetention:  intPtr(30),
				CurrentKmsKeyId:   "arn:aws:kms:ca-central-1:123456789012:key/test-key",
			},
		},
		{
			name:           "retention_rule_missing_retention",
			configRuleName: "test-retention-rule",
			configItem: types.ConfigurationItem{
				ResourceType: "AWS::Logs::LogGroup",
				ResourceName: "/aws/lambda/test",
				AwsRegion:    "ca-central-1",
				AwsAccountId: "123456789012",
				Configuration: types.LogGroupConfiguration{
					LogGroupName:    "/aws/lambda/test",
					KmsKeyId:        "arn:aws:kms:ca-central-1:123456789012:key/test-key",
					RetentionInDays: nil, // Missing retention
				},
			},
			expectedResult: types.ComplianceResult{
				LogGroupName:      "/aws/lambda/test",
				Region:            "ca-central-1",
				AccountId:         "123456789012",
				MissingEncryption: false,
				MissingRetention:  true,
				CurrentRetention:  nil,
				CurrentKmsKeyId:   "arn:aws:kms:ca-central-1:123456789012:key/test-key",
			},
		},
		{
			name:           "retention_rule_has_retention",
			configRuleName: "test-retention-rule",
			configItem: types.ConfigurationItem{
				ResourceType: "AWS::Logs::LogGroup",
				ResourceName: "/aws/lambda/test",
				AwsRegion:    "ca-central-1",
				AwsAccountId: "123456789012",
				Configuration: types.LogGroupConfiguration{
					LogGroupName:    "/aws/lambda/test",
					KmsKeyId:        "arn:aws:kms:ca-central-1:123456789012:key/test-key",
					RetentionInDays: intPtr(365),
				},
			},
			expectedResult: types.ComplianceResult{
				LogGroupName:      "/aws/lambda/test",
				Region:            "ca-central-1",
				AccountId:         "123456789012",
				MissingEncryption: false,
				MissingRetention:  false,
				CurrentRetention:  intPtr(365),
				CurrentKmsKeyId:   "arn:aws:kms:ca-central-1:123456789012:key/test-key",
			},
		},
		{
			name:           "unknown_rule_type",
			configRuleName: "unknown-rule",
			configItem: types.ConfigurationItem{
				ResourceType: "AWS::Logs::LogGroup",
				ResourceName: "/aws/lambda/test",
				AwsRegion:    "ca-central-1",
				AwsAccountId: "123456789012",
				Configuration: types.LogGroupConfiguration{
					LogGroupName:    "/aws/lambda/test",
					KmsKeyId:        "",
					RetentionInDays: nil,
				},
			},
			expectedResult: types.ComplianceResult{
				LogGroupName:      "/aws/lambda/test",
				Region:            "ca-central-1",
				AccountId:         "123456789012",
				MissingEncryption: false,
				MissingRetention:  false,
				CurrentRetention:  nil,
				CurrentKmsKeyId:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockComplianceService{}
			handler := NewComplianceHandler(mockService)

			// Use reflection or expose the method, or test indirectly via HandleConfigEvent
			// Since analyzeComplianceForRule is private, we test it indirectly via HandleConfigEvent
			eventJSON, err := json.Marshal(types.ConfigEvent{
				ConfigRuleName: tt.configRuleName,
				AccountId:      tt.configItem.AwsAccountId,
				ConfigRuleInvokingEvent: types.ConfigRuleInvokingEvent{
					ConfigurationItem: tt.configItem,
				},
			})
			require.NoError(t, err)

			// Call HandleConfigEvent which internally calls analyzeComplianceForRule
			err = handler.HandleConfigEvent(context.Background(), eventJSON)

			// For unknown rules or compliant resources, no error expected
			// For non-compliant resources, remediation should be called (or error if remediation fails)
			if tt.configRuleName == "unknown-rule" {
				assert.NoError(t, err, "Unknown rules should not cause errors")
			} else if tt.expectedResult.MissingEncryption || tt.expectedResult.MissingRetention {
				// Should attempt remediation (success depends on mock)
				assert.NoError(t, err, "Remediation should succeed")
				assert.True(t, mockService.RemediateLogGroupCalled, "RemediateLogGroup should be called")
			} else {
				// Already compliant, no remediation needed
				assert.NoError(t, err, "Compliant resources should not cause errors")
				assert.False(t, mockService.RemediateLogGroupCalled, "RemediateLogGroup should not be called")
			}
		})
	}
}
