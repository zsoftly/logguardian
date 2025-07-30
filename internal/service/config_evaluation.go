package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	logguardiantypes "github.com/zsoftly/logguardian/internal/types"
)

// ConfigEvaluationService handles AWS Config rule evaluation processing
type ConfigEvaluationService struct {
	configClient ConfigServiceClientInterface
	config       ServiceConfig
}

// NewConfigEvaluationService creates a new Config evaluation service
func NewConfigEvaluationService(cfg aws.Config) *ConfigEvaluationService {
	// Load configuration from environment variables
	config := ServiceConfig{
		DefaultKMSKeyAlias:   getEnvOrDefault("KMS_KEY_ALIAS", "alias/cloudwatch-logs-compliance"),
		DefaultRetentionDays: getEnvAsInt32OrDefault("DEFAULT_RETENTION_DAYS", 365),
		DryRun:               getEnvAsBoolOrDefault("DRY_RUN", false),
		BatchLimit:           getEnvAsInt32OrDefault("BATCH_LIMIT", 100),
	}

	return &ConfigEvaluationService{
		configClient: configservice.NewFromConfig(cfg),
		config:       config,
	}
}

// GetNonCompliantResources retrieves non-compliant log groups from Config API
func (s *ConfigEvaluationService) GetNonCompliantResources(ctx context.Context, configRuleName string, region string) ([]logguardiantypes.NonCompliantResource, error) {
	slog.Info("Retrieving non-compliant resources from Config",
		"config_rule", configRuleName,
		"region", region)

	var nonCompliantResources []logguardiantypes.NonCompliantResource
	var nextToken *string

	// Paginate through all results to handle large numbers of resources
	for {
		input := &configservice.GetComplianceDetailsByConfigRuleInput{
			ConfigRuleName: aws.String(configRuleName),
			ComplianceTypes: []types.ComplianceType{
				types.ComplianceTypeNonCompliant,
			},
			NextToken: nextToken,
			Limit:     s.config.BatchLimit, // Configurable batch limit
		}

		// Add retry logic with exponential backoff for rate limits
		output, err := s.getComplianceDetailsWithRetry(ctx, input, 3)
		if err != nil {
			slog.Error("Failed to get compliance details",
				"config_rule", configRuleName,
				"error", err)
			return nil, fmt.Errorf("failed to get compliance details for rule %s: %w", configRuleName, err)
		}

		// Process evaluation results
		for _, evalResult := range output.EvaluationResults {
			// Filter for CloudWatch Log Groups only
			if evalResult.EvaluationResultIdentifier.EvaluationResultQualifier.ResourceType != nil &&
				*evalResult.EvaluationResultIdentifier.EvaluationResultQualifier.ResourceType == "AWS::Logs::LogGroup" {

				resource := logguardiantypes.NonCompliantResource{
					ResourceId:     aws.ToString(evalResult.EvaluationResultIdentifier.EvaluationResultQualifier.ResourceId),
					ResourceType:   aws.ToString(evalResult.EvaluationResultIdentifier.EvaluationResultQualifier.ResourceType),
					ResourceName:   aws.ToString(evalResult.EvaluationResultIdentifier.EvaluationResultQualifier.ResourceId),
					Region:         region,
					ComplianceType: string(evalResult.ComplianceType),
					Annotation:     aws.ToString(evalResult.Annotation),
					LastEvaluated:  aws.ToTime(evalResult.ResultRecordedTime),
				}

				nonCompliantResources = append(nonCompliantResources, resource)
			}
		}

		// Check if there are more results
		if output.NextToken == nil || aws.ToString(output.NextToken) == "" {
			break
		}
		nextToken = output.NextToken

		// Add small delay between requests to avoid rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	slog.Info("Retrieved non-compliant resources",
		"config_rule", configRuleName,
		"region", region,
		"count", len(nonCompliantResources))

	return nonCompliantResources, nil
}

// ValidateResourceExistence checks if resources still exist before processing
// Note: For Config rule evaluations, we trust that AWS Config has recently evaluated
// these resources. If a log group is deleted between evaluation and remediation,
// the remediation will fail gracefully and the resource won't appear in the next
// Config rule evaluation.
func (s *ConfigEvaluationService) ValidateResourceExistence(ctx context.Context, resources []logguardiantypes.NonCompliantResource) ([]logguardiantypes.NonCompliantResource, error) {
	slog.Info("Trusting Config rule evaluation - skipping resource existence validation",
		"count", len(resources),
		"reason", "Config rules provide recently evaluated resources")

	// Return all resources - let remediation handle any deleted resources gracefully
	return resources, nil
}

// FilterResourcesByComplianceType filters resources by specific compliance issues
func (s *ConfigEvaluationService) FilterResourcesByComplianceType(resources []logguardiantypes.NonCompliantResource, complianceTypes []string) []logguardiantypes.NonCompliantResource {
	if len(complianceTypes) == 0 {
		return resources
	}

	var filtered []logguardiantypes.NonCompliantResource
	typeMap := make(map[string]bool)
	for _, t := range complianceTypes {
		typeMap[strings.ToUpper(t)] = true
	}

	for _, resource := range resources {
		if typeMap[strings.ToUpper(resource.ComplianceType)] {
			filtered = append(filtered, resource)
		}
	}

	slog.Info("Filtered resources by compliance type",
		"original_count", len(resources),
		"filtered_count", len(filtered),
		"types", complianceTypes)

	return filtered
}

// getComplianceDetailsWithRetry implements retry logic with exponential backoff
func (s *ConfigEvaluationService) getComplianceDetailsWithRetry(ctx context.Context, input *configservice.GetComplianceDetailsByConfigRuleInput, maxRetries int) (*configservice.GetComplianceDetailsByConfigRuleOutput, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		output, err := s.configClient.GetComplianceDetailsByConfigRule(ctx, input)
		if err == nil {
			return output, nil
		}

		lastErr = err

		// Check if it's a rate limit error
		if isRateLimitError(err) && attempt < maxRetries {
			// Exponential backoff: 1s, 2s, 4s, etc.
			delay := time.Duration(1<<attempt) * time.Second
			slog.Warn("Rate limit hit, retrying",
				"attempt", attempt+1,
				"delay", delay,
				"error", err)
			time.Sleep(delay)
			continue
		}

		// For non-rate-limit errors or final attempt, return the error
		break
	}

	return nil, lastErr
}

// isRateLimitError checks if an error is a rate limit error
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "Throttling") ||
		strings.Contains(errStr, "TooManyRequests") ||
		strings.Contains(errStr, "RequestLimitExceeded") ||
		strings.Contains(errStr, "ThrottledException")
}
