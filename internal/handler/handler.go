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
