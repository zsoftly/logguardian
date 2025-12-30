package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zsoftly/logguardian/internal/types"
)

// AssertComplianceResult validates that a ComplianceResult matches expected values
func AssertComplianceResult(t *testing.T, expected, actual types.ComplianceResult) {
	t.Helper()
	assert.Equal(t, expected.LogGroupName, actual.LogGroupName, "LogGroupName should match")
	assert.Equal(t, expected.Region, actual.Region, "Region should match")
	assert.Equal(t, expected.AccountId, actual.AccountId, "AccountId should match")
	assert.Equal(t, expected.MissingEncryption, actual.MissingEncryption, "MissingEncryption should match")
	assert.Equal(t, expected.MissingRetention, actual.MissingRetention, "MissingRetention should match")
}

// AssertRemediationSuccess validates that a RemediationResult indicates success
func AssertRemediationSuccess(t *testing.T, result *types.RemediationResult) {
	t.Helper()
	assert.NotNil(t, result, "RemediationResult should not be nil")
	assert.True(t, result.Success, "Remediation should be successful")
	assert.Nil(t, result.Error, "Remediation should have no error")
}

// AssertRemediationFailure validates that a RemediationResult indicates failure
func AssertRemediationFailure(t *testing.T, result *types.RemediationResult, expectedErrorContains string) {
	t.Helper()
	assert.NotNil(t, result, "RemediationResult should not be nil")
	assert.False(t, result.Success, "Remediation should be unsuccessful")
	assert.NotNil(t, result.Error, "Remediation should have an error")
	if expectedErrorContains != "" {
		assert.Contains(t, result.Error.Error(), expectedErrorContains, "Error message should contain expected text")
	}
}

// AssertBatchResult validates that a BatchRemediationResult matches expected values
func AssertBatchResult(t *testing.T, result *types.BatchRemediationResult, expectedTotal, expectedSuccess, expectedFailure int) {
	t.Helper()
	assert.NotNil(t, result, "BatchRemediationResult should not be nil")
	assert.Equal(t, expectedTotal, result.TotalProcessed, "TotalProcessed should match")
	assert.Equal(t, expectedSuccess, result.SuccessCount, "SuccessCount should match")
	assert.Equal(t, expectedFailure, result.FailureCount, "FailureCount should match")
}
