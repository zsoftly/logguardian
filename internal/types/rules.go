package types

import (
	"strings"
)

// RuleType represents the type of AWS Config rule
type RuleType int

const (
	RuleTypeUnknown RuleType = iota
	RuleTypeEncryption
	RuleTypeRetention
)

// String returns the string representation of RuleType
func (rt RuleType) String() string {
	switch rt {
	case RuleTypeEncryption:
		return "encryption"
	case RuleTypeRetention:
		return "retention"
	default:
		return "unknown"
	}
}

// RuleClassifier provides simple rule classification logic
type RuleClassifier struct {
	// No complex patterns needed - just simple string matching
}

// NewRuleClassifier creates a new rule classifier
func NewRuleClassifier() *RuleClassifier {
	return &RuleClassifier{}
}

// ClassifyRule determines the type of Config rule using simple string matching
func (rc *RuleClassifier) ClassifyRule(configRuleName string) RuleType {
	if configRuleName == "" {
		return RuleTypeUnknown
	}

	// Normalize for case-insensitive matching
	normalizedName := strings.ToLower(configRuleName)

	// Simple substring matching - rules have descriptive names
	if strings.Contains(normalizedName, "encryption") || strings.Contains(normalizedName, "encrypted") {
		return RuleTypeEncryption
	}

	if strings.Contains(normalizedName, "retention") {
		return RuleTypeRetention
	}

	return RuleTypeUnknown
}

// IsEncryptionRule checks if the rule is an encryption-focused Config rule
func (rc *RuleClassifier) IsEncryptionRule(configRuleName string) bool {
	return rc.ClassifyRule(configRuleName) == RuleTypeEncryption
}

// IsRetentionRule checks if the rule is a retention-focused Config rule
func (rc *RuleClassifier) IsRetentionRule(configRuleName string) bool {
	return rc.ClassifyRule(configRuleName) == RuleTypeRetention
}
