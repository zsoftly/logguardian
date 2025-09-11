package container

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/zsoftly/logguardian/internal/handler"
	"github.com/zsoftly/logguardian/internal/service"
	"github.com/zsoftly/logguardian/internal/types"
)

type CommandProcessor struct {
	handler      *handler.ComplianceHandler
	service      service.ComplianceServiceInterface
	options      ProcessorOptions
	executionLog []ExecutionLogEntry
}

type ProcessorOptions struct {
	DryRun       bool
	ExecutionID  string
	OutputFormat string
}

type CommandRequest struct {
	Type           string
	ConfigRuleName string
	Region         string
	BatchSize      int
}

type ExecutionResult struct {
	ExecutionID    string              `json:"execution_id"`
	Status         string              `json:"status"`
	Mode           string              `json:"mode"`
	ConfigRuleName string              `json:"config_rule_name"`
	Region         string              `json:"region"`
	TotalProcessed int                 `json:"total_processed"`
	SuccessCount   int                 `json:"success_count"`
	FailureCount   int                 `json:"failure_count"`
	Duration       string              `json:"duration"`
	Timestamp      time.Time           `json:"timestamp"`
	Resources      []ResourceResult    `json:"resources,omitempty"`
	DryRunSummary  *DryRunSummary      `json:"dry_run_summary,omitempty"`
	Error          string              `json:"error,omitempty"`
	ExecutionLog   []ExecutionLogEntry `json:"execution_log,omitempty"`
}

type ResourceResult struct {
	ResourceID        string    `json:"resource_id"`
	ResourceName      string    `json:"resource_name"`
	Status            string    `json:"status"`
	EncryptionApplied bool      `json:"encryption_applied"`
	RetentionApplied  bool      `json:"retention_applied"`
	Error             string    `json:"error,omitempty"`
	Timestamp         time.Time `json:"timestamp"`
}

type DryRunSummary struct {
	WouldApplyEncryption int `json:"would_apply_encryption"`
	WouldApplyRetention  int `json:"would_apply_retention"`
	AlreadyCompliant     int `json:"already_compliant"`
	TotalResources       int `json:"total_resources"`
}

type ExecutionLogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Details   any       `json:"details,omitempty"`
}

func NewCommandProcessor(awsCfg aws.Config, options ProcessorOptions) *CommandProcessor {
	var complianceService service.ComplianceServiceInterface

	if options.DryRun {
		// Create a dry-run wrapper for the compliance service
		realService := service.NewComplianceService(awsCfg)
		complianceService = NewDryRunComplianceService(realService)
	} else {
		complianceService = service.NewComplianceService(awsCfg)
	}

	h := handler.NewComplianceHandler(complianceService)

	return &CommandProcessor{
		handler:      h,
		service:      complianceService,
		options:      options,
		executionLog: []ExecutionLogEntry{},
	}
}

func (p *CommandProcessor) Execute(ctx context.Context, request CommandRequest) (*ExecutionResult, error) {
	startTime := time.Now()

	p.logEntry("INFO", "Starting command execution", map[string]any{
		"type":        request.Type,
		"config_rule": request.ConfigRuleName,
		"region":      request.Region,
		"batch_size":  request.BatchSize,
		"dry_run":     p.options.DryRun,
	})

	result := &ExecutionResult{
		ExecutionID:    p.options.ExecutionID,
		Status:         "running",
		Mode:           p.getMode(),
		ConfigRuleName: request.ConfigRuleName,
		Region:         request.Region,
		Timestamp:      startTime,
		Resources:      []ResourceResult{},
	}

	switch request.Type {
	case "config-rule-evaluation":
		err := p.processConfigRuleEvaluation(ctx, request, result)
		if err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			p.logEntry("ERROR", "Execution failed", map[string]any{"error": err.Error()})
			return result, err
		}
	default:
		err := fmt.Errorf("unsupported request type: %s", request.Type)
		result.Status = "failed"
		result.Error = err.Error()
		return result, err
	}

	result.Status = "completed"
	result.Duration = time.Since(startTime).String()
	result.ExecutionLog = p.executionLog

	p.logEntry("INFO", "Command execution completed", map[string]any{
		"duration":        result.Duration,
		"total_processed": result.TotalProcessed,
		"success_count":   result.SuccessCount,
		"failure_count":   result.FailureCount,
	})

	return result, nil
}

func (p *CommandProcessor) processConfigRuleEvaluation(ctx context.Context, request CommandRequest, result *ExecutionResult) error {
	// Step 1: Get non-compliant resources
	p.logEntry("INFO", "Retrieving non-compliant resources", map[string]any{
		"config_rule": request.ConfigRuleName,
		"region":      request.Region,
	})

	nonCompliantResources, err := p.service.GetNonCompliantResources(ctx, request.ConfigRuleName, request.Region)
	if err != nil {
		return fmt.Errorf("failed to retrieve non-compliant resources: %w", err)
	}

	if len(nonCompliantResources) == 0 {
		p.logEntry("INFO", "No non-compliant resources found", nil)
		result.TotalProcessed = 0
		return nil
	}

	p.logEntry("INFO", "Found non-compliant resources", map[string]any{
		"count": len(nonCompliantResources),
	})

	// Step 2: Validate resource existence
	validResources, err := p.service.ValidateResourceExistence(ctx, nonCompliantResources)
	if err != nil {
		return fmt.Errorf("failed to validate resource existence: %w", err)
	}

	if len(validResources) == 0 {
		p.logEntry("INFO", "No valid resources found after validation", nil)
		result.TotalProcessed = 0
		return nil
	}

	p.logEntry("INFO", "Validated resources for processing", map[string]any{
		"valid_count":    len(validResources),
		"filtered_count": len(nonCompliantResources) - len(validResources),
	})

	// Step 3: Process resources
	if p.options.DryRun {
		return p.processDryRun(ctx, request, validResources, result)
	} else {
		return p.processResources(ctx, request, validResources, result)
	}
}

func (p *CommandProcessor) processResources(ctx context.Context, request CommandRequest, resources []types.NonCompliantResource, result *ExecutionResult) error {
	batchRequest := types.BatchComplianceRequest{
		ConfigRuleName:      request.ConfigRuleName,
		NonCompliantResults: resources,
		Region:              request.Region,
		BatchSize:           request.BatchSize,
	}

	batchResult, err := p.service.ProcessNonCompliantResourcesOptimized(ctx, batchRequest)
	if err != nil {
		return fmt.Errorf("batch processing failed: %w", err)
	}

	result.TotalProcessed = batchResult.TotalProcessed
	result.SuccessCount = batchResult.SuccessCount
	result.FailureCount = batchResult.FailureCount

	// Convert batch results to resource results
	for _, r := range batchResult.Results {
		resourceResult := ResourceResult{
			ResourceID:        r.LogGroupName,
			ResourceName:      r.LogGroupName,
			Status:            getResourceStatus(r),
			EncryptionApplied: r.EncryptionApplied,
			RetentionApplied:  r.RetentionApplied,
			Timestamp:         time.Now(),
		}
		if r.Error != nil {
			resourceResult.Error = r.Error.Error()
		}
		result.Resources = append(result.Resources, resourceResult)
	}

	return nil
}

func (p *CommandProcessor) processDryRun(ctx context.Context, request CommandRequest, resources []types.NonCompliantResource, result *ExecutionResult) error {
	dryRunSummary := &DryRunSummary{
		TotalResources: len(resources),
	}

	p.logEntry("INFO", "Running in dry-run mode", map[string]any{
		"total_resources": len(resources),
	})

	// Analyze each resource to determine what would be done
	ruleClassifier := types.NewRuleClassifier()
	ruleType := ruleClassifier.ClassifyRule(request.ConfigRuleName)

	for _, resource := range resources {
		// Get current state
		compliance, err := p.analyzeResourceCompliance(ctx, resource, ruleType)
		if err != nil {
			p.logEntry("WARN", "Failed to analyze resource", map[string]any{
				"resource": resource.ResourceName,
				"error":    err.Error(),
			})
			result.FailureCount++
			continue
		}

		resourceResult := ResourceResult{
			ResourceID:   resource.ResourceId,
			ResourceName: resource.ResourceName,
			Status:       "dry-run",
			Timestamp:    time.Now(),
		}

		if compliance.MissingEncryption {
			dryRunSummary.WouldApplyEncryption++
			resourceResult.EncryptionApplied = true
			p.logEntry("INFO", "Would apply encryption", map[string]any{
				"resource": resource.ResourceName,
			})
		}

		if compliance.MissingRetention {
			dryRunSummary.WouldApplyRetention++
			resourceResult.RetentionApplied = true
			p.logEntry("INFO", "Would apply retention", map[string]any{
				"resource": resource.ResourceName,
			})
		}

		if !compliance.MissingEncryption && !compliance.MissingRetention {
			dryRunSummary.AlreadyCompliant++
			resourceResult.Status = "compliant"
			p.logEntry("INFO", "Resource already compliant", map[string]any{
				"resource": resource.ResourceName,
			})
		}

		result.Resources = append(result.Resources, resourceResult)
		result.SuccessCount++
	}

	result.TotalProcessed = len(resources)
	result.DryRunSummary = dryRunSummary

	return nil
}

func (p *CommandProcessor) analyzeResourceCompliance(ctx context.Context, resource types.NonCompliantResource, ruleType types.RuleType) (types.ComplianceResult, error) {
	// Context will be used for AWS API calls when fetching actual log group configuration
	_ = ctx // Currently unused but kept for future AWS API integration
	// For now, we'll return based on the rule type
	result := types.ComplianceResult{
		LogGroupName: resource.ResourceName,
		Region:       resource.Region,
		AccountId:    resource.AccountId,
	}

	switch ruleType {
	case types.RuleTypeEncryption:
		result.MissingEncryption = true
		result.MissingRetention = false
	case types.RuleTypeRetention:
		result.MissingRetention = true
		result.MissingEncryption = false
	default:
		// Unknown rule type
		result.MissingEncryption = false
		result.MissingRetention = false
	}

	return result, nil
}

func (p *CommandProcessor) logEntry(level, message string, details any) {
	entry := ExecutionLogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Details:   details,
	}
	p.executionLog = append(p.executionLog, entry)

	// Also log to slog
	switch level {
	case "ERROR":
		slog.Error(message, "details", details)
	case "WARN":
		slog.Warn(message, "details", details)
	case "INFO":
		slog.Info(message, "details", details)
	case "DEBUG":
		slog.Debug(message, "details", details)
	}
}

func (p *CommandProcessor) getMode() string {
	if p.options.DryRun {
		return "dry-run"
	}
	return "apply"
}

func getResourceStatus(result types.RemediationResult) string {
	if result.Success {
		return "success"
	}
	return "failed"
}
