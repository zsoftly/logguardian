package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/zsoftly/logguardian/internal/types"
)

// ComplianceService handles log group compliance remediation
type ComplianceService struct {
	logsClient        CloudWatchLogsClientInterface
	kmsClient         KMSClientInterface
	configClient      ConfigServiceClientInterface
	configEvalService *ConfigEvaluationService
	config            ServiceConfig
}

// ServiceConfig holds configuration for the compliance service
type ServiceConfig struct {
	DefaultKMSKeyAlias   string
	DefaultRetentionDays int32
	DryRun               bool
	BatchLimit           int32
}

// NewComplianceService creates a new compliance service
func NewComplianceService(cfg aws.Config) *ComplianceService {
	// Load configuration from environment variables
	config := ServiceConfig{
		DefaultKMSKeyAlias:   getEnvOrDefault("KMS_KEY_ALIAS", "alias/cloudwatch-logs-compliance"),
		DefaultRetentionDays: getEnvAsInt32OrDefault("DEFAULT_RETENTION_DAYS", 365),
		DryRun:               getEnvAsBoolOrDefault("DRY_RUN", false),
		BatchLimit:           getEnvAsInt32OrDefault("BATCH_LIMIT", 100),
	}

	return &ComplianceService{
		logsClient:        cloudwatchlogs.NewFromConfig(cfg),
		kmsClient:         kms.NewFromConfig(cfg),
		configClient:      configservice.NewFromConfig(cfg),
		configEvalService: NewConfigEvaluationService(cfg),
		config:            config,
	}
}

// RemediateLogGroup applies compliance remediation to a log group
func (s *ComplianceService) RemediateLogGroup(ctx context.Context, compliance types.ComplianceResult) (*types.RemediationResult, error) {
	result := &types.RemediationResult{
		LogGroupName: compliance.LogGroupName,
		Region:       compliance.Region,
		Success:      true,
	}

	slog.Info("Starting remediation",
		"log_group", compliance.LogGroupName,
		"region", compliance.Region,
		"dry_run", s.config.DryRun)

	// Apply KMS encryption if missing
	if compliance.MissingEncryption {
		if err := s.applyEncryption(ctx, compliance.LogGroupName); err != nil {
			result.Success = false
			result.Error = fmt.Errorf("failed to apply encryption: %w", err)
			return result, err
		}
		result.EncryptionApplied = true
		slog.Info("Applied KMS encryption", "log_group", compliance.LogGroupName)
	}

	// Apply retention policy if missing
	if compliance.MissingRetention {
		if err := s.applyRetentionPolicy(ctx, compliance.LogGroupName); err != nil {
			result.Success = false
			result.Error = fmt.Errorf("failed to apply retention policy: %w", err)
			return result, err
		}
		result.RetentionApplied = true
		slog.Info("Applied retention policy",
			"log_group", compliance.LogGroupName,
			"retention_days", s.config.DefaultRetentionDays)
	}

	return result, nil
}

// applyEncryption associates a KMS key with the log group
func (s *ComplianceService) applyEncryption(ctx context.Context, logGroupName string) error {
	if s.config.DryRun {
		slog.Info("DRY RUN: Would apply KMS encryption",
			"log_group", logGroupName,
			"kms_key_alias", s.config.DefaultKMSKeyAlias)
		return nil
	}

	// First, resolve the KMS key alias to get the key ID
	keyID, err := s.resolveKMSKeyAlias(ctx, s.config.DefaultKMSKeyAlias)
	if err != nil {
		return fmt.Errorf("failed to resolve KMS key alias %s: %w", s.config.DefaultKMSKeyAlias, err)
	}

	// Associate the KMS key with the log group
	input := &cloudwatchlogs.AssociateKmsKeyInput{
		LogGroupName: aws.String(logGroupName),
		KmsKeyId:     aws.String(keyID),
	}

	_, err = s.logsClient.AssociateKmsKey(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to associate KMS key with log group %s: %w", logGroupName, err)
	}

	slog.Info("Successfully associated KMS key",
		"log_group", logGroupName,
		"kms_key_id", keyID)

	return nil
}

// applyRetentionPolicy sets the retention policy on the log group
func (s *ComplianceService) applyRetentionPolicy(ctx context.Context, logGroupName string) error {
	if s.config.DryRun {
		slog.Info("DRY RUN: Would apply retention policy",
			"log_group", logGroupName,
			"retention_days", s.config.DefaultRetentionDays)
		return nil
	}

	input := &cloudwatchlogs.PutRetentionPolicyInput{
		LogGroupName:    aws.String(logGroupName),
		RetentionInDays: aws.Int32(s.config.DefaultRetentionDays),
	}

	_, err := s.logsClient.PutRetentionPolicy(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to set retention policy for log group %s: %w", logGroupName, err)
	}

	slog.Info("Successfully set retention policy",
		"log_group", logGroupName,
		"retention_days", s.config.DefaultRetentionDays)

	return nil
}

// resolveKMSKeyAlias resolves a KMS key alias to the actual key ID
func (s *ComplianceService) resolveKMSKeyAlias(ctx context.Context, keyAlias string) (string, error) {
	input := &kms.DescribeKeyInput{
		KeyId: aws.String(keyAlias),
	}

	result, err := s.kmsClient.DescribeKey(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to describe KMS key %s: %w", keyAlias, err)
	}

	if result.KeyMetadata == nil || result.KeyMetadata.KeyId == nil {
		return "", fmt.Errorf("invalid KMS key metadata for %s", keyAlias)
	}

	return *result.KeyMetadata.KeyId, nil
}

// Utility functions for environment variable handling
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt32OrDefault(key string, defaultValue int32) int32 {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.ParseInt(valueStr, 10, 32); err == nil {
			return int32(value)
		}
	}
	return defaultValue
}

func getEnvAsBoolOrDefault(key string, defaultValue bool) bool {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.ParseBool(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

// ProcessNonCompliantResources processes multiple non-compliant resources in batches
func (s *ComplianceService) ProcessNonCompliantResources(ctx context.Context, request types.BatchComplianceRequest) (*types.BatchRemediationResult, error) {
	startTime := time.Now()

	slog.Info("Starting batch remediation",
		"config_rule", request.ConfigRuleName,
		"region", request.Region,
		"total_resources", len(request.NonCompliantResults),
		"batch_size", request.BatchSize)

	result := &types.BatchRemediationResult{
		TotalProcessed: len(request.NonCompliantResults),
		Results:        make([]types.RemediationResult, 0),
	}

	// Process resources in batches to avoid overwhelming the AWS APIs
	batchSize := request.BatchSize
	if batchSize <= 0 {
		batchSize = 10 // Default batch size
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	rateLimitCounter := 0

	for i := 0; i < len(request.NonCompliantResults); i += batchSize {
		end := i + batchSize
		if end > len(request.NonCompliantResults) {
			end = len(request.NonCompliantResults)
		}

		batch := request.NonCompliantResults[i:end]
		wg.Add(1)

		go func(batchResources []types.NonCompliantResource, batchIndex int) {
			defer wg.Done()

			slog.Info("Processing batch",
				"batch_index", batchIndex,
				"batch_size", len(batchResources))

			for _, resource := range batchResources {
				// Convert to ComplianceResult format
				compliance := s.convertToComplianceResult(resource)

				// Process the resource
				remediationResult, err := s.RemediateLogGroup(ctx, compliance)

				mu.Lock()
				if err != nil {
					// Handle rate limiting with exponential backoff
					if isRateLimitError(err) {
						rateLimitCounter++
						slog.Warn("Rate limit encountered",
							"resource", resource.ResourceName,
							"error", err)

						// Exponential backoff: start with 1 second for first retry
						delay := 1 * time.Second
						slog.Info("Retrying with exponential backoff", "delay", delay)
						time.Sleep(delay)
						remediationResult, err = s.RemediateLogGroup(ctx, compliance)
					}

					if err != nil {
						result.FailureCount++
						remediationResult = &types.RemediationResult{
							LogGroupName: compliance.LogGroupName,
							Region:       compliance.Region,
							Success:      false,
							Error:        err,
						}
					} else {
						result.SuccessCount++
					}
				} else {
					result.SuccessCount++
				}

				result.Results = append(result.Results, *remediationResult)
				mu.Unlock()

				// Small delay between resources in the same batch
				time.Sleep(100 * time.Millisecond)
			}

			slog.Info("Batch completed", "batch_index", batchIndex)
		}(batch, i/batchSize)

		// Rate limiting: delay between batches
		time.Sleep(500 * time.Millisecond)
	}

	// Wait for all batches to complete
	wg.Wait()

	result.ProcessingDuration = time.Since(startTime)
	result.RateLimitHits = rateLimitCounter

	slog.Info("Batch remediation completed",
		"total_processed", result.TotalProcessed,
		"success_count", result.SuccessCount,
		"failure_count", result.FailureCount,
		"duration", result.ProcessingDuration,
		"rate_limit_hits", result.RateLimitHits)

	return result, nil
}

// GetNonCompliantResources retrieves non-compliant log groups from Config API
func (s *ComplianceService) GetNonCompliantResources(ctx context.Context, configRuleName string, region string) ([]types.NonCompliantResource, error) {
	return s.configEvalService.GetNonCompliantResources(ctx, configRuleName, region)
}

// ValidateResourceExistence checks if resources still exist before processing
func (s *ComplianceService) ValidateResourceExistence(ctx context.Context, resources []types.NonCompliantResource) ([]types.NonCompliantResource, error) {
	return s.configEvalService.ValidateResourceExistence(ctx, resources)
}

// Helper methods

// convertToComplianceResult converts a NonCompliantResource to ComplianceResult
func (s *ComplianceService) convertToComplianceResult(resource types.NonCompliantResource) types.ComplianceResult {
	// Parse the compliance annotation to determine what's missing
	missingEncryption := false
	missingRetention := false

	annotation := resource.Annotation
	if annotation != "" {
		if strings.Contains(strings.ToLower(annotation), "encryption") ||
			strings.Contains(strings.ToLower(annotation), "kms") {
			missingEncryption = true
		}
		if strings.Contains(strings.ToLower(annotation), "retention") {
			missingRetention = true
		}
	}

	// If annotation doesn't specify, assume both are missing for non-compliant resources
	if !missingEncryption && !missingRetention {
		missingEncryption = true
		missingRetention = true
	}

	return types.ComplianceResult{
		LogGroupName:      resource.ResourceName,
		Region:            resource.Region,
		AccountId:         resource.AccountId,
		MissingEncryption: missingEncryption,
		MissingRetention:  missingRetention,
		CurrentRetention:  nil, // Will be determined during remediation
		CurrentKmsKeyId:   "",  // Will be determined during remediation
	}
}
