package types

import (
	"testing"
)

func TestRuleClassifier_ClassifyRule(t *testing.T) {
	classifier := NewRuleClassifier()

	tests := []struct {
		name         string
		configRule   string
		expectedType RuleType
		description  string
	}{
		// Encryption rules
		{
			name:         "AWS standard encryption rule",
			configRule:   "cloudwatch-log-group-encrypted",
			expectedType: RuleTypeEncryption,
			description:  "Standard AWS Config encryption rule should be classified as encryption",
		},
		{
			name:         "LogGuardian encryption rule",
			configRule:   "logguardian-encryption-check",
			expectedType: RuleTypeEncryption,
			description:  "LogGuardian encryption rule should be classified as encryption",
		},
		{
			name:         "Custom encryption rule",
			configRule:   "my-org-log-group-encrypted",
			expectedType: RuleTypeEncryption,
			description:  "Custom rule with encrypted in name should be classified as encryption",
		},
		{
			name:         "KMS encryption rule",
			configRule:   "kms-encryption-check",
			expectedType: RuleTypeEncryption,
			description:  "Rule with encryption in name should be classified as encryption",
		},

		// Retention rules
		{
			name:         "AWS standard retention rule",
			configRule:   "cw-loggroup-retention-period-check",
			expectedType: RuleTypeRetention,
			description:  "Standard AWS Config retention rule should be classified as retention",
		},
		{
			name:         "LogGuardian retention rule",
			configRule:   "logguardian-retention-check",
			expectedType: RuleTypeRetention,
			description:  "LogGuardian retention rule should be classified as retention",
		},
		{
			name:         "Custom retention rule",
			configRule:   "audit-retention-period-check",
			expectedType: RuleTypeRetention,
			description:  "Custom rule with retention in name should be classified as retention",
		},

		// Unknown rules
		{
			name:         "Unrelated backup rule",
			configRule:   "s3-backup-policy-check",
			expectedType: RuleTypeUnknown,
			description:  "Unrelated backup rule should not be classified",
		},
		{
			name:         "Network security rule",
			configRule:   "vpc-security-group-check",
			expectedType: RuleTypeUnknown,
			description:  "Network security rule should not be classified",
		},

		// Edge cases
		{
			name:         "Empty rule name",
			configRule:   "",
			expectedType: RuleTypeUnknown,
			description:  "Empty rule name should return unknown",
		},
		{
			name:         "Case sensitivity test - uppercase",
			configRule:   "CLOUDWATCH-LOG-GROUP-ENCRYPTED",
			expectedType: RuleTypeEncryption,
			description:  "Rule classification should be case-insensitive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.ClassifyRule(tt.configRule)
			if result != tt.expectedType {
				t.Errorf("ClassifyRule(%q) = %v, expected %v. %s",
					tt.configRule, result, tt.expectedType, tt.description)
			}
		})
	}
}

func TestRuleClassifier_IsEncryptionRule(t *testing.T) {
	classifier := NewRuleClassifier()

	encryptionRules := []string{
		"cloudwatch-log-group-encrypted",
		"logguardian-encryption-check",
		"my-org-log-group-encrypted",
		"kms-encryption-rule",
	}

	nonEncryptionRules := []string{
		"cw-loggroup-retention-period-check",
		"s3-backup-policy-check",
		"vpc-security-group-check",
		"",
	}

	for _, rule := range encryptionRules {
		t.Run("encryption_"+rule, func(t *testing.T) {
			if !classifier.IsEncryptionRule(rule) {
				t.Errorf("IsEncryptionRule(%q) = false, expected true", rule)
			}
		})
	}

	for _, rule := range nonEncryptionRules {
		t.Run("non_encryption_"+rule, func(t *testing.T) {
			if classifier.IsEncryptionRule(rule) {
				t.Errorf("IsEncryptionRule(%q) = true, expected false", rule)
			}
		})
	}
}

func TestRuleClassifier_IsRetentionRule(t *testing.T) {
	classifier := NewRuleClassifier()

	retentionRules := []string{
		"cw-loggroup-retention-period-check",
		"logguardian-retention-check",
		"audit-retention-period-check",
		"log-retention-policy",
	}

	nonRetentionRules := []string{
		"cloudwatch-log-group-encrypted",
		"s3-backup-policy-check",
		"vpc-security-group-check",
		"",
	}

	for _, rule := range retentionRules {
		t.Run("retention_"+rule, func(t *testing.T) {
			if !classifier.IsRetentionRule(rule) {
				t.Errorf("IsRetentionRule(%q) = false, expected true", rule)
			}
		})
	}

	for _, rule := range nonRetentionRules {
		t.Run("non_retention_"+rule, func(t *testing.T) {
			if classifier.IsRetentionRule(rule) {
				t.Errorf("IsRetentionRule(%q) = true, expected false", rule)
			}
		})
	}
}

func TestRuleType_String(t *testing.T) {
	tests := []struct {
		ruleType    RuleType
		expectedStr string
	}{
		{RuleTypeEncryption, "encryption"},
		{RuleTypeRetention, "retention"},
		{RuleTypeUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedStr, func(t *testing.T) {
			if result := tt.ruleType.String(); result != tt.expectedStr {
				t.Errorf("RuleType.String() = %q, expected %q", result, tt.expectedStr)
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkRuleClassifier_ClassifyRule(b *testing.B) {
	classifier := NewRuleClassifier()
	testRules := []string{
		"cloudwatch-log-group-encrypted",
		"cw-loggroup-retention-period-check",
		"unknown-rule-name",
		"my-custom-log-group-encrypted",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, rule := range testRules {
			classifier.ClassifyRule(rule)
		}
	}
}
