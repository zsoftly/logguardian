package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/zsoftly/logguardian/internal/service"
	"github.com/zsoftly/logguardian/internal/types"
)

// ComplianceHandler handles AWS Config compliance events
type ComplianceHandler struct {
	complianceService service.ComplianceServiceInterface
	ruleClassifier    *types.RuleClassifier
}

// NewComplianceHandler creates a new compliance handler
func NewComplianceHandler(complianceService service.ComplianceServiceInterface) *ComplianceHandler {
	return &ComplianceHandler{
		complianceService: complianceService,
		ruleClassifier:    types.NewRuleClassifier(),
	}
}

// HandleConfigEvent handles AWS Config rule evaluation events
func (h *ComplianceHandler) HandleConfigEvent(ctx context.Context, event json.RawMessage) error {
	slog.Info("Received Config compliance event", "event_size", len(event))

	// Parse the event
	var configEvent types.ConfigEvent
	if err := json.Unmarshal(event, &configEvent); err != nil {
		slog.Error("Failed to parse Config event", "error", err)
		return fmt.Errorf("failed to parse Config event: %w", err)
	}

	slog.Info("Processing compliance event",
		"config_rule", configEvent.ConfigRuleName,
		"account_id", configEvent.AccountId,
		"resource_type", configEvent.ConfigRuleInvokingEvent.ConfigurationItem.ResourceType,
		"resource_name", configEvent.ConfigRuleInvokingEvent.ConfigurationItem.ResourceName,
		"region", configEvent.ConfigRuleInvokingEvent.ConfigurationItem.AwsRegion)

	// Skip if resource was deleted or is not a log group
	configItem := configEvent.ConfigRuleInvokingEvent.ConfigurationItem
	if configItem.ConfigurationItemStatus == "ResourceDeleted" {
		slog.Info("Skipping deleted resource", "resource_name", configItem.ResourceName)
		return nil
	}

	if configItem.ResourceType != "AWS::Logs::LogGroup" {
		slog.Warn("Unexpected resource type", "resource_type", configItem.ResourceType)
		return nil
	}

	// Check compliance status based on specific rule
	compliance := h.analyzeComplianceForRule(configEvent.ConfigRuleName, configItem)

	slog.Info("Rule-specific compliance analysis completed",
		"log_group", compliance.LogGroupName,
		"config_rule", configEvent.ConfigRuleName,
		"missing_encryption", compliance.MissingEncryption,
		"missing_retention", compliance.MissingRetention,
		"current_retention", compliance.CurrentRetention)

	// Apply remediation if needed for this specific rule's compliance requirement
	if compliance.MissingEncryption || compliance.MissingRetention {
		result, err := h.complianceService.RemediateLogGroup(ctx, compliance)
		if err != nil {
			slog.Error("Remediation failed",
				"log_group", compliance.LogGroupName,
				"error", err)
			return fmt.Errorf("remediation failed for %s: %w", compliance.LogGroupName, err)
		}

		slog.Info("Remediation completed",
			"log_group", result.LogGroupName,
			"encryption_applied", result.EncryptionApplied,
			"retention_applied", result.RetentionApplied,
			"success", result.Success)
	} else {
		slog.Info("Log group already compliant", "log_group", compliance.LogGroupName)
	}

	return nil
}

// HandleConfigRuleEvaluationRequest handles requests to process Config rule evaluation results
func (h *ComplianceHandler) HandleConfigRuleEvaluationRequest(ctx context.Context, configRuleName, region string, batchSize int) error {
	slog.Info("Processing Config rule evaluation request",
		"config_rule", configRuleName,
		"region", region,
		"batch_size", batchSize)

	// Step 1: Get non-compliant resources from Config API
	nonCompliantResources, err := h.complianceService.GetNonCompliantResources(ctx, configRuleName, region)
	if err != nil {
		slog.Error("Failed to retrieve non-compliant resources",
			"config_rule", configRuleName,
			"error", err)
		return fmt.Errorf("failed to retrieve non-compliant resources: %w", err)
	}

	if len(nonCompliantResources) == 0 {
		slog.Info("No non-compliant resources found",
			"config_rule", configRuleName,
			"region", region)
		return nil
	}

	slog.Info("Found non-compliant resources",
		"config_rule", configRuleName,
		"region", region,
		"count", len(nonCompliantResources))

	// Step 2: Validate resource existence before processing
	validResources, err := h.complianceService.ValidateResourceExistence(ctx, nonCompliantResources)
	if err != nil {
		slog.Error("Failed to validate resource existence",
			"config_rule", configRuleName,
			"error", err)
		return fmt.Errorf("failed to validate resource existence: %w", err)
	}

	if len(validResources) == 0 {
		slog.Info("No valid resources found after validation",
			"config_rule", configRuleName,
			"region", region)
		return nil
	}

	slog.Info("Validated resources for processing",
		"config_rule", configRuleName,
		"region", region,
		"valid_count", len(validResources),
		"filtered_count", len(nonCompliantResources)-len(validResources))

	// Step 3: Create batch request and process
	batchRequest := types.BatchComplianceRequest{
		ConfigRuleName:      configRuleName,
		NonCompliantResults: validResources,
		Region:              region,
		BatchSize:           batchSize,
	}

	// Step 4: Process the batch using optimized method with KMS validation caching
	result, err := h.complianceService.ProcessNonCompliantResourcesOptimized(ctx, batchRequest)
	if err != nil {
		slog.Error("Optimized batch processing failed",
			"config_rule", configRuleName,
			"error", err)
		return fmt.Errorf("optimized batch processing failed: %w", err)
	}

	slog.Info("Config rule evaluation processing completed",
		"config_rule", configRuleName,
		"region", region,
		"total_processed", result.TotalProcessed,
		"success_count", result.SuccessCount,
		"failure_count", result.FailureCount,
		"duration", result.ProcessingDuration,
		"rate_limit_hits", result.RateLimitHits)

	return nil
}

// analyzeComplianceForRule checks what remediation is needed based on the specific Config rule
func (h *ComplianceHandler) analyzeComplianceForRule(configRuleName string, configItem types.ConfigurationItem) types.ComplianceResult {
	config := configItem.Configuration

	result := types.ComplianceResult{
		LogGroupName:     config.LogGroupName,
		Region:           configItem.AwsRegion,
		AccountId:        configItem.AwsAccountId,
		CurrentRetention: config.RetentionInDays,
		CurrentKmsKeyId:  config.KmsKeyId,
	}

	// Each Config rule evaluates ONLY its specific compliance requirement
	// This ensures each rule evaluates ALL resources for its requirement independently
	ruleType := h.ruleClassifier.ClassifyRule(configRuleName)

	switch ruleType {
	case types.RuleTypeEncryption:
		// Encryption rule: ONLY evaluate encryption compliance
		result.MissingEncryption = config.KmsKeyId == ""
		result.MissingRetention = false // Not this rule's concern

		slog.Info("Encryption rule evaluation",
			"log_group", config.LogGroupName,
			"has_encryption", config.KmsKeyId != "",
			"kms_key_id", config.KmsKeyId,
			"rule_type", ruleType.String(),
			"audit_action", "encryption_compliance_check")

	case types.RuleTypeRetention:
		// Retention rule: ONLY evaluate retention compliance
		result.MissingRetention = config.RetentionInDays == nil
		result.MissingEncryption = false // Not this rule's concern

		slog.Info("Retention rule evaluation",
			"log_group", config.LogGroupName,
			"has_retention", config.RetentionInDays != nil,
			"retention_days", config.RetentionInDays,
			"rule_type", ruleType.String(),
			"audit_action", "retention_compliance_check")

	default:
		// Unknown rule - log and skip
		slog.Warn("Unsupported Config rule - no compliance evaluation performed",
			"config_rule", configRuleName,
			"log_group", config.LogGroupName,
			"rule_type", "unknown",
			"audit_action", "unsupported_rule_skip")
		result.MissingEncryption = false
		result.MissingRetention = false
	}

	return result
}
