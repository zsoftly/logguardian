package container

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zsoftly/logguardian/internal/types"
)

func TestNewDryRunComplianceService(t *testing.T) {
	mockService := new(MockComplianceService)
	dryRunService := NewDryRunComplianceService(mockService)

	assert.NotNil(t, dryRunService)
	assert.Equal(t, mockService, dryRunService.realService)
}

func TestDryRunComplianceService_GetNonCompliantResources(t *testing.T) {
	ctx := context.Background()
	mockService := new(MockComplianceService)
	dryRunService := NewDryRunComplianceService(mockService)

	expectedResources := []types.NonCompliantResource{
		{
			ResourceId:   "log-group-1",
			ResourceName: "log-group-1",
			ResourceType: "AWS::Logs::LogGroup",
			Region:       "us-east-1",
		},
		{
			ResourceId:   "log-group-2",
			ResourceName: "log-group-2",
			ResourceType: "AWS::Logs::LogGroup",
			Region:       "us-east-1",
		},
	}

	mockService.On("GetNonCompliantResources", ctx, "test-rule", "us-east-1").
		Return(expectedResources, nil)

	// This should delegate to the real service
	result, err := dryRunService.GetNonCompliantResources(ctx, "test-rule", "us-east-1")

	assert.NoError(t, err)
	assert.Equal(t, expectedResources, result)
	mockService.AssertExpectations(t)
}

func TestDryRunComplianceService_ValidateResourceExistence(t *testing.T) {
	ctx := context.Background()
	mockService := new(MockComplianceService)
	dryRunService := NewDryRunComplianceService(mockService)

	inputResources := []types.NonCompliantResource{
		{ResourceId: "log-group-1", ResourceName: "log-group-1"},
		{ResourceId: "log-group-2", ResourceName: "log-group-2"},
	}

	expectedResources := []types.NonCompliantResource{
		{ResourceId: "log-group-1", ResourceName: "log-group-1"},
	}

	mockService.On("ValidateResourceExistence", ctx, inputResources).
		Return(expectedResources, nil)

	// This should delegate to the real service
	result, err := dryRunService.ValidateResourceExistence(ctx, inputResources)

	assert.NoError(t, err)
	assert.Equal(t, expectedResources, result)
	mockService.AssertExpectations(t)
}

func TestDryRunComplianceService_RemediateLogGroup(t *testing.T) {
	ctx := context.Background()
	dryRunService := NewDryRunComplianceService(nil) // Don't need real service for dry-run

	tests := []struct {
		name       string
		compliance types.ComplianceResult
		expected   *types.RemediationResult
	}{
		{
			name: "needs encryption and retention",
			compliance: types.ComplianceResult{
				LogGroupName:      "test-log-group",
				Region:            "us-east-1",
				MissingEncryption: true,
				MissingRetention:  true,
			},
			expected: &types.RemediationResult{
				LogGroupName:      "test-log-group",
				Region:            "us-east-1",
				EncryptionApplied: true,
				RetentionApplied:  true,
				Success:           true,
				Error:             nil,
			},
		},
		{
			name: "needs only encryption",
			compliance: types.ComplianceResult{
				LogGroupName:      "test-log-group",
				Region:            "us-east-1",
				MissingEncryption: true,
				MissingRetention:  false,
			},
			expected: &types.RemediationResult{
				LogGroupName:      "test-log-group",
				Region:            "us-east-1",
				EncryptionApplied: true,
				RetentionApplied:  false,
				Success:           true,
				Error:             nil,
			},
		},
		{
			name: "needs only retention",
			compliance: types.ComplianceResult{
				LogGroupName:      "test-log-group",
				Region:            "us-east-1",
				MissingEncryption: false,
				MissingRetention:  true,
			},
			expected: &types.RemediationResult{
				LogGroupName:      "test-log-group",
				Region:            "us-east-1",
				EncryptionApplied: false,
				RetentionApplied:  true,
				Success:           true,
				Error:             nil,
			},
		},
		{
			name: "already compliant",
			compliance: types.ComplianceResult{
				LogGroupName:      "test-log-group",
				Region:            "us-east-1",
				MissingEncryption: false,
				MissingRetention:  false,
			},
			expected: &types.RemediationResult{
				LogGroupName:      "test-log-group",
				Region:            "us-east-1",
				EncryptionApplied: false,
				RetentionApplied:  false,
				Success:           true,
				Error:             nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dryRunService.RemediateLogGroup(ctx, tt.compliance)

			assert.NoError(t, err)
			assert.Equal(t, *tt.expected, *result)
		})
	}
}

func TestDryRunComplianceService_ProcessNonCompliantResourcesOptimized(t *testing.T) {
	ctx := context.Background()
	dryRunService := NewDryRunComplianceService(nil) // Don't need real service for dry-run

	request := types.BatchComplianceRequest{
		ConfigRuleName: "test-rule",
		Region:         "us-east-1",
		BatchSize:      10,
		NonCompliantResults: []types.NonCompliantResource{
			{
				ResourceId:   "log-group-1",
				ResourceName: "log-group-1",
				ResourceType: "AWS::Logs::LogGroup",
				Region:       "us-east-1",
				Annotation:   "Missing encryption",
			},
			{
				ResourceId:   "log-group-2",
				ResourceName: "log-group-2",
				ResourceType: "AWS::Logs::LogGroup",
				Region:       "us-east-1",
				Annotation:   "Missing retention",
			},
			{
				ResourceId:   "log-group-3",
				ResourceName: "log-group-3",
				ResourceType: "AWS::Logs::LogGroup",
				Region:       "us-east-1",
				Annotation:   "",
			},
		},
	}

	result, err := dryRunService.ProcessNonCompliantResourcesOptimized(ctx, request)

	assert.NoError(t, err)
	assert.Equal(t, 3, result.TotalProcessed)
	assert.Equal(t, 3, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Len(t, result.Results, 3)

	// Check individual results
	for i, r := range result.Results {
		assert.Equal(t, request.NonCompliantResults[i].ResourceName, r.LogGroupName)
		assert.Equal(t, request.NonCompliantResults[i].Region, r.Region)
		assert.True(t, r.Success)
		assert.Nil(t, r.Error)

		// Resources with annotations should have remediation flags set
		if request.NonCompliantResults[i].Annotation != "" {
			assert.True(t, r.EncryptionApplied || r.RetentionApplied)
		}
	}
}
