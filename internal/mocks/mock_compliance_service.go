package mocks

import (
	"context"

	"github.com/zsoftly/logguardian/internal/service"
	"github.com/zsoftly/logguardian/internal/types"
)

// MockComplianceService is a mock implementation of ComplianceServiceInterface
// for unit testing Lambda handlers without real AWS service calls.
type MockComplianceService struct {
	// RemediateLogGroupFunc allows tests to control the behavior of RemediateLogGroup
	RemediateLogGroupFunc func(ctx context.Context, compliance types.ComplianceResult) (*types.RemediationResult, error)

	// ProcessNonCompliantResourcesOptimizedFunc allows tests to control batch processing
	ProcessNonCompliantResourcesOptimizedFunc func(ctx context.Context, request types.BatchComplianceRequest) (*types.BatchRemediationResult, error)

	// GetNonCompliantResourcesFunc allows tests to control resource retrieval
	GetNonCompliantResourcesFunc func(ctx context.Context, configRuleName string, region string) ([]types.NonCompliantResource, error)

	// ValidateResourceExistenceFunc allows tests to control resource validation
	ValidateResourceExistenceFunc func(ctx context.Context, resources []types.NonCompliantResource) ([]types.NonCompliantResource, error)
}

// RemediateLogGroup implements ComplianceServiceInterface
func (m *MockComplianceService) RemediateLogGroup(ctx context.Context, compliance types.ComplianceResult) (*types.RemediationResult, error) {
	if m.RemediateLogGroupFunc != nil {
		return m.RemediateLogGroupFunc(ctx, compliance)
	}
	// Default behavior: return success
	return &types.RemediationResult{
		LogGroupName:      compliance.LogGroupName,
		Region:            compliance.Region,
		Success:           true,
		EncryptionApplied: compliance.MissingEncryption,
		RetentionApplied:  compliance.MissingRetention,
	}, nil
}

// ProcessNonCompliantResourcesOptimized implements ComplianceServiceInterface
func (m *MockComplianceService) ProcessNonCompliantResourcesOptimized(ctx context.Context, request types.BatchComplianceRequest) (*types.BatchRemediationResult, error) {
	if m.ProcessNonCompliantResourcesOptimizedFunc != nil {
		return m.ProcessNonCompliantResourcesOptimizedFunc(ctx, request)
	}
	// Default behavior: return success
	return &types.BatchRemediationResult{
		TotalProcessed: len(request.NonCompliantResults),
		SuccessCount:   len(request.NonCompliantResults),
		FailureCount:   0,
		Results:        []types.RemediationResult{},
	}, nil
}

// GetNonCompliantResources implements ComplianceServiceInterface
func (m *MockComplianceService) GetNonCompliantResources(ctx context.Context, configRuleName string, region string) ([]types.NonCompliantResource, error) {
	if m.GetNonCompliantResourcesFunc != nil {
		return m.GetNonCompliantResourcesFunc(ctx, configRuleName, region)
	}
	// Default behavior: return empty list
	return []types.NonCompliantResource{}, nil
}

// ValidateResourceExistence implements ComplianceServiceInterface
func (m *MockComplianceService) ValidateResourceExistence(ctx context.Context, resources []types.NonCompliantResource) ([]types.NonCompliantResource, error) {
	if m.ValidateResourceExistenceFunc != nil {
		return m.ValidateResourceExistenceFunc(ctx, resources)
	}
	// Default behavior: return all resources as valid
	return resources, nil
}

// Ensure MockComplianceService implements the interface
var _ service.ComplianceServiceInterface = (*MockComplianceService)(nil)
