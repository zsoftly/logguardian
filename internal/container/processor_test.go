package container

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zsoftly/logguardian/internal/types"
)

// MockComplianceService is a mock implementation of ComplianceServiceInterface
type MockComplianceService struct {
	mock.Mock
}

func (m *MockComplianceService) GetNonCompliantResources(ctx context.Context, configRuleName, region string) ([]types.NonCompliantResource, error) {
	args := m.Called(ctx, configRuleName, region)
	return args.Get(0).([]types.NonCompliantResource), args.Error(1)
}

func (m *MockComplianceService) ValidateResourceExistence(ctx context.Context, resources []types.NonCompliantResource) ([]types.NonCompliantResource, error) {
	args := m.Called(ctx, resources)
	return args.Get(0).([]types.NonCompliantResource), args.Error(1)
}

func (m *MockComplianceService) RemediateLogGroup(ctx context.Context, compliance types.ComplianceResult) (*types.RemediationResult, error) {
	args := m.Called(ctx, compliance)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.RemediationResult), args.Error(1)
}

func (m *MockComplianceService) ProcessNonCompliantResourcesOptimized(ctx context.Context, request types.BatchComplianceRequest) (*types.BatchRemediationResult, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.BatchRemediationResult), args.Error(1)
}

func (m *MockComplianceService) EvaluateCompliance(ctx context.Context, logGroupName, region string) (types.ComplianceResult, error) {
	args := m.Called(ctx, logGroupName, region)
	return args.Get(0).(types.ComplianceResult), args.Error(1)
}

func TestNewCommandProcessor(t *testing.T) {
	tests := []struct {
		name    string
		options ProcessorOptions
	}{
		{
			name: "regular mode",
			options: ProcessorOptions{
				DryRun:       false,
				ExecutionID:  "test-123",
				OutputFormat: "json",
			},
		},
		{
			name: "dry-run mode",
			options: ProcessorOptions{
				DryRun:       true,
				ExecutionID:  "test-456",
				OutputFormat: "text",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := aws.Config{}
			processor := NewCommandProcessor(cfg, tt.options)

			assert.NotNil(t, processor)
			assert.NotNil(t, processor.handler)
			assert.NotNil(t, processor.service)
			assert.Equal(t, tt.options.ExecutionID, processor.options.ExecutionID)
			assert.Equal(t, tt.options.DryRun, processor.options.DryRun)
		})
	}
}

func TestCommandProcessor_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		request        CommandRequest
		options        ProcessorOptions
		mockSetup      func(*MockComplianceService)
		expectedStatus string
		expectedError  bool
	}{
		{
			name: "successful execution",
			request: CommandRequest{
				Type:           "config-rule-evaluation",
				ConfigRuleName: "test-rule",
				Region:         "us-east-1",
				BatchSize:      10,
			},
			options: ProcessorOptions{
				DryRun:       false,
				ExecutionID:  "test-123",
				OutputFormat: "json",
			},
			mockSetup: func(m *MockComplianceService) {
				// Mock GetNonCompliantResources
				nonCompliant := []types.NonCompliantResource{
					{
						ResourceId:   "log-group-1",
						ResourceName: "log-group-1",
						ResourceType: "AWS::Logs::LogGroup",
						Region:       "us-east-1",
					},
				}
				m.On("GetNonCompliantResources", ctx, "test-rule", "us-east-1").
					Return(nonCompliant, nil)

				// Mock ValidateResourceExistence
				m.On("ValidateResourceExistence", ctx, nonCompliant).
					Return(nonCompliant, nil)

				// Mock ProcessNonCompliantResourcesOptimized
				batchRequest := types.BatchComplianceRequest{
					ConfigRuleName:      "test-rule",
					NonCompliantResults: nonCompliant,
					Region:              "us-east-1",
					BatchSize:           10,
				}
				batchResult := &types.BatchRemediationResult{
					TotalProcessed: 1,
					SuccessCount:   1,
					FailureCount:   0,
					Results: []types.RemediationResult{
						{
							LogGroupName:      "log-group-1",
							Region:            "us-east-1",
							EncryptionApplied: true,
							Success:           true,
						},
					},
				}
				m.On("ProcessNonCompliantResourcesOptimized", ctx, batchRequest).
					Return(batchResult, nil)
			},
			expectedStatus: "completed",
			expectedError:  false,
		},
		{
			name: "no non-compliant resources",
			request: CommandRequest{
				Type:           "config-rule-evaluation",
				ConfigRuleName: "test-rule",
				Region:         "us-east-1",
				BatchSize:      10,
			},
			options: ProcessorOptions{
				DryRun:       false,
				ExecutionID:  "test-456",
				OutputFormat: "json",
			},
			mockSetup: func(m *MockComplianceService) {
				// Mock GetNonCompliantResources returning empty
				m.On("GetNonCompliantResources", ctx, "test-rule", "us-east-1").
					Return([]types.NonCompliantResource{}, nil)
			},
			expectedStatus: "completed",
			expectedError:  false,
		},
		{
			name: "unsupported request type",
			request: CommandRequest{
				Type: "unsupported-type",
			},
			options: ProcessorOptions{
				DryRun:       false,
				ExecutionID:  "test-789",
				OutputFormat: "json",
			},
			mockSetup:      func(m *MockComplianceService) {},
			expectedStatus: "failed",
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockComplianceService)
			tt.mockSetup(mockService)

			// Create processor with mock service
			processor := &CommandProcessor{
				service:      mockService,
				options:      tt.options,
				executionLog: []ExecutionLogEntry{},
			}

			// Need to create handler manually since we're using mock
			cfg := aws.Config{}
			realProcessor := NewCommandProcessor(cfg, tt.options)
			processor.handler = realProcessor.handler

			result, err := processor.Execute(ctx, tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedStatus, result.Status)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedStatus, result.Status)
				assert.Equal(t, tt.options.ExecutionID, result.ExecutionID)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestCommandProcessor_GetMode(t *testing.T) {
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
			processor := &CommandProcessor{
				options: ProcessorOptions{
					DryRun: tt.dryRun,
				},
			}
			assert.Equal(t, tt.expected, processor.getMode())
		})
	}
}

func TestGetResourceStatus(t *testing.T) {
	tests := []struct {
		name     string
		result   types.RemediationResult
		expected string
	}{
		{
			name: "successful result",
			result: types.RemediationResult{
				Success: true,
			},
			expected: "success",
		},
		{
			name: "failed result",
			result: types.RemediationResult{
				Success: false,
			},
			expected: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, getResourceStatus(tt.result))
		})
	}
}

func TestCommandProcessor_LogEntry(t *testing.T) {
	processor := &CommandProcessor{
		executionLog: []ExecutionLogEntry{},
	}

	testCases := []struct {
		level   string
		message string
		details any
	}{
		{"INFO", "Test info message", map[string]any{"key": "value"}},
		{"WARN", "Test warning", nil},
		{"ERROR", "Test error", "error details"},
		{"DEBUG", "Test debug", 123},
	}

	for i, tc := range testCases {
		processor.logEntry(tc.level, tc.message, tc.details)

		assert.Len(t, processor.executionLog, i+1)
		entry := processor.executionLog[i]
		assert.Equal(t, tc.level, entry.Level)
		assert.Equal(t, tc.message, entry.Message)
		assert.Equal(t, tc.details, entry.Details)
		assert.WithinDuration(t, time.Now(), entry.Timestamp, time.Second)
	}
}
