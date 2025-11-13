package testutil

import (
	"encoding/json"
	"time"

	"github.com/zsoftly/logguardian/internal/types"
)

// NewTestLambdaRequest creates a test LambdaRequest for unit testing
func NewTestLambdaRequest() types.LambdaRequest {
	return types.LambdaRequest{
		Type:           "config-rule-evaluation",
		ConfigRuleName: "test-config-rule",
		Region:         "ca-central-1",
		BatchSize:      10,
	}
}

// NewTestConfigEventRequest creates a test LambdaRequest with ConfigEvent payload
func NewTestConfigEventRequest() (types.LambdaRequest, error) {
	configEvent := types.ConfigEvent{
		ConfigRuleName: "test-encryption-rule",
		AccountId:      "123456789012",
		ConfigRuleArn:  "arn:aws:config:ca-central-1:123456789012:config-rule/test-rule",
		ConfigRuleInvokingEvent: types.ConfigRuleInvokingEvent{
			ConfigurationItem: types.ConfigurationItem{
				ResourceType:            "AWS::Logs::LogGroup",
				ResourceName:            "/aws/lambda/test-function",
				ResourceId:              "/aws/lambda/test-function",
				AwsRegion:               "ca-central-1",
				AwsAccountId:            "123456789012",
				ConfigurationItemStatus: "OK",
				Configuration: types.LogGroupConfiguration{
					LogGroupName:    "/aws/lambda/test-function",
					RetentionInDays: nil,
					KmsKeyId:        "",
				},
			},
			MessageType: "ConfigurationItemChangeNotification",
		},
	}

	eventJSON, err := json.Marshal(configEvent)
	if err != nil {
		return types.LambdaRequest{}, err
	}

	return types.LambdaRequest{
		Type:        "config-event",
		ConfigEvent: eventJSON,
	}, nil
}

// NewTestComplianceResult creates a test ComplianceResult
func NewTestComplianceResult() types.ComplianceResult {
	retention := int32(30)
	return types.ComplianceResult{
		LogGroupName:      "/aws/lambda/test-function",
		Region:            "ca-central-1",
		AccountId:         "123456789012",
		MissingEncryption: true,
		MissingRetention:  true,
		CurrentRetention:  &retention,
		CurrentKmsKeyId:   "",
	}
}

// NewTestNonCompliantResource creates a test NonCompliantResource
func NewTestNonCompliantResource() types.NonCompliantResource {
	return types.NonCompliantResource{
		ResourceId:     "/aws/lambda/test-function",
		ResourceType:   "AWS::Logs::LogGroup",
		ResourceName:   "/aws/lambda/test-function",
		Region:         "ca-central-1",
		AccountId:      "123456789012",
		ComplianceType: "NON_COMPLIANT",
		Annotation:     "Log group is not encrypted",
		LastEvaluated:  time.Now(),
	}
}

// NewTestRemediationResult creates a test RemediationResult
func NewTestRemediationResult() *types.RemediationResult {
	return &types.RemediationResult{
		LogGroupName:      "/aws/lambda/test-function",
		Region:            "ca-central-1",
		Success:           true,
		EncryptionApplied: true,
		RetentionApplied:  true,
	}
}

// NewTestBatchComplianceRequest creates a test BatchComplianceRequest
func NewTestBatchComplianceRequest(resources []types.NonCompliantResource) types.BatchComplianceRequest {
	return types.BatchComplianceRequest{
		ConfigRuleName:      "test-config-rule",
		NonCompliantResults: resources,
		Region:              "ca-central-1",
		BatchSize:           10,
	}
}

// NewTestBatchRemediationResult creates a test BatchRemediationResult
func NewTestBatchRemediationResult(total, success, failure int) *types.BatchRemediationResult {
	return &types.BatchRemediationResult{
		TotalProcessed:     total,
		SuccessCount:       success,
		FailureCount:       failure,
		Results:            []types.RemediationResult{},
		ProcessingDuration: 100 * time.Millisecond,
		RateLimitHits:      0,
	}
}
