package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	configtypes "github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/zsoftly/logguardian/internal/types"
)

// Constants for retention policy validation
const (
	DefaultMinRetentionDays = 30   // Default minimum retention period
	MinRetentionDays        = 1    // AWS CloudWatch Logs minimum
	MaxRetentionDays        = 3653 // AWS CloudWatch Logs maximum (10 years)
)

// ConfigServiceClientInterface defines the interface for Config service operations
type ConfigServiceClientInterface interface {
	PutEvaluations(ctx context.Context, params *configservice.PutEvaluationsInput, optFns ...func(*configservice.Options)) (*configservice.PutEvaluationsOutput, error)
}

// CustomConfigRuleHandler handles custom Config rule evaluations
type CustomConfigRuleHandler struct {
	configClient ConfigServiceClientInterface
}

// NewCustomConfigRuleHandler creates a new custom Config rule handler
func NewCustomConfigRuleHandler(cfg aws.Config) (*CustomConfigRuleHandler, error) {
	if cfg.Region == "" {
		return nil, fmt.Errorf("AWS region is required")
	}
	return &CustomConfigRuleHandler{
		configClient: configservice.NewFromConfig(cfg),
	}, nil
}

// RetentionRuleParameters represents the input parameters for the retention rule
type RetentionRuleParameters struct {
	MinRetentionTime int32 `json:"MinRetentionTime"`
}

func main() {
	// Set up structured logging with JSON output for Lambda
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		slog.Error("Failed to load AWS config", "error", err)
		os.Exit(1)
	}

	// Create handler
	handler, err := NewCustomConfigRuleHandler(cfg)
	if err != nil {
		slog.Error("Failed to create handler", "error", err)
		os.Exit(1)
	}

	// Start Lambda
	lambda.Start(handler.HandleConfigRuleEvent)
}

// HandleConfigRuleEvent handles AWS Config rule evaluation events
func (h *CustomConfigRuleHandler) HandleConfigRuleEvent(ctx context.Context, event types.ConfigEvent) error {
	// Validate required event fields
	if err := h.validateEvent(event); err != nil {
		slog.Error("Invalid event received", "error", err)
		return fmt.Errorf("invalid event: %w", err)
	}

	slog.Info("Received custom Config rule evaluation event",
		"config_rule", event.ConfigRuleName,
		"account_id", event.AccountId,
		"resource_type", event.ConfigRuleInvokingEvent.ConfigurationItem.ResourceType,
		"resource_name", event.ConfigRuleInvokingEvent.ConfigurationItem.ResourceName,
		"region", event.ConfigRuleInvokingEvent.ConfigurationItem.AwsRegion)

	// Parse rule parameters
	var ruleParams RetentionRuleParameters
	if err := h.parseRuleParameters(event.RuleParameters, &ruleParams); err != nil {
		slog.Error("Failed to parse rule parameters", "error", err)
		return fmt.Errorf("failed to parse rule parameters: %w", err)
	}

	// Get default retention days from environment if not in parameters
	if ruleParams.MinRetentionTime == 0 {
		if defaultDays := os.Getenv("DEFAULT_RETENTION_DAYS"); defaultDays != "" {
			if days, err := strconv.ParseInt(defaultDays, 10, 32); err == nil && days >= MinRetentionDays && days <= MaxRetentionDays {
				ruleParams.MinRetentionTime = int32(days)
			} else {
				slog.Warn("Invalid DEFAULT_RETENTION_DAYS environment variable, using default",
					"env_value", defaultDays, "default", DefaultMinRetentionDays)
				ruleParams.MinRetentionTime = DefaultMinRetentionDays
			}
		} else {
			ruleParams.MinRetentionTime = DefaultMinRetentionDays
		}
	}

	slog.Info("Using retention rule parameters",
		"min_retention_days", ruleParams.MinRetentionTime,
		"config_rule", event.ConfigRuleName)

	// Skip if resource was deleted or is not a log group
	configItem := event.ConfigRuleInvokingEvent.ConfigurationItem
	if configItem.ConfigurationItemStatus == "ResourceDeleted" {
		slog.Info("Skipping deleted resource", "resource_name", configItem.ResourceName)
		return nil
	}

	if configItem.ResourceType != "AWS::Logs::LogGroup" {
		slog.Warn("Unexpected resource type", "resource_type", configItem.ResourceType)
		return nil
	}

	// Evaluate retention compliance
	complianceType, annotation := h.evaluateRetentionCompliance(configItem, ruleParams.MinRetentionTime)

	slog.Info("Retention compliance evaluation completed",
		"log_group", configItem.Configuration.LogGroupName,
		"current_retention", configItem.Configuration.RetentionInDays,
		"min_required", ruleParams.MinRetentionTime,
		"compliance_result", complianceType,
		"annotation", annotation,
		"audit_action", "custom_retention_rule_evaluation")

	// Submit evaluation result to AWS Config
	if err := h.submitEvaluationResult(ctx, event, configItem, complianceType, annotation); err != nil {
		slog.Error("Failed to submit evaluation result",
			"log_group", configItem.Configuration.LogGroupName,
			"error", err)
		return fmt.Errorf("failed to submit evaluation result: %w", err)
	}

	slog.Info("Successfully submitted evaluation result",
		"log_group", configItem.Configuration.LogGroupName,
		"compliance_result", complianceType)

	return nil
}

// validateEvent validates required fields in the Config event
func (h *CustomConfigRuleHandler) validateEvent(event types.ConfigEvent) error {
	if event.ResultToken == "" {
		return fmt.Errorf("ResultToken is required")
	}
	if event.ConfigRuleName == "" {
		return fmt.Errorf("ConfigRuleName is required")
	}
	if event.ConfigRuleInvokingEvent.ConfigurationItem.ResourceType == "" {
		return fmt.Errorf("ResourceType is required")
	}
	if event.ConfigRuleInvokingEvent.ConfigurationItem.ResourceName == "" {
		return fmt.Errorf("ResourceName is required")
	}
	return nil
}

// parseRuleParameters parses the rule parameters from the Config event
func (h *CustomConfigRuleHandler) parseRuleParameters(params map[string]string, target *RetentionRuleParameters) error {
	if len(params) == 0 {
		return nil // No parameters provided, will use defaults
	}

	// Parse MinRetentionTime parameter
	if minRetentionStr, exists := params["MinRetentionTime"]; exists {
		if minRetention, err := strconv.ParseInt(minRetentionStr, 10, 32); err != nil {
			return fmt.Errorf("failed to parse MinRetentionTime parameter: %w", err)
		} else if minRetention < MinRetentionDays || minRetention > MaxRetentionDays {
			return fmt.Errorf("MinRetentionTime must be between %d and %d days, got %d", MinRetentionDays, MaxRetentionDays, minRetention)
		} else {
			target.MinRetentionTime = int32(minRetention)
		}
	}

	return nil
}

// evaluateRetentionCompliance evaluates retention compliance for a log group
func (h *CustomConfigRuleHandler) evaluateRetentionCompliance(configItem types.ConfigurationItem, minRetentionDays int32) (string, string) {
	retention := configItem.Configuration.RetentionInDays

	// Core logic: null retention (infinite retention) is NON_COMPLIANT
	if retention == nil {
		return "NON_COMPLIANT", fmt.Sprintf("No retention policy set (infinite retention). Minimum required: %d days", minRetentionDays)
	}

	// Check if retention meets minimum requirement
	if *retention < minRetentionDays {
		return "NON_COMPLIANT", fmt.Sprintf("Retention period (%d days) below minimum requirement (%d days)", *retention, minRetentionDays)
	}

	return "COMPLIANT", fmt.Sprintf("Retention period (%d days) meets minimum requirement (%d days)", *retention, minRetentionDays)
}

// submitEvaluationResult submits the evaluation result to AWS Config
func (h *CustomConfigRuleHandler) submitEvaluationResult(ctx context.Context, event types.ConfigEvent, configItem types.ConfigurationItem, complianceType, annotation string) error {
	// Add timeout for AWS API call
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	evaluationResult := configtypes.Evaluation{
		ComplianceResourceId:   aws.String(configItem.ResourceId),
		ComplianceResourceType: aws.String(configItem.ResourceType),
		ComplianceType:         configtypes.ComplianceType(complianceType),
		OrderingTimestamp:      aws.Time(time.Now()),
		Annotation:             aws.String(annotation),
	}

	input := &configservice.PutEvaluationsInput{
		Evaluations: []configtypes.Evaluation{evaluationResult},
		ResultToken: aws.String(event.ResultToken),
	}

	_, err := h.configClient.PutEvaluations(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to submit evaluation result for resource %s (rule: %s): %w",
			configItem.ResourceName, event.ConfigRuleName, err)
	}

	return nil
}
