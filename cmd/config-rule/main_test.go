package main

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/zsoftly/logguardian/internal/types"
)

// MockConfigServiceClient implements a mock Config service client for testing
type MockConfigServiceClient struct {
	PutEvaluationsFunc func(ctx context.Context, params *configservice.PutEvaluationsInput, optFns ...func(*configservice.Options)) (*configservice.PutEvaluationsOutput, error)
}

func (m *MockConfigServiceClient) PutEvaluations(ctx context.Context, params *configservice.PutEvaluationsInput, optFns ...func(*configservice.Options)) (*configservice.PutEvaluationsOutput, error) {
	if m.PutEvaluationsFunc != nil {
		return m.PutEvaluationsFunc(ctx, params, optFns...)
	}
	return &configservice.PutEvaluationsOutput{}, nil
}

func TestEvaluateRetentionCompliance(t *testing.T) {
	handler := &CustomConfigRuleHandler{}

	tests := []struct {
		name               string
		retentionInDays    *int32
		minRetentionDays   int32
		expectedCompliance string
		expectedAnnotation string
	}{
		{
			name:               "null retention should be NON_COMPLIANT",
			retentionInDays:    nil,
			minRetentionDays:   30,
			expectedCompliance: "NON_COMPLIANT",
			expectedAnnotation: "No retention policy set (infinite retention). Minimum required: 30 days",
		},
		{
			name:               "retention below minimum should be NON_COMPLIANT",
			retentionInDays:    aws.Int32(7),
			minRetentionDays:   30,
			expectedCompliance: "NON_COMPLIANT",
			expectedAnnotation: "Retention period (7 days) below minimum requirement (30 days)",
		},
		{
			name:               "retention equal to minimum should be COMPLIANT",
			retentionInDays:    aws.Int32(30),
			minRetentionDays:   30,
			expectedCompliance: "COMPLIANT",
			expectedAnnotation: "Retention period (30 days) meets minimum requirement (30 days)",
		},
		{
			name:               "retention above minimum should be COMPLIANT",
			retentionInDays:    aws.Int32(90),
			minRetentionDays:   30,
			expectedCompliance: "COMPLIANT",
			expectedAnnotation: "Retention period (90 days) meets minimum requirement (30 days)",
		},
		{
			name:               "zero retention should be NON_COMPLIANT",
			retentionInDays:    aws.Int32(0),
			minRetentionDays:   1,
			expectedCompliance: "NON_COMPLIANT",
			expectedAnnotation: "Retention period (0 days) below minimum requirement (1 day)",
		},
		{
			name:               "very high retention should be COMPLIANT",
			retentionInDays:    aws.Int32(3653),
			minRetentionDays:   365,
			expectedCompliance: "COMPLIANT",
			expectedAnnotation: "Retention period (3653 days) meets minimum requirement (365 days)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configItem := types.ConfigurationItem{
				ResourceType: "AWS::Logs::LogGroup",
				ResourceId:   "/aws/lambda/test-function",
				ResourceName: "/aws/lambda/test-function",
				Configuration: types.LogGroupConfiguration{
					LogGroupName:    "/aws/lambda/test-function",
					RetentionInDays: tt.retentionInDays,
				},
			}

			complianceType, annotation := handler.evaluateRetentionCompliance(configItem, tt.minRetentionDays)

			if complianceType != tt.expectedCompliance {
				t.Errorf("Expected compliance type %s, got %s", tt.expectedCompliance, complianceType)
			}

			if annotation != tt.expectedAnnotation {
				t.Errorf("Expected annotation %q, got %q", tt.expectedAnnotation, annotation)
			}
		})
	}
}

func TestParseRuleParameters(t *testing.T) {
	handler := &CustomConfigRuleHandler{}

	tests := []struct {
		name           string
		inputParams    map[string]string
		expectedResult RetentionRuleParameters
		expectError    bool
	}{
		{
			name: "valid parameters",
			inputParams: map[string]string{
				"MinRetentionTime": "30",
			},
			expectedResult: RetentionRuleParameters{
				MinRetentionTime: 30,
			},
			expectError: false,
		},
		{
			name:        "empty parameters",
			inputParams: map[string]string{},
			expectedResult: RetentionRuleParameters{
				MinRetentionTime: 0,
			},
			expectError: false,
		},
		{
			name:        "nil parameters",
			inputParams: nil,
			expectedResult: RetentionRuleParameters{
				MinRetentionTime: 0,
			},
			expectError: false,
		},
		{
			name: "invalid number format",
			inputParams: map[string]string{
				"MinRetentionTime": "invalid",
			},
			expectedResult: RetentionRuleParameters{
				MinRetentionTime: 0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result RetentionRuleParameters
			err := handler.parseRuleParameters(tt.inputParams, &result)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && result.MinRetentionTime != tt.expectedResult.MinRetentionTime {
				t.Errorf("Expected MinRetentionTime %d, got %d", tt.expectedResult.MinRetentionTime, result.MinRetentionTime)
			}
		})
	}
}

func TestSubmitEvaluationResult(t *testing.T) {
	tests := []struct {
		name           string
		event          types.ConfigEvent
		configItem     types.ConfigurationItem
		complianceType string
		annotation     string
		mockError      error
		expectError    bool
	}{
		{
			name: "successful submission",
			event: types.ConfigEvent{
				ResultToken:    "test-token",
				ConfigRuleName: "test-rule",
			},
			configItem: types.ConfigurationItem{
				ResourceType: "AWS::Logs::LogGroup",
				ResourceId:   "/aws/lambda/test",
				ResourceName: "/aws/lambda/test",
			},
			complianceType: "COMPLIANT",
			annotation:     "Test annotation",
			mockError:      nil,
			expectError:    false,
		},
		{
			name: "failed submission",
			event: types.ConfigEvent{
				ResultToken:    "test-token",
				ConfigRuleName: "test-rule",
			},
			configItem: types.ConfigurationItem{
				ResourceType: "AWS::Logs::LogGroup",
				ResourceId:   "/aws/lambda/test",
				ResourceName: "/aws/lambda/test",
			},
			complianceType: "NON_COMPLIANT",
			annotation:     "Test annotation",
			mockError:      errors.New("service unavailable"),
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockConfigServiceClient{
				PutEvaluationsFunc: func(ctx context.Context, params *configservice.PutEvaluationsInput, optFns ...func(*configservice.Options)) (*configservice.PutEvaluationsOutput, error) {
					// Verify input parameters
					if len(params.Evaluations) != 1 {
						t.Errorf("Expected 1 evaluation, got %d", len(params.Evaluations))
					}

					evaluation := params.Evaluations[0]
					if *evaluation.ComplianceResourceId != tt.configItem.ResourceId {
						t.Errorf("Expected resource ID %s, got %s", tt.configItem.ResourceId, *evaluation.ComplianceResourceId)
					}
					if *evaluation.ComplianceResourceType != tt.configItem.ResourceType {
						t.Errorf("Expected resource type %s, got %s", tt.configItem.ResourceType, *evaluation.ComplianceResourceType)
					}
					if string(evaluation.ComplianceType) != tt.complianceType {
						t.Errorf("Expected compliance type %s, got %s", tt.complianceType, string(evaluation.ComplianceType))
					}
					if *evaluation.Annotation != tt.annotation {
						t.Errorf("Expected annotation %s, got %s", tt.annotation, *evaluation.Annotation)
					}
					if *params.ResultToken != tt.event.ResultToken {
						t.Errorf("Expected result token %s, got %s", tt.event.ResultToken, *params.ResultToken)
					}

					return &configservice.PutEvaluationsOutput{}, tt.mockError
				},
			}

			handler := &CustomConfigRuleHandler{
				configClient: mockClient,
			}

			err := handler.submitEvaluationResult(context.Background(), tt.event, tt.configItem, tt.complianceType, tt.annotation)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestHandleConfigRuleEvent_NonLogGroupResource(t *testing.T) {
	handler := &CustomConfigRuleHandler{
		configClient: &MockConfigServiceClient{},
	}

	// Test with non-log group resource
	event := types.ConfigEvent{
		ConfigRuleName: "test-rule",
		ResultToken:    "test-token",
		ConfigRuleInvokingEvent: types.ConfigRuleInvokingEvent{
			ConfigurationItem: types.ConfigurationItem{
				ResourceType:            "AWS::S3::Bucket",
				ResourceName:            "test-bucket",
				ConfigurationItemStatus: "ResourceDiscovered",
			},
		},
	}

	// Should return nil (no error) and not process the resource
	err := handler.HandleConfigRuleEvent(context.Background(), event)
	if err != nil {
		t.Errorf("Expected no error for non-log group resource, got: %v", err)
	}
}

func TestHandleConfigRuleEvent_DeletedResource(t *testing.T) {
	handler := &CustomConfigRuleHandler{
		configClient: &MockConfigServiceClient{},
	}

	// Test with deleted resource
	event := types.ConfigEvent{
		ConfigRuleName: "test-rule",
		ResultToken:    "test-token",
		ConfigRuleInvokingEvent: types.ConfigRuleInvokingEvent{
			ConfigurationItem: types.ConfigurationItem{
				ResourceType:            "AWS::Logs::LogGroup",
				ResourceName:            "/aws/lambda/deleted-function",
				ConfigurationItemStatus: "ResourceDeleted",
			},
		},
	}

	// Should return nil (no error) and not process the resource
	err := handler.HandleConfigRuleEvent(context.Background(), event)
	if err != nil {
		t.Errorf("Expected no error for deleted resource, got: %v", err)
	}
}

// Test for new validation functionality
func TestValidateEvent(t *testing.T) {
	handler := &CustomConfigRuleHandler{}

	tests := []struct {
		name        string
		event       types.ConfigEvent
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid event",
			event: types.ConfigEvent{
				ConfigRuleName: "test-rule",
				ResultToken:    "test-token",
				ConfigRuleInvokingEvent: types.ConfigRuleInvokingEvent{
					ConfigurationItem: types.ConfigurationItem{
						ResourceType: "AWS::Logs::LogGroup",
						ResourceName: "/aws/lambda/test",
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing result token",
			event: types.ConfigEvent{
				ConfigRuleName: "test-rule",
				ConfigRuleInvokingEvent: types.ConfigRuleInvokingEvent{
					ConfigurationItem: types.ConfigurationItem{
						ResourceType: "AWS::Logs::LogGroup",
						ResourceName: "/aws/lambda/test",
					},
				},
			},
			expectError: true,
			errorMsg:    "ResultToken is required",
		},
		{
			name: "missing config rule name",
			event: types.ConfigEvent{
				ResultToken: "test-token",
				ConfigRuleInvokingEvent: types.ConfigRuleInvokingEvent{
					ConfigurationItem: types.ConfigurationItem{
						ResourceType: "AWS::Logs::LogGroup",
						ResourceName: "/aws/lambda/test",
					},
				},
			},
			expectError: true,
			errorMsg:    "ConfigRuleName is required",
		},
		{
			name: "missing resource type",
			event: types.ConfigEvent{
				ConfigRuleName: "test-rule",
				ResultToken:    "test-token",
				ConfigRuleInvokingEvent: types.ConfigRuleInvokingEvent{
					ConfigurationItem: types.ConfigurationItem{
						ResourceName: "/aws/lambda/test",
					},
				},
			},
			expectError: true,
			errorMsg:    "ResourceType is required",
		},
		{
			name: "missing resource name",
			event: types.ConfigEvent{
				ConfigRuleName: "test-rule",
				ResultToken:    "test-token",
				ConfigRuleInvokingEvent: types.ConfigRuleInvokingEvent{
					ConfigurationItem: types.ConfigurationItem{
						ResourceType: "AWS::Logs::LogGroup",
					},
				},
			},
			expectError: true,
			errorMsg:    "ResourceName is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateEvent(tt.event)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// Test for constructor validation
func TestNewCustomConfigRuleHandler(t *testing.T) {
	tests := []struct {
		name        string
		cfg         aws.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			cfg: aws.Config{
				Region: "ca-central-1",
			},
			expectError: false,
		},
		{
			name:        "missing region",
			cfg:         aws.Config{},
			expectError: true,
			errorMsg:    "AWS region is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := NewCustomConfigRuleHandler(tt.cfg)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
				if handler != nil {
					t.Error("Expected nil handler on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if handler == nil {
					t.Error("Expected non-nil handler")
				}
			}
		})
	}
}

// Test parameter bounds validation
func TestParseRuleParameters_BoundsValidation(t *testing.T) {
	handler := &CustomConfigRuleHandler{}

	tests := []struct {
		name        string
		params      map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid retention days",
			params: map[string]string{
				"MinRetentionTime": "30",
			},
			expectError: false,
		},
		{
			name: "retention days too low",
			params: map[string]string{
				"MinRetentionTime": "0",
			},
			expectError: true,
			errorMsg:    "MinRetentionTime must be between 1 and 3653 days, got 0",
		},
		{
			name: "retention days too high",
			params: map[string]string{
				"MinRetentionTime": "5000",
			},
			expectError: true,
			errorMsg:    "MinRetentionTime must be between 1 and 3653 days, got 5000",
		},
		{
			name: "negative retention days",
			params: map[string]string{
				"MinRetentionTime": "-10",
			},
			expectError: true,
			errorMsg:    "MinRetentionTime must be between 1 and 3653 days, got -10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result RetentionRuleParameters
			err := handler.parseRuleParameters(tt.params, &result)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}
