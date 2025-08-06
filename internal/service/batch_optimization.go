package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/zsoftly/logguardian/internal/types"
)

// BatchKMSValidationCache caches KMS key validation results for a batch operation
type BatchKMSValidationCache struct {
	keyInfo         *KMSKeyInfo
	policyValidated bool
	validationError error
	validatedAt     time.Time
	keyAlias        string
	mu              sync.RWMutex
}

// BatchRemediationContext holds shared context for a batch remediation operation
type BatchRemediationContext struct {
	kmsCache           *BatchKMSValidationCache
	region             string
	configRuleName     string
	batchStartTime     time.Time
	dryRun             bool
	defaultKMSKeyAlias string
	retentionDays      int32
}

// NewBatchRemediationContext creates a new batch context with KMS validation
func (s *ComplianceService) NewBatchRemediationContext(ctx context.Context, request types.BatchComplianceRequest) (*BatchRemediationContext, error) {
	batchCtx := &BatchRemediationContext{
		region:             request.Region,
		configRuleName:     request.ConfigRuleName,
		batchStartTime:     time.Now(),
		dryRun:             s.config.DryRun,
		defaultKMSKeyAlias: s.config.DefaultKMSKeyAlias,
		retentionDays:      s.config.DefaultRetentionDays,
		kmsCache:           &BatchKMSValidationCache{keyAlias: s.config.DefaultKMSKeyAlias},
	}

	slog.Info("Initializing batch remediation context",
		"config_rule", request.ConfigRuleName,
		"region", request.Region,
		"kms_key_alias", s.config.DefaultKMSKeyAlias,
		"dry_run", s.config.DryRun,
		"audit_action", "batch_context_init")

	// Pre-validate KMS key once for the entire batch
	if err := batchCtx.validateKMSKeyForBatch(ctx, s); err != nil {
		slog.Error("Failed to validate KMS key for batch operation",
			"config_rule", request.ConfigRuleName,
			"region", request.Region,
			"kms_key_alias", s.config.DefaultKMSKeyAlias,
			"error", err,
			"audit_action", "batch_kms_validation_failed")
		return nil, fmt.Errorf("batch KMS validation failed: %w", err)
	}

	slog.Info("Batch remediation context initialized successfully",
		"config_rule", request.ConfigRuleName,
		"region", request.Region,
		"kms_key_validated", batchCtx.kmsCache.keyInfo != nil,
		"policy_validated", batchCtx.kmsCache.policyValidated,
		"audit_action", "batch_context_ready")

	return batchCtx, nil
}

// validateKMSKeyForBatch performs one-time KMS key validation for the entire batch
func (bctx *BatchRemediationContext) validateKMSKeyForBatch(ctx context.Context, s *ComplianceService) error {
	bctx.kmsCache.mu.Lock()
	defer bctx.kmsCache.mu.Unlock()

	slog.Info("Performing batch KMS key validation",
		"kms_key_alias", bctx.kmsCache.keyAlias,
		"region", bctx.region,
		"audit_action", "batch_kms_validation_start")

	// Step 1: Validate KMS key accessibility
	keyInfo, err := s.validateKMSKeyAccessibility(ctx, bctx.kmsCache.keyAlias)
	if err != nil {
		bctx.kmsCache.validationError = fmt.Errorf("KMS key accessibility validation failed: %w", err)
		slog.Error("Batch KMS key accessibility validation failed",
			"kms_key_alias", bctx.kmsCache.keyAlias,
			"region", bctx.region,
			"error", err,
			"audit_action", "batch_kms_accessibility_failed")
		return bctx.kmsCache.validationError
	}

	bctx.kmsCache.keyInfo = keyInfo

	slog.Info("Batch KMS key accessibility validation successful",
		"kms_key_alias", bctx.kmsCache.keyAlias,
		"kms_key_id", keyInfo.KeyId,
		"kms_key_arn", keyInfo.Arn,
		"key_state", keyInfo.KeyState,
		"key_region", keyInfo.Region,
		"current_region", bctx.region,
		"is_cross_region", keyInfo.Region != bctx.region,
		"audit_action", "batch_kms_accessibility_success")

	// Step 2: Validate KMS key policy for CloudWatch Logs
	if err := s.validateKMSKeyPolicyForCloudWatchLogs(ctx, keyInfo.KeyId); err != nil {
		// Policy validation failure is a warning, not a fatal error
		slog.Warn("Batch KMS key policy validation warning",
			"kms_key_id", keyInfo.KeyId,
			"error", err,
			"audit_action", "batch_kms_policy_validation_warning",
			"note", "Proceeding with batch operation - ensure key policy allows CloudWatch Logs service")
	} else {
		bctx.kmsCache.policyValidated = true
		slog.Info("Batch KMS key policy validation successful",
			"kms_key_id", keyInfo.KeyId,
			"audit_action", "batch_kms_policy_validation_success")
	}

	bctx.kmsCache.validatedAt = time.Now()

	slog.Info("Batch KMS validation completed successfully",
		"kms_key_alias", bctx.kmsCache.keyAlias,
		"kms_key_id", keyInfo.KeyId,
		"policy_validated", bctx.kmsCache.policyValidated,
		"validation_duration", time.Since(bctx.batchStartTime),
		"audit_action", "batch_kms_validation_complete")

	return nil
}

// GetValidatedKMSKeyInfo returns the pre-validated KMS key info for the batch
func (bctx *BatchRemediationContext) GetValidatedKMSKeyInfo() (*KMSKeyInfo, error) {
	bctx.kmsCache.mu.RLock()
	defer bctx.kmsCache.mu.RUnlock()

	if bctx.kmsCache.validationError != nil {
		return nil, bctx.kmsCache.validationError
	}

	if bctx.kmsCache.keyInfo == nil {
		return nil, fmt.Errorf("KMS key not validated for batch operation")
	}

	return bctx.kmsCache.keyInfo, nil
}

// ProcessNonCompliantResourcesOptimized processes multiple non-compliant resources with optimized KMS validation
func (s *ComplianceService) ProcessNonCompliantResourcesOptimized(ctx context.Context, request types.BatchComplianceRequest) (*types.BatchRemediationResult, error) {
	startTime := time.Now()

	slog.Info("Starting optimized batch remediation",
		"config_rule", request.ConfigRuleName,
		"region", request.Region,
		"total_resources", len(request.NonCompliantResults),
		"batch_size", request.BatchSize,
		"audit_action", "batch_remediation_start")

	// Initialize batch context with one-time KMS validation
	batchCtx, err := s.NewBatchRemediationContext(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize batch context: %w", err)
	}

	result := &types.BatchRemediationResult{
		TotalProcessed: len(request.NonCompliantResults),
		Results:        make([]types.RemediationResult, 0, len(request.NonCompliantResults)),
	}

	// Process resources in batches to avoid overwhelming the AWS APIs
	batchSize := request.BatchSize
	if batchSize <= 0 {
		batchSize = 10 // Default batch size
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	rateLimitCounter := 0

	// Process in parallel batches
	for i := 0; i < len(request.NonCompliantResults); i += batchSize {
		end := i + batchSize
		if end > len(request.NonCompliantResults) {
			end = len(request.NonCompliantResults)
		}

		batch := request.NonCompliantResults[i:end]
		wg.Add(1)

		go func(batchResources []types.NonCompliantResource, batchIndex int) {
			defer wg.Done()

			slog.Info("Processing optimized batch",
				"batch_index", batchIndex,
				"batch_size", len(batchResources),
				"config_rule", request.ConfigRuleName)

			// Process each resource in the batch using pre-validated KMS info
			for _, resource := range batchResources {
				// Convert to ComplianceResult format for this specific Config rule
				compliance := s.convertToComplianceResultForRule(batchCtx.configRuleName, resource)

				// Use optimized remediation with pre-validated KMS info
				remediationResult, err := s.remediateLogGroupWithBatchContext(ctx, compliance, batchCtx)

				mu.Lock()
				if err != nil {
					// Handle rate limiting with exponential backoff
					if isRateLimitError(err) {
						rateLimitCounter++
						slog.Warn("Rate limit encountered in optimized batch",
							"resource", resource.ResourceName,
							"batch_index", batchIndex,
							"error", err)

						// Exponential backoff with jitter
						delay := time.Duration(1+rateLimitCounter) * time.Second
						slog.Info("Retrying with exponential backoff", "delay", delay, "batch_index", batchIndex)
						time.Sleep(delay)

						// Retry with batch context
						remediationResult, err = s.remediateLogGroupWithBatchContext(ctx, compliance, batchCtx)
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

				// Small delay between resources in the same batch to prevent overwhelming APIs
				time.Sleep(50 * time.Millisecond) // Reduced from 100ms since KMS validation is cached
			}

			slog.Info("Optimized batch completed",
				"batch_index", batchIndex,
				"batch_size", len(batchResources))

		}(batch, i/batchSize)

		// Rate limiting: delay between batches (reduced since KMS validation is cached)
		time.Sleep(200 * time.Millisecond) // Reduced from 500ms
	}

	// Wait for all batches to complete
	wg.Wait()

	result.ProcessingDuration = time.Since(startTime)
	result.RateLimitHits = rateLimitCounter

	slog.Info("Optimized batch remediation completed",
		"total_processed", result.TotalProcessed,
		"success_count", result.SuccessCount,
		"failure_count", result.FailureCount,
		"processing_duration", result.ProcessingDuration,
		"rate_limit_hits", rateLimitCounter,
		"kms_validation_cached", true,
		"performance_improvement", "eliminated_repeated_kms_validation",
		"audit_action", "batch_remediation_complete")

	return result, nil
}

// remediateLogGroupWithBatchContext applies compliance remediation using pre-validated batch context
func (s *ComplianceService) remediateLogGroupWithBatchContext(ctx context.Context, compliance types.ComplianceResult, batchCtx *BatchRemediationContext) (*types.RemediationResult, error) {
	result := &types.RemediationResult{
		LogGroupName: compliance.LogGroupName,
		Region:       compliance.Region,
		Success:      true,
	}

	slog.Info("Starting optimized remediation with batch context",
		"log_group", compliance.LogGroupName,
		"region", compliance.Region,
		"dry_run", batchCtx.dryRun,
		"kms_pre_validated", batchCtx.kmsCache.keyInfo != nil)

	// Apply KMS encryption if missing (using pre-validated KMS info)
	if compliance.MissingEncryption {
		if err := s.applyEncryptionWithBatchContext(ctx, compliance.LogGroupName, batchCtx); err != nil {
			result.Success = false
			result.Error = fmt.Errorf("failed to apply encryption: %w", err)
			return result, err
		}
		result.EncryptionApplied = true
		slog.Info("Applied KMS encryption using batch context",
			"log_group", compliance.LogGroupName,
			"kms_key_id", batchCtx.kmsCache.keyInfo.KeyId)
	}

	// Apply retention policy if missing (no optimization needed here, but using batch context for consistency)
	if compliance.MissingRetention {
		if err := s.applyRetentionPolicyWithBatchContext(ctx, compliance.LogGroupName, batchCtx); err != nil {
			result.Success = false
			result.Error = fmt.Errorf("failed to apply retention policy: %w", err)
			return result, err
		}
		result.RetentionApplied = true
		slog.Info("Applied retention policy using batch context",
			"log_group", compliance.LogGroupName,
			"retention_days", batchCtx.retentionDays)
	}

	return result, nil
}

// applyEncryptionWithBatchContext applies KMS encryption using pre-validated batch context
func (s *ComplianceService) applyEncryptionWithBatchContext(ctx context.Context, logGroupName string, batchCtx *BatchRemediationContext) error {
	if batchCtx.dryRun {
		slog.Info("DRY RUN: Would apply KMS encryption with batch context",
			"log_group", logGroupName,
			"kms_key_alias", batchCtx.kmsCache.keyAlias,
			"kms_key_id", batchCtx.kmsCache.keyInfo.KeyId,
			"audit_action", AuditActionEncryptionDryRun,
			"batch_optimized", true)
		return nil
	}

	// Get pre-validated KMS key info from batch context
	keyInfo, err := batchCtx.GetValidatedKMSKeyInfo()
	if err != nil {
		return fmt.Errorf("failed to get validated KMS key info: %w", err)
	}

	slog.Info("Applying KMS encryption with pre-validated key info",
		"log_group", logGroupName,
		"kms_key_id", keyInfo.KeyId,
		"kms_key_arn", keyInfo.Arn,
		"batch_optimized", true,
		"audit_action", AuditActionEncryptionStart)

	// Associate KMS key with retry logic (same as before)
	if err := s.associateKMSKeyWithRetry(ctx, logGroupName, keyInfo.Arn); err != nil {
		slog.Error("Failed to associate KMS key with batch context",
			"log_group", logGroupName,
			"kms_key_arn", keyInfo.Arn,
			"error", err,
			"audit_action", AuditActionEncryptionFailed)
		return fmt.Errorf("failed to associate KMS key %s with log group %s: %w", keyInfo.Arn, logGroupName, err)
	}

	slog.Info("Successfully applied KMS encryption with batch optimization",
		"log_group", logGroupName,
		"kms_key_id", keyInfo.KeyId,
		"kms_key_arn", keyInfo.Arn,
		"batch_optimized", true,
		"audit_action", AuditActionEncryptionSuccess)

	return nil
}

// applyRetentionPolicyWithBatchContext applies retention policy using batch context
func (s *ComplianceService) applyRetentionPolicyWithBatchContext(ctx context.Context, logGroupName string, batchCtx *BatchRemediationContext) error {
	if batchCtx.dryRun {
		slog.Info("DRY RUN: Would apply retention policy with batch context",
			"log_group", logGroupName,
			"retention_days", batchCtx.retentionDays,
			"batch_optimized", true)
		return nil
	}

	input := &cloudwatchlogs.PutRetentionPolicyInput{
		LogGroupName:    aws.String(logGroupName),
		RetentionInDays: aws.Int32(batchCtx.retentionDays),
	}

	_, err := s.logsClient.PutRetentionPolicy(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to set retention policy for log group %s: %w", logGroupName, err)
	}

	slog.Info("Successfully set retention policy with batch optimization",
		"log_group", logGroupName,
		"retention_days", batchCtx.retentionDays,
		"batch_optimized", true)

	return nil
}
