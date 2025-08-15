package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/zsoftly/logguardian/internal/types"
)

func TestCustomConfigRuleIntegration(t *testing.T) {
	// Set up test environment
	os.Setenv("DEFAULT_RETENTION_DAYS", "30")

	// Mock config client that captures evaluation submissions
	var capturedEvaluations []mockEvaluation
	mockClient := &MockConfigServiceClient{
		PutEvaluationsFunc: func(ctx context.Context, params *configservice.PutEvaluationsInput, optFns ...func(*configservice.Options)) (*configservice.PutEvaluationsOutput, error) {
			for _, eval := range params.Evaluations {
				capturedEvaluations = append(capturedEvaluations, mockEvaluation{
					ResourceId:     *eval.ComplianceResourceId,
					ResourceType:   *eval.ComplianceResourceType,
					ComplianceType: string(eval.ComplianceType),
					Annotation:     *eval.Annotation,
				})
			}
			return &configservice.PutEvaluationsOutput{}, nil
		},
	}

	handler := &CustomConfigRuleHandler{
		configClient: mockClient,
	}

	testCases := []struct {
		name                   string
		testDataFile           string
		expectedComplianceType string
		expectedAnnotation     string
	}{
		{
			name:                   "null retention should be NON_COMPLIANT",
			testDataFile:           "custom-retention-rule-null-retention.json",
			expectedComplianceType: "NON_COMPLIANT",
			expectedAnnotation:     "No retention policy set (infinite retention). Minimum required: 30 days",
		},
		{
			name:                   "compliant retention should be COMPLIANT",
			testDataFile:           "custom-retention-rule-compliant.json",
			expectedComplianceType: "COMPLIANT",
			expectedAnnotation:     "Retention period (90 days) meets minimum requirement (30 days)",
		},
		{
			name:                   "low retention should be NON_COMPLIANT",
			testDataFile:           "custom-retention-rule-low-retention.json",
			expectedComplianceType: "NON_COMPLIANT",
			expectedAnnotation:     "Retention period (7 days) below minimum requirement (30 days)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset captured evaluations
			capturedEvaluations = nil

			// Load test data
			testData, err := os.ReadFile("../../testdata/" + tc.testDataFile)
			if err != nil {
				t.Fatalf("Failed to read test data file %s: %v", tc.testDataFile, err)
			}

			var event types.ConfigEvent
			if err := json.Unmarshal(testData, &event); err != nil {
				t.Fatalf("Failed to unmarshal test data: %v", err)
			}

			// Execute handler
			err = handler.HandleConfigRuleEvent(context.Background(), event)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			// Verify evaluation was submitted
			if len(capturedEvaluations) != 1 {
				t.Fatalf("Expected 1 evaluation, got %d", len(capturedEvaluations))
			}

			eval := capturedEvaluations[0]
			if eval.ComplianceType != tc.expectedComplianceType {
				t.Errorf("Expected compliance type %s, got %s", tc.expectedComplianceType, eval.ComplianceType)
			}

			if eval.Annotation != tc.expectedAnnotation {
				t.Errorf("Expected annotation %q, got %q", tc.expectedAnnotation, eval.Annotation)
			}

			t.Logf("Test passed: %s -> %s: %s", tc.testDataFile, eval.ComplianceType, eval.Annotation)
		})
	}
}

type mockEvaluation struct {
	ResourceId     string
	ResourceType   string
	ComplianceType string
	Annotation     string
}
