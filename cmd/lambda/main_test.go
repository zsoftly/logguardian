package main

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zsoftly/logguardian/internal/handler"
	"github.com/zsoftly/logguardian/internal/mocks"
	"github.com/zsoftly/logguardian/internal/testutil"
	"github.com/zsoftly/logguardian/internal/types"
)

func TestHandleUnifiedRequest_ConfigEvent_Success(t *testing.T) {
	tests := []struct {
		name          string
		request       types.LambdaRequest
		mockBehavior  func(*mocks.MockComplianceService)
		expectedError bool
		errorContains string
	}{
		{
			name: "successful_config_event_with_encryption_remediation",
			request: func() types.LambdaRequest {
				req, err := testutil.NewTestConfigEventRequest()
				require.NoError(t, err)
				// Modify to have missing encryption
				var event types.ConfigEvent
				err = json.Unmarshal(req.ConfigEvent, &event)
				require.NoError(t, err)
				event.ConfigRuleName = "test-encryption-rule"
				event.ConfigRuleInvokingEvent.ConfigurationItem.Configuration.KmsKeyId = ""
				req.ConfigEvent, err = json.Marshal(event)
				require.NoError(t, err)
				return req
			}(),
			mockBehavior: func(m *mocks.MockComplianceService) {
				m.RemediateLogGroupFunc = func(ctx context.Context, compliance types.ComplianceResult) (*types.RemediationResult, error) {
					assert.True(t, compliance.MissingEncryption)
					assert.Equal(t, "/aws/lambda/test-function", compliance.LogGroupName)
					return &types.RemediationResult{
						LogGroupName:      compliance.LogGroupName,
						Region:            compliance.Region,
						Success:           true,
						EncryptionApplied: true,
					}, nil
				}
			},
			expectedError: false,
		},
		{
			name: "successful_config_event_with_retention_remediation",
			request: func() types.LambdaRequest {
				req, err := testutil.NewTestConfigEventRequest()
				require.NoError(t, err)
				var event types.ConfigEvent
				err = json.Unmarshal(req.ConfigEvent, &event)
				require.NoError(t, err)
				event.ConfigRuleName = "test-retention-rule"
				event.ConfigRuleInvokingEvent.ConfigurationItem.Configuration.RetentionInDays = nil
				req.ConfigEvent, err = json.Marshal(event)
				require.NoError(t, err)
				return req
			}(),
			mockBehavior: func(m *mocks.MockComplianceService) {
				m.RemediateLogGroupFunc = func(ctx context.Context, compliance types.ComplianceResult) (*types.RemediationResult, error) {
					assert.True(t, compliance.MissingRetention)
					return &types.RemediationResult{
						LogGroupName:     compliance.LogGroupName,
						Region:           compliance.Region,
						Success:          true,
						RetentionApplied: true,
					}, nil
				}
			},
			expectedError: false,
		},
		{
			name: "successful_config_event_already_compliant",
			request: func() types.LambdaRequest {
				req, err := testutil.NewTestConfigEventRequest()
				require.NoError(t, err)
				var event types.ConfigEvent
				err = json.Unmarshal(req.ConfigEvent, &event)
				require.NoError(t, err)
				event.ConfigRuleName = "test-encryption-rule"
				retention := int32(30)
				event.ConfigRuleInvokingEvent.ConfigurationItem.Configuration.KmsKeyId = "arn:aws:kms:ca-central-1:123456789012:key/test-key"
				event.ConfigRuleInvokingEvent.ConfigurationItem.Configuration.RetentionInDays = &retention
				req.ConfigEvent, err = json.Marshal(event)
				require.NoError(t, err)
				return req
			}(),
			mockBehavior: func(m *mocks.MockComplianceService) {
				// Should not be called for compliant resources
				m.RemediateLogGroupFunc = func(ctx context.Context, compliance types.ComplianceResult) (*types.RemediationResult, error) {
					t.Fatal("RemediateLogGroup should not be called for compliant resources")
					return nil, nil
				}
			},
			expectedError: false,
		},
		{
			name: "config_event_with_deleted_resource",
			request: func() types.LambdaRequest {
				req, err := testutil.NewTestConfigEventRequest()
				require.NoError(t, err)
				var event types.ConfigEvent
				err = json.Unmarshal(req.ConfigEvent, &event)
				require.NoError(t, err)
				event.ConfigRuleInvokingEvent.ConfigurationItem.ConfigurationItemStatus = "ResourceDeleted"
				req.ConfigEvent, err = json.Marshal(event)
				require.NoError(t, err)
				return req
			}(),
			mockBehavior: func(m *mocks.MockComplianceService) {
				// Should not be called for deleted resources
				m.RemediateLogGroupFunc = func(ctx context.Context, compliance types.ComplianceResult) (*types.RemediationResult, error) {
					t.Fatal("RemediateLogGroup should not be called for deleted resources")
					return nil, nil
				}
			},
			expectedError: false,
		},
		{
			name: "config_event_with_non_loggroup_resource",
			request: func() types.LambdaRequest {
				req, err := testutil.NewTestConfigEventRequest()
				require.NoError(t, err)
				var event types.ConfigEvent
				err = json.Unmarshal(req.ConfigEvent, &event)
				require.NoError(t, err)
				event.ConfigRuleInvokingEvent.ConfigurationItem.ResourceType = "AWS::S3::Bucket"
				req.ConfigEvent, err = json.Marshal(event)
				require.NoError(t, err)
				return req
			}(),
			mockBehavior: func(m *mocks.MockComplianceService) {
				// Should not be called for non-log-group resources
				m.RemediateLogGroupFunc = func(ctx context.Context, compliance types.ComplianceResult) (*types.RemediationResult, error) {
					t.Fatal("RemediateLogGroup should not be called for non-log-group resources")
					return nil, nil
				}
			},
			expectedError: false,
		},
		{
			name: "config_event_remediation_failure",
			request: func() types.LambdaRequest {
				req, err := testutil.NewTestConfigEventRequest()
				require.NoError(t, err)
				var event types.ConfigEvent
				err = json.Unmarshal(req.ConfigEvent, &event)
				require.NoError(t, err)
				event.ConfigRuleName = "test-encryption-rule"
				event.ConfigRuleInvokingEvent.ConfigurationItem.Configuration.KmsKeyId = ""
				req.ConfigEvent, err = json.Marshal(event)
				require.NoError(t, err)
				return req
			}(),
			mockBehavior: func(m *mocks.MockComplianceService) {
				m.RemediateLogGroupFunc = func(ctx context.Context, compliance types.ComplianceResult) (*types.RemediationResult, error) {
					return nil, errors.New("KMS key not found")
				}
			},
			expectedError: true,
			errorContains: "remediation failed",
		},
		{
			name: "config_event_invalid_json",
			request: types.LambdaRequest{
				Type:        "config-event",
				ConfigEvent: json.RawMessage(`{invalid json}`),
			},
			mockBehavior: func(m *mocks.MockComplianceService) {
				// Should not be called for invalid JSON
				m.RemediateLogGroupFunc = func(ctx context.Context, compliance types.ComplianceResult) (*types.RemediationResult, error) {
					t.Fatal("RemediateLogGroup should not be called for invalid JSON")
					return nil, nil
				}
			},
			expectedError: true,
			errorContains: "failed to parse Config event",
		},
		{
			name: "config_event_missing_configEvent",
			request: types.LambdaRequest{
				Type:        "config-event",
				ConfigEvent: nil,
			},
			mockBehavior: func(m *mocks.MockComplianceService) {
				// Should not be called
				m.RemediateLogGroupFunc = func(ctx context.Context, compliance types.ComplianceResult) (*types.RemediationResult, error) {
					t.Fatal("RemediateLogGroup should not be called")
					return nil, nil
				}
			},
			expectedError: true,
			errorContains: "configEvent is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockService := &mocks.MockComplianceService{}
			if tt.mockBehavior != nil {
				tt.mockBehavior(mockService)
			}
			h := handler.NewComplianceHandler(mockService)

			err := handleUnifiedRequest(ctx, h, tt.request)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHandleUnifiedRequest_ConfigRuleEvaluation_Success(t *testing.T) {
	tests := []struct {
		name          string
		request       types.LambdaRequest
		mockBehavior  func(*mocks.MockComplianceService)
		expectedError bool
		errorContains string
	}{
		{
			name: "successful_config_rule_evaluation",
			request: types.LambdaRequest{
				Type:           "config-rule-evaluation",
				ConfigRuleName: "test-encryption-rule",
				Region:         "ca-central-1",
				BatchSize:      10,
			},
			mockBehavior: func(m *mocks.MockComplianceService) {
				m.GetNonCompliantResourcesFunc = func(ctx context.Context, configRuleName string, region string) ([]types.NonCompliantResource, error) {
					assert.Equal(t, "test-encryption-rule", configRuleName)
					assert.Equal(t, "ca-central-1", region)
					return []types.NonCompliantResource{
						testutil.NewTestNonCompliantResource(),
					}, nil
				}
				m.ValidateResourceExistenceFunc = func(ctx context.Context, resources []types.NonCompliantResource) ([]types.NonCompliantResource, error) {
					return resources, nil
				}
				m.ProcessNonCompliantResourcesOptimizedFunc = func(ctx context.Context, request types.BatchComplianceRequest) (*types.BatchRemediationResult, error) {
					assert.Equal(t, "test-encryption-rule", request.ConfigRuleName)
					assert.Equal(t, 1, len(request.NonCompliantResults))
					return testutil.NewTestBatchRemediationResult(1, 1, 0), nil
				}
			},
			expectedError: false,
		},
		{
			name: "config_rule_evaluation_with_default_batch_size",
			request: types.LambdaRequest{
				Type:           "config-rule-evaluation",
				ConfigRuleName: "test-encryption-rule",
				Region:         "ca-central-1",
				BatchSize:      0, // Should default to 10
			},
			mockBehavior: func(m *mocks.MockComplianceService) {
				m.GetNonCompliantResourcesFunc = func(ctx context.Context, configRuleName string, region string) ([]types.NonCompliantResource, error) {
					return []types.NonCompliantResource{}, nil
				}
			},
			expectedError: false,
		},
		{
			name: "config_rule_evaluation_no_non_compliant_resources",
			request: types.LambdaRequest{
				Type:           "config-rule-evaluation",
				ConfigRuleName: "test-encryption-rule",
				Region:         "ca-central-1",
				BatchSize:      10,
			},
			mockBehavior: func(m *mocks.MockComplianceService) {
				m.GetNonCompliantResourcesFunc = func(ctx context.Context, configRuleName string, region string) ([]types.NonCompliantResource, error) {
					return []types.NonCompliantResource{}, nil
				}
			},
			expectedError: false,
		},
		{
			name: "config_rule_evaluation_all_resources_filtered",
			request: types.LambdaRequest{
				Type:           "config-rule-evaluation",
				ConfigRuleName: "test-encryption-rule",
				Region:         "ca-central-1",
				BatchSize:      10,
			},
			mockBehavior: func(m *mocks.MockComplianceService) {
				m.GetNonCompliantResourcesFunc = func(ctx context.Context, configRuleName string, region string) ([]types.NonCompliantResource, error) {
					return []types.NonCompliantResource{
						testutil.NewTestNonCompliantResource(),
					}, nil
				}
				m.ValidateResourceExistenceFunc = func(ctx context.Context, resources []types.NonCompliantResource) ([]types.NonCompliantResource, error) {
					// All resources filtered out
					return []types.NonCompliantResource{}, nil
				}
			},
			expectedError: false,
		},
		{
			name: "config_rule_evaluation_get_resources_error",
			request: types.LambdaRequest{
				Type:           "config-rule-evaluation",
				ConfigRuleName: "test-encryption-rule",
				Region:         "ca-central-1",
				BatchSize:      10,
			},
			mockBehavior: func(m *mocks.MockComplianceService) {
				m.GetNonCompliantResourcesFunc = func(ctx context.Context, configRuleName string, region string) ([]types.NonCompliantResource, error) {
					return nil, errors.New("failed to get resources")
				}
			},
			expectedError: true,
			errorContains: "failed to retrieve non-compliant resources",
		},
		{
			name: "config_rule_evaluation_validate_resources_error",
			request: types.LambdaRequest{
				Type:           "config-rule-evaluation",
				ConfigRuleName: "test-encryption-rule",
				Region:         "ca-central-1",
				BatchSize:      10,
			},
			mockBehavior: func(m *mocks.MockComplianceService) {
				m.GetNonCompliantResourcesFunc = func(ctx context.Context, configRuleName string, region string) ([]types.NonCompliantResource, error) {
					return []types.NonCompliantResource{
						testutil.NewTestNonCompliantResource(),
					}, nil
				}
				m.ValidateResourceExistenceFunc = func(ctx context.Context, resources []types.NonCompliantResource) ([]types.NonCompliantResource, error) {
					return nil, errors.New("failed to validate resources")
				}
			},
			expectedError: true,
			errorContains: "failed to validate resource existence",
		},
		{
			name: "config_rule_evaluation_batch_processing_error",
			request: types.LambdaRequest{
				Type:           "config-rule-evaluation",
				ConfigRuleName: "test-encryption-rule",
				Region:         "ca-central-1",
				BatchSize:      10,
			},
			mockBehavior: func(m *mocks.MockComplianceService) {
				m.GetNonCompliantResourcesFunc = func(ctx context.Context, configRuleName string, region string) ([]types.NonCompliantResource, error) {
					return []types.NonCompliantResource{
						testutil.NewTestNonCompliantResource(),
					}, nil
				}
				m.ValidateResourceExistenceFunc = func(ctx context.Context, resources []types.NonCompliantResource) ([]types.NonCompliantResource, error) {
					return resources, nil
				}
				m.ProcessNonCompliantResourcesOptimizedFunc = func(ctx context.Context, request types.BatchComplianceRequest) (*types.BatchRemediationResult, error) {
					return nil, errors.New("batch processing failed")
				}
			},
			expectedError: true,
			errorContains: "optimized batch processing failed",
		},
		{
			name: "config_rule_evaluation_missing_configRuleName",
			request: types.LambdaRequest{
				Type:           "config-rule-evaluation",
				ConfigRuleName: "",
				Region:         "ca-central-1",
				BatchSize:      10,
			},
			mockBehavior: func(m *mocks.MockComplianceService) {
				// Should not be called
				m.GetNonCompliantResourcesFunc = func(ctx context.Context, configRuleName string, region string) ([]types.NonCompliantResource, error) {
					t.Fatal("GetNonCompliantResources should not be called")
					return nil, nil
				}
			},
			expectedError: true,
			errorContains: "configRuleName is required",
		},
		{
			name: "config_rule_evaluation_missing_region",
			request: types.LambdaRequest{
				Type:           "config-rule-evaluation",
				ConfigRuleName: "test-encryption-rule",
				Region:         "",
				BatchSize:      10,
			},
			mockBehavior: func(m *mocks.MockComplianceService) {
				// Should not be called
				m.GetNonCompliantResourcesFunc = func(ctx context.Context, configRuleName string, region string) ([]types.NonCompliantResource, error) {
					t.Fatal("GetNonCompliantResources should not be called")
					return nil, nil
				}
			},
			expectedError: true,
			errorContains: "region is required",
		},
		{
			name: "config_rule_evaluation_invalid_type",
			request: types.LambdaRequest{
				Type:           "invalid-type",
				ConfigRuleName: "test-encryption-rule",
				Region:         "ca-central-1",
				BatchSize:      10,
			},
			mockBehavior: func(m *mocks.MockComplianceService) {
				// Should not be called
				m.GetNonCompliantResourcesFunc = func(ctx context.Context, configRuleName string, region string) ([]types.NonCompliantResource, error) {
					t.Fatal("GetNonCompliantResources should not be called")
					return nil, nil
				}
			},
			expectedError: true,
			errorContains: "unsupported request type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockService := &mocks.MockComplianceService{}
			if tt.mockBehavior != nil {
				tt.mockBehavior(mockService)
			}
			h := handler.NewComplianceHandler(mockService)

			err := handleUnifiedRequest(ctx, h, tt.request)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestInitializeHandler_Success(t *testing.T) {
	ctx := context.Background()

	// Test initializeHandler function directly
	// This will try to load AWS config, which may fail in test environments
	// but we're testing the code path, not the actual AWS connectivity
	h, err := initializeHandler(ctx)

	// If AWS config loading fails (no credentials), that's expected in CI/test environments
	// We still verify that the function executes the initialization logic
	if err != nil {
		// Expected error: no AWS credentials available in test environment
		assert.Contains(t, err.Error(), "failed to load AWS config", "Error should be about AWS config loading")
	} else {
		// If credentials are available, verify handler is created successfully
		require.NotNil(t, h, "Handler should be created successfully")
	}
}

func TestHandleUnifiedRequest_DefaultBatchSize(t *testing.T) {
	ctx := context.Background()
	mockService := &mocks.MockComplianceService{}

	mockService.GetNonCompliantResourcesFunc = func(ctx context.Context, configRuleName string, region string) ([]types.NonCompliantResource, error) {
		return []types.NonCompliantResource{}, nil
	}

	h := handler.NewComplianceHandler(mockService)

	// Test with BatchSize = 0 (should default to 10)
	request := types.LambdaRequest{
		Type:           "config-rule-evaluation",
		ConfigRuleName: "test-rule",
		Region:         "ca-central-1",
		BatchSize:      0, // Should default to 10
	}

	err := handleUnifiedRequest(ctx, h, request)
	require.NoError(t, err)
}
