package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/zsoftly/logguardian/internal/handler"
	"github.com/zsoftly/logguardian/internal/mocks"
	"github.com/zsoftly/logguardian/internal/testutil"
	"github.com/zsoftly/logguardian/internal/types"
)

// BenchmarkHandleUnifiedRequest_ConfigEvent benchmarks the config-event handler
func BenchmarkHandleUnifiedRequest_ConfigEvent(b *testing.B) {
	ctx := context.Background()
	mockService := &mocks.MockComplianceService{}
	mockService.RemediateLogGroupFunc = func(ctx context.Context, compliance types.ComplianceResult) (*types.RemediationResult, error) {
		return &types.RemediationResult{
			LogGroupName:      compliance.LogGroupName,
			Region:            compliance.Region,
			Success:           true,
			EncryptionApplied: true,
		}, nil
	}
	h := handler.NewComplianceHandler(mockService)

	req, err := testutil.NewTestConfigEventRequest()
	if err != nil {
		b.Fatalf("Failed to create test request: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handleUnifiedRequest(ctx, h, req)
	}
}

// BenchmarkHandleUnifiedRequest_ConfigRuleEvaluation benchmarks the config-rule-evaluation handler
func BenchmarkHandleUnifiedRequest_ConfigRuleEvaluation(b *testing.B) {
	ctx := context.Background()
	mockService := &mocks.MockComplianceService{}

	// Create test resources
	resources := make([]types.NonCompliantResource, 10)
	for i := range resources {
		resources[i] = testutil.NewTestNonCompliantResource()
		resources[i].ResourceName = "/aws/lambda/test-function-" + string(rune(i))
	}

	mockService.GetNonCompliantResourcesFunc = func(ctx context.Context, configRuleName string, region string) ([]types.NonCompliantResource, error) {
		return resources, nil
	}
	mockService.ValidateResourceExistenceFunc = func(ctx context.Context, resources []types.NonCompliantResource) ([]types.NonCompliantResource, error) {
		return resources, nil
	}
	mockService.ProcessNonCompliantResourcesOptimizedFunc = func(ctx context.Context, request types.BatchComplianceRequest) (*types.BatchRemediationResult, error) {
		return testutil.NewTestBatchRemediationResult(len(request.NonCompliantResults), len(request.NonCompliantResults), 0), nil
	}

	h := handler.NewComplianceHandler(mockService)

	request := types.LambdaRequest{
		Type:           "config-rule-evaluation",
		ConfigRuleName: "test-encryption-rule",
		Region:         "ca-central-1",
		BatchSize:      10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handleUnifiedRequest(ctx, h, request)
	}
}

// BenchmarkHandleUnifiedRequest_ConfigRuleEvaluation_LargeBatch benchmarks with a large batch
func BenchmarkHandleUnifiedRequest_ConfigRuleEvaluation_LargeBatch(b *testing.B) {
	ctx := context.Background()
	mockService := &mocks.MockComplianceService{}

	// Create 100 test resources to simulate large batch
	resources := make([]types.NonCompliantResource, 100)
	for i := range resources {
		resources[i] = testutil.NewTestNonCompliantResource()
		resources[i].ResourceName = "/aws/lambda/test-function-" + string(rune(i))
	}

	mockService.GetNonCompliantResourcesFunc = func(ctx context.Context, configRuleName string, region string) ([]types.NonCompliantResource, error) {
		return resources, nil
	}
	mockService.ValidateResourceExistenceFunc = func(ctx context.Context, resources []types.NonCompliantResource) ([]types.NonCompliantResource, error) {
		return resources, nil
	}
	mockService.ProcessNonCompliantResourcesOptimizedFunc = func(ctx context.Context, request types.BatchComplianceRequest) (*types.BatchRemediationResult, error) {
		return testutil.NewTestBatchRemediationResult(len(request.NonCompliantResults), len(request.NonCompliantResults), 0), nil
	}

	h := handler.NewComplianceHandler(mockService)

	request := types.LambdaRequest{
		Type:           "config-rule-evaluation",
		ConfigRuleName: "test-encryption-rule",
		Region:         "ca-central-1",
		BatchSize:      10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handleUnifiedRequest(ctx, h, request)
	}
}

// BenchmarkHandleUnifiedRequest_ConfigEvent_JSONParsing benchmarks JSON parsing performance
func BenchmarkHandleUnifiedRequest_ConfigEvent_JSONParsing(b *testing.B) {
	ctx := context.Background()
	mockService := &mocks.MockComplianceService{}
	mockService.RemediateLogGroupFunc = func(ctx context.Context, compliance types.ComplianceResult) (*types.RemediationResult, error) {
		return &types.RemediationResult{Success: true}, nil
	}
	h := handler.NewComplianceHandler(mockService)

	// Create a realistic config event JSON
	configEvent := types.ConfigEvent{
		ConfigRuleName: "test-encryption-rule",
		AccountId:      "123456789012",
		ConfigRuleInvokingEvent: types.ConfigRuleInvokingEvent{
			ConfigurationItem: types.ConfigurationItem{
				ResourceType:            "AWS::Logs::LogGroup",
				ResourceName:            "/aws/lambda/test-function",
				ResourceId:              "/aws/lambda/test-function",
				AwsRegion:               "ca-central-1",
				ConfigurationItemStatus: "OK",
				Configuration: types.LogGroupConfiguration{
					LogGroupName: "/aws/lambda/test-function",
					KmsKeyId:     "",
				},
			},
		},
	}

	eventJSON, err := json.Marshal(configEvent)
	if err != nil {
		b.Fatalf("Failed to marshal config event: %v", err)
	}

	request := types.LambdaRequest{
		Type:        "config-event",
		ConfigEvent: eventJSON,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handleUnifiedRequest(ctx, h, request)
	}
}
