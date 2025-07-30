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
}

// NewComplianceHandler creates a new compliance handler
func NewComplianceHandler(complianceService service.ComplianceServiceInterface) *ComplianceHandler {
	return &ComplianceHandler{
		complianceService: complianceService,
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

	// Check compliance status
	compliance := h.analyzeCompliance(configItem)

	slog.Info("Compliance analysis completed",
		"log_group", compliance.LogGroupName,
		"missing_encryption", compliance.MissingEncryption,
		"missing_retention", compliance.MissingRetention,
		"current_retention", compliance.CurrentRetention)

	// Apply remediation if needed
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

// HandleBatchConfigEvaluation handles batch processing of Config rule evaluation results
func (h *ComplianceHandler) HandleBatchConfigEvaluation(ctx context.Context, event json.RawMessage) error {
	slog.Info("Received batch Config evaluation event", "event_size", len(event))

	// Parse the batch event
	var batchRequest types.BatchComplianceRequest
	if err := json.Unmarshal(event, &batchRequest); err != nil {
		slog.Error("Failed to parse batch Config event", "error", err)
		return fmt.Errorf("failed to parse batch Config event: %w", err)
	}

	slog.Info("Processing batch compliance event",
		"config_rule", batchRequest.ConfigRuleName,
		"region", batchRequest.Region,
		"resource_count", len(batchRequest.NonCompliantResults),
		"batch_size", batchRequest.BatchSize)

	// Process the batch of non-compliant resources
	result, err := h.complianceService.ProcessNonCompliantResources(ctx, batchRequest)
	if err != nil {
		slog.Error("Batch remediation failed",
			"config_rule", batchRequest.ConfigRuleName,
			"error", err)
		return fmt.Errorf("batch remediation failed for rule %s: %w", batchRequest.ConfigRuleName, err)
	}

	slog.Info("Batch remediation completed",
		"config_rule", batchRequest.ConfigRuleName,
		"total_processed", result.TotalProcessed,
		"success_count", result.SuccessCount,
		"failure_count", result.FailureCount,
		"duration", result.ProcessingDuration,
		"rate_limit_hits", result.RateLimitHits)

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

	// Step 4: Process the batch
	result, err := h.complianceService.ProcessNonCompliantResources(ctx, batchRequest)
	if err != nil {
		slog.Error("Batch processing failed",
			"config_rule", configRuleName,
			"error", err)
		return fmt.Errorf("batch processing failed: %w", err)
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

// analyzeCompliance checks what remediation is needed for a log group
func (h *ComplianceHandler) analyzeCompliance(configItem types.ConfigurationItem) types.ComplianceResult {
	config := configItem.Configuration

	return types.ComplianceResult{
		LogGroupName:      config.LogGroupName,
		Region:            configItem.AwsRegion,
		AccountId:         configItem.AwsAccountId,
		MissingEncryption: config.KmsKeyId == "",
		MissingRetention:  config.RetentionInDays == nil,
		CurrentRetention:  config.RetentionInDays,
		CurrentKmsKeyId:   config.KmsKeyId,
	}
}
