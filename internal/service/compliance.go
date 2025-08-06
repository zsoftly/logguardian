package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cloudwatchlogstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	kmstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/aws/smithy-go"
	"github.com/zsoftly/logguardian/internal/types"
)

// Audit action constants for consistent logging and compliance tracking
const (
	// Encryption audit actions
	AuditActionEncryptionStart   = "encryption_start"
	AuditActionEncryptionSuccess = "encryption_success"
	AuditActionEncryptionFailed  = "encryption_failed"
	AuditActionEncryptionDryRun  = "encryption_dry_run"

	// Key validation audit actions
	AuditActionKeyValidationSuccess = "key_validation_success"
	AuditActionKeyValidationFailed  = "key_validation_failed"
	AuditActionCrossRegionKeyUsage  = "cross_region_key_usage"

	// Policy validation audit actions
	AuditActionPolicyValidationSuccess = "policy_validation_success"
	AuditActionPolicyValidationWarning = "policy_validation_warning"

	// Comprehensive validation audit actions
	AuditActionComprehensiveKMSValidation = "comprehensive_kms_validation"

	// Failure reason constants
	FailureReasonKeyNotFound      = "key_not_found"
	FailureReasonAccessDenied     = "access_denied"
	FailureReasonGeneralError     = "general_error"
	FailureReasonInvalidMetadata  = "invalid_metadata"
	FailureReasonMissingKeyID     = "missing_key_id"
	FailureReasonMissingKeyARN    = "missing_key_arn"
	FailureReasonUnusableKeyState = "unusable_key_state"

	// Failure stage constants
	FailureStageKeyValidation    = "key_validation"
	FailureStagePolicyValidation = "policy_validation"
	FailureStageKeyAssociation   = "key_association"

	// Retry logic constants
	MaxExponentialBackoffAttempts = 10   // Maximum attempts before capping multiplier to prevent overflow
	MaxBackoffMultiplier          = 1024 // 2^10, maximum multiplier for exponential backoff
)

// ComplianceService handles log group compliance remediation
type ComplianceService struct {
	logsClient        CloudWatchLogsClientInterface
	kmsClient         KMSClientInterface
	configClient      ConfigServiceClientInterface
	configEvalService *ConfigEvaluationService
	ruleClassifier    *types.RuleClassifier
	config            ServiceConfig
}

// ServiceConfig holds configuration for the compliance service
type ServiceConfig struct {
	DefaultKMSKeyAlias   string
	DefaultRetentionDays int32
	DryRun               bool
	BatchLimit           int32
	Region               string
	MaxKMSRetries        int32
	RetryBaseDelay       time.Duration
	BatchResourceDelay   time.Duration
	BatchGroupDelay      time.Duration
}

// NewComplianceService creates a new compliance service
func NewComplianceService(cfg aws.Config) *ComplianceService {
	// Load configuration from environment variables
	region := getEnvOrDefault("AWS_REGION", "")
	if region == "" {
		region = getEnvOrDefault("AWS_DEFAULT_REGION", "ca-central-1")
	}
	config := ServiceConfig{
		DefaultKMSKeyAlias:   getEnvOrDefault("KMS_KEY_ALIAS", "alias/cloudwatch-logs-compliance"),
		DefaultRetentionDays: getEnvAsInt32OrDefault("DEFAULT_RETENTION_DAYS", 365),
		DryRun:               getEnvAsBoolOrDefault("DRY_RUN", false),
		BatchLimit:           getEnvAsInt32OrDefault("BATCH_LIMIT", 100),
		Region:               region,
		MaxKMSRetries:        getEnvAsInt32OrDefault("MAX_KMS_RETRIES", 3),
		RetryBaseDelay:       time.Duration(getEnvAsInt32OrDefault("RETRY_BASE_DELAY_MS", 1000)) * time.Millisecond,
		BatchResourceDelay:   time.Duration(getEnvAsInt32OrDefault("BATCH_RESOURCE_DELAY_MS", 50)) * time.Millisecond,
		BatchGroupDelay:      time.Duration(getEnvAsInt32OrDefault("BATCH_GROUP_DELAY_MS", 200)) * time.Millisecond,
	}

	return &ComplianceService{
		logsClient:        cloudwatchlogs.NewFromConfig(cfg),
		kmsClient:         kms.NewFromConfig(cfg),
		configClient:      configservice.NewFromConfig(cfg),
		configEvalService: NewConfigEvaluationService(cfg),
		ruleClassifier:    types.NewRuleClassifier(),
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
	// Cache the current region to avoid repeated function calls
	currentRegion := s.getCurrentRegion()

	if s.config.DryRun {
		slog.Info("DRY RUN: Would apply KMS encryption",
			"log_group", logGroupName,
			"kms_key_alias", s.config.DefaultKMSKeyAlias,
			"audit_action", AuditActionEncryptionDryRun,
			"timestamp", time.Now().UTC().Format(time.RFC3339))
		return nil
	}

	slog.Info("Starting KMS encryption process",
		"log_group", logGroupName,
		"kms_key_alias", s.config.DefaultKMSKeyAlias,
		"audit_action", AuditActionEncryptionStart,
		"timestamp", time.Now().UTC().Format(time.RFC3339))

	// Step 1: Validate KMS key existence and accessibility
	keyInfo, err := s.validateKMSKeyAccessibility(ctx, s.config.DefaultKMSKeyAlias)
	if err != nil {
		slog.Error("KMS key validation failed during encryption",
			"log_group", logGroupName,
			"kms_key_alias", s.config.DefaultKMSKeyAlias,
			"error", err,
			"audit_action", AuditActionEncryptionFailed,
			"failure_stage", FailureStageKeyValidation,
			"timestamp", time.Now().UTC().Format(time.RFC3339))
		return fmt.Errorf("KMS key validation failed for %s: %w", s.config.DefaultKMSKeyAlias, err)
	}

	slog.Info("KMS key validation successful",
		"log_group", logGroupName,
		"kms_key_alias", s.config.DefaultKMSKeyAlias,
		"kms_key_id", keyInfo.KeyId,
		"kms_key_arn", keyInfo.Arn,
		"key_state", keyInfo.KeyState,
		"key_region", keyInfo.Region,
		"audit_action", AuditActionKeyValidationSuccess)

	// Step 2: Verify key policies allow CloudWatch Logs service
	if err := s.validateKMSKeyPolicyForCloudWatchLogs(ctx, keyInfo.KeyId); err != nil {
		slog.Error("KMS key policy validation failed during encryption",
			"log_group", logGroupName,
			"kms_key_id", keyInfo.KeyId,
			"error", err,
			"audit_action", AuditActionEncryptionFailed,
			"failure_stage", FailureStagePolicyValidation,
			"timestamp", time.Now().UTC().Format(time.RFC3339))
		return fmt.Errorf("KMS key policy validation failed for %s: %w", keyInfo.KeyId, err)
	}

	slog.Info("KMS key policy validation successful",
		"log_group", logGroupName,
		"kms_key_id", keyInfo.KeyId,
		"audit_action", AuditActionPolicyValidationSuccess)

	// Step 3: Apply encryption with proper error handling
	if err := s.associateKMSKeyWithRetry(ctx, logGroupName, keyInfo.Arn); err != nil {
		slog.Error("Failed to associate KMS key with log group",
			"log_group", logGroupName,
			"kms_key_arn", keyInfo.Arn,
			"error", err,
			"audit_action", AuditActionEncryptionFailed,
			"failure_stage", FailureStageKeyAssociation,
			"timestamp", time.Now().UTC().Format(time.RFC3339))
		return fmt.Errorf("failed to associate KMS key with log group %s: %w", logGroupName, err)
	}

	// Step 4: Log operation for comprehensive audit trail
	slog.Info("Successfully applied KMS encryption",
		"log_group", logGroupName,
		"kms_key_alias", s.config.DefaultKMSKeyAlias,
		"kms_key_id", keyInfo.KeyId,
		"kms_key_arn", keyInfo.Arn,
		"key_region", keyInfo.Region,
		"current_region", currentRegion,
		"is_cross_region", keyInfo.Region != currentRegion,
		"operation", "associate_kms_key",
		"audit_action", AuditActionEncryptionSuccess,
		"security_enhancement", "log_group_encrypted",
		"compliance_status", "encryption_applied",
		"timestamp", time.Now().UTC().Format(time.RFC3339))

	return nil
}

// ValidateKMSKeyComprehensively provides a comprehensive validation report for a KMS key
// This function is useful for troubleshooting and audit purposes
func (s *ComplianceService) ValidateKMSKeyComprehensively(ctx context.Context, keyAlias string) (*types.KMSValidationReport, error) {
	report := &types.KMSValidationReport{
		KeyAlias:            keyAlias,
		CurrentRegion:       s.getCurrentRegion(),
		ValidationTimestamp: time.Now().UTC(),
		ValidationErrors:    []string{},
		ValidationWarnings:  []string{},
		RecommendedActions:  []string{},
	}

	// Step 1: Test key accessibility
	keyInfo, err := s.validateKMSKeyAccessibility(ctx, keyAlias)
	if err != nil {
		report.KeyExists = false
		report.KeyAccessible = false
		report.ValidationErrors = append(report.ValidationErrors, err.Error())

		if isKMSKeyNotFoundError(err) {
			report.RecommendedActions = append(report.RecommendedActions,
				fmt.Sprintf("Create KMS key with alias %s in region %s", keyAlias, report.CurrentRegion))
		} else if isKMSAccessDeniedError(err) {
			report.RecommendedActions = append(report.RecommendedActions,
				"Ensure Lambda execution role has kms:DescribeKey permissions")
		}

		return report, nil
	}

	// Key exists and is accessible
	report.KeyExists = true
	report.KeyAccessible = true
	report.KeyId = keyInfo.KeyId
	report.KeyArn = keyInfo.Arn
	report.KeyState = keyInfo.KeyState
	report.KeyRegion = keyInfo.Region
	report.IsCrossRegion = keyInfo.Region != report.CurrentRegion

	if report.IsCrossRegion {
		report.ValidationWarnings = append(report.ValidationWarnings,
			fmt.Sprintf("KMS key is in region %s but Lambda is running in %s", keyInfo.Region, report.CurrentRegion))
		report.RecommendedActions = append(report.RecommendedActions,
			"Consider using a KMS key in the same region for better performance and to avoid cross-region charges")
	}

	// Step 2: Test key policy accessibility and CloudWatch Logs permissions
	policyInput := &kms.GetKeyPolicyInput{
		KeyId:      aws.String(keyInfo.KeyId),
		PolicyName: aws.String("default"),
	}

	policyResult, err := s.kmsClient.GetKeyPolicy(ctx, policyInput)
	if err != nil {
		report.PolicyAccessible = false
		report.ValidationWarnings = append(report.ValidationWarnings,
			fmt.Sprintf("Cannot access key policy: %v", err))
		report.RecommendedActions = append(report.RecommendedActions,
			"Ensure Lambda execution role has kms:GetKeyPolicy permissions")
	} else {
		report.PolicyAccessible = true

		// Check for CloudWatch Logs access in policy
		if policyResult.Policy != nil {
			policy := *policyResult.Policy
			report.CloudWatchLogsAccess = s.checkCloudWatchLogsPolicyAccess(policy)

			if !report.CloudWatchLogsAccess {
				report.ValidationWarnings = append(report.ValidationWarnings,
					"KMS key policy may not allow CloudWatch Logs service access")
				report.RecommendedActions = append(report.RecommendedActions,
					"Update KMS key policy to allow CloudWatch Logs service principal: logs.amazonaws.com")
			}
		}
	}

	// Log the comprehensive validation results
	slog.Info("Comprehensive KMS key validation completed",
		"key_alias", keyAlias,
		"key_exists", report.KeyExists,
		"key_accessible", report.KeyAccessible,
		"policy_accessible", report.PolicyAccessible,
		"cloudwatch_logs_access", report.CloudWatchLogsAccess,
		"is_cross_region", report.IsCrossRegion,
		"validation_errors", len(report.ValidationErrors),
		"validation_warnings", len(report.ValidationWarnings),
		"audit_action", AuditActionComprehensiveKMSValidation)

	return report, nil
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

// KMSKeyInfo holds comprehensive information about a KMS key
type KMSKeyInfo struct {
	KeyId    string
	Arn      string
	KeyState string
	Region   string
}

// validateKMSKeyAccessibility validates KMS key existence and accessibility
func (s *ComplianceService) validateKMSKeyAccessibility(ctx context.Context, keyAlias string) (*KMSKeyInfo, error) {
	// Cache the current region to avoid repeated function calls
	currentRegion := s.getCurrentRegion()

	slog.Info("Validating KMS key accessibility",
		"kms_key_alias", keyAlias,
		"current_region", currentRegion)

	input := &kms.DescribeKeyInput{
		KeyId: aws.String(keyAlias),
	}

	result, err := s.kmsClient.DescribeKey(ctx, input)
	if err != nil {
		// Check for specific KMS errors
		if isKMSKeyNotFoundError(err) {
			// Log detailed error for audit trail
			slog.Error("KMS key not found during validation",
				"kms_key_alias", keyAlias,
				"current_region", currentRegion,
				"error", err,
				"audit_action", AuditActionKeyValidationFailed,
				"failure_reason", FailureReasonKeyNotFound)
			return nil, fmt.Errorf("KMS key not found: %s. Please ensure the key exists and is accessible in region %s", keyAlias, currentRegion)
		}
		if isKMSAccessDeniedError(err) {
			// Log detailed error for audit trail
			slog.Error("KMS key access denied during validation",
				"kms_key_alias", keyAlias,
				"current_region", currentRegion,
				"error", err,
				"audit_action", AuditActionKeyValidationFailed,
				"failure_reason", FailureReasonAccessDenied)
			return nil, fmt.Errorf("access denied to KMS key: %s. Please ensure proper IAM permissions are configured for region %s", keyAlias, currentRegion)
		}

		// Log general errors with audit information
		slog.Error("KMS key validation failed",
			"kms_key_alias", keyAlias,
			"current_region", currentRegion,
			"error", err,
			"audit_action", AuditActionKeyValidationFailed,
			"failure_reason", FailureReasonGeneralError)
		return nil, fmt.Errorf("failed to describe KMS key %s in region %s: %w", keyAlias, currentRegion, err)
	}

	if result.KeyMetadata == nil {
		slog.Error("Invalid KMS key metadata received",
			"kms_key_alias", keyAlias,
			"audit_action", AuditActionKeyValidationFailed,
			"failure_reason", FailureReasonInvalidMetadata)
		return nil, fmt.Errorf("invalid KMS key metadata for %s", keyAlias)
	}

	keyMetadata := result.KeyMetadata

	// Validate required fields
	if keyMetadata.KeyId == nil {
		slog.Error("KMS key ID missing in metadata",
			"kms_key_alias", keyAlias,
			"audit_action", AuditActionKeyValidationFailed,
			"failure_reason", FailureReasonMissingKeyID)
		return nil, fmt.Errorf("KMS key ID is missing for %s", keyAlias)
	}
	if keyMetadata.Arn == nil {
		slog.Error("KMS key ARN missing in metadata",
			"kms_key_alias", keyAlias,
			"audit_action", AuditActionKeyValidationFailed,
			"failure_reason", FailureReasonMissingKeyARN)
		return nil, fmt.Errorf("KMS key ARN is missing for %s", keyAlias)
	}

	keyInfo := &KMSKeyInfo{
		KeyId:    *keyMetadata.KeyId,
		Arn:      *keyMetadata.Arn,
		KeyState: string(keyMetadata.KeyState),
	}

	// Extract region from ARN (format: arn:aws:kms:region:account:key/key-id)
	if parts := strings.Split(*keyMetadata.Arn, ":"); len(parts) >= 4 {
		keyInfo.Region = parts[3]

		// Cross-region validation: warn if key is in different region
		if keyInfo.Region != currentRegion {
			slog.Warn("KMS key is in different region than current",
				"kms_key_alias", keyAlias,
				"key_region", keyInfo.Region,
				"current_region", currentRegion,
				"audit_action", AuditActionCrossRegionKeyUsage,
				"note", "Using cross-region KMS key - ensure proper permissions and network access")
		}
	}

	// Validate key state
	if err := s.validateKMSKeyState(keyMetadata.KeyState); err != nil {
		slog.Error("KMS key is not in usable state",
			"kms_key_alias", keyAlias,
			"kms_key_id", keyInfo.KeyId,
			"key_state", keyInfo.KeyState,
			"error", err,
			"audit_action", AuditActionKeyValidationFailed,
			"failure_reason", FailureReasonUnusableKeyState)
		return nil, fmt.Errorf("KMS key %s is not in a usable state: %w", keyAlias, err)
	}

	// Log successful validation with comprehensive audit information
	slog.Info("KMS key accessibility validation completed successfully",
		"kms_key_alias", keyAlias,
		"kms_key_id", keyInfo.KeyId,
		"kms_key_arn", keyInfo.Arn,
		"key_state", keyInfo.KeyState,
		"key_region", keyInfo.Region,
		"current_region", currentRegion,
		"is_cross_region", keyInfo.Region != currentRegion,
		"audit_action", AuditActionKeyValidationSuccess,
		"validation_timestamp", time.Now().UTC().Format(time.RFC3339))

	return keyInfo, nil
}

// validateKMSKeyState checks if the KMS key is in a usable state
func (s *ComplianceService) validateKMSKeyState(keyState kmstypes.KeyState) error {
	switch keyState {
	case kmstypes.KeyStateEnabled:
		return nil // Key is usable
	case kmstypes.KeyStateDisabled:
		return fmt.Errorf("key is disabled")
	case kmstypes.KeyStatePendingDeletion:
		return fmt.Errorf("key is pending deletion")
	case kmstypes.KeyStatePendingImport:
		return fmt.Errorf("key is pending import")
	case kmstypes.KeyStateUnavailable:
		return fmt.Errorf("key is unavailable")
	default:
		return fmt.Errorf("KMS key is in unknown state: %s", string(keyState))
	}
}

// checkCloudWatchLogsPolicyAccess checks if a policy contains CloudWatch Logs service access
func (s *ComplianceService) checkCloudWatchLogsPolicyAccess(policy string) bool {
	// Cache the current region to avoid repeated function calls
	currentRegion := s.getCurrentRegion()

	// Check if the policy contains CloudWatch Logs service principals
	// Support both generic and region-specific service principals
	requiredPrincipals := []string{
		"logs.amazonaws.com",
		fmt.Sprintf("logs.%s.amazonaws.com", currentRegion),
	}

	// Also check for AWS service principal patterns in the policy
	servicePatterns := []string{
		`"Service": "logs.amazonaws.com"`,
		fmt.Sprintf(`"Service": "logs.%s.amazonaws.com"`, currentRegion),
		`"Service": ["logs.amazonaws.com"`,
		`"Service":["logs.amazonaws.com"`,
		`"Service": [ "logs.amazonaws.com"`,
		`"logs.amazonaws.com"`,
	}

	// Check for required principals first
	for _, principal := range requiredPrincipals {
		if strings.Contains(policy, principal) {
			return true
		}
	}

	// Check for service patterns
	for _, pattern := range servicePatterns {
		if strings.Contains(policy, pattern) {
			return true
		}
	}

	return false
}

// validateKMSKeyPolicyForCloudWatchLogs verifies key policies allow CloudWatch Logs service
func (s *ComplianceService) validateKMSKeyPolicyForCloudWatchLogs(ctx context.Context, keyId string) error {
	slog.Info("Validating KMS key policy for CloudWatch Logs access",
		"kms_key_id", keyId)

	// Get the key policy
	policyInput := &kms.GetKeyPolicyInput{
		KeyId:      aws.String(keyId),
		PolicyName: aws.String("default"),
	}

	policyResult, err := s.kmsClient.GetKeyPolicy(ctx, policyInput)
	if err != nil {
		// If we can't access the policy, log a warning but don't fail
		// This allows customers to use keys where they don't have GetKeyPolicy permissions
		slog.Warn("Cannot access KMS key policy for validation",
			"kms_key_id", keyId,
			"error", err,
			"note", "Proceeding with encryption attempt - ensure key policy allows CloudWatch Logs service")
		return nil
	}

	if policyResult.Policy == nil {
		slog.Warn("KMS key policy is empty",
			"kms_key_id", keyId,
			"note", "Proceeding with encryption attempt - ensure key policy allows CloudWatch Logs service")
		return nil
	}

	policy := *policyResult.Policy
	policyContainsLogsService := s.checkCloudWatchLogsPolicyAccess(policy)

	// Log comprehensive audit information
	slog.Info("KMS key policy validation audit",
		"kms_key_id", keyId,
		"policy_accessible", true,
		"cloudwatch_logs_access_found", policyContainsLogsService,
		"validation_timestamp", time.Now().UTC().Format(time.RFC3339))

	if !policyContainsLogsService {
		slog.Warn("KMS key policy may not include CloudWatch Logs service access",
			"kms_key_id", keyId,
			"note", "Ensure the key policy allows the CloudWatch Logs service to use this key",
			"audit_action", AuditActionPolicyValidationWarning)
	} else {
		slog.Info("KMS key policy validation successful",
			"kms_key_id", keyId,
			"cloudwatch_logs_access", "confirmed",
			"audit_action", AuditActionPolicyValidationSuccess)
	}

	return nil
}

// associateKMSKeyWithRetry associates a KMS key with the log group with retry logic
func (s *ComplianceService) associateKMSKeyWithRetry(ctx context.Context, logGroupName, kmsKeyArn string) error {
	maxRetries := int(s.config.MaxKMSRetries)
	var lastErr error

	// Maximum delay cap to prevent excessive wait times (30 seconds)
	const maxDelay = 30 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with overflow protection
			// Use bit shifting for efficient power-of-2 calculations
			var multiplier int64
			if attempt < MaxExponentialBackoffAttempts {
				// Bit shifting is faster and more direct: 1<<attempt gives us 2^attempt
				multiplier = int64(1 << attempt)
			} else {
				// Cap at MaxBackoffMultiplier (2^10) to prevent excessive delays
				multiplier = MaxBackoffMultiplier
			}
			delay := time.Duration(multiplier) * s.config.RetryBaseDelay
			if delay > maxDelay {
				delay = maxDelay
			}
			slog.Info("Retrying KMS key association",
				"log_group", logGroupName,
				"kms_key_id", kmsKeyArn,
				"attempt", attempt+1,
				"delay", delay)
			time.Sleep(delay)
		}

		input := &cloudwatchlogs.AssociateKmsKeyInput{
			LogGroupName: aws.String(logGroupName),
			KmsKeyId:     aws.String(kmsKeyArn),
		}

		_, err := s.logsClient.AssociateKmsKey(ctx, input)
		if err == nil {
			slog.Info("Successfully associated KMS key",
				"log_group", logGroupName,
				"kms_key_id", kmsKeyArn,
				"attempts", attempt+1)
			return nil
		}

		lastErr = err

		// Check for specific errors that shouldn't be retried
		if isKMSKeyNotFoundError(err) || isKMSAccessDeniedError(err) || isInvalidLogGroupError(err) {
			slog.Error("Non-retryable error encountered",
				"log_group", logGroupName,
				"kms_key_id", kmsKeyArn,
				"error", err,
				"attempt", attempt+1)
			break
		}

		// Check for rate limiting errors
		if isRateLimitError(err) {
			slog.Warn("Rate limit encountered during KMS key association",
				"log_group", logGroupName,
				"kms_key_id", kmsKeyArn,
				"attempt", attempt+1,
				"error", err)
			continue
		}

		slog.Warn("KMS key association failed, will retry",
			"log_group", logGroupName,
			"kms_key_id", kmsKeyArn,
			"attempt", attempt+1,
			"error", err)
	}

	return fmt.Errorf("failed to associate KMS key after %d attempts: %w", maxRetries, lastErr)
}

// getCurrentRegion returns the current AWS region from the service configuration
func (s *ComplianceService) getCurrentRegion() string {
	return s.config.Region
}

// checkAPIErrorCode is a helper function to check if an error matches any of the provided error codes
func checkAPIErrorCode(err error, errorCodes []string) bool {
	if err == nil {
		return false
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		errorCode := apiErr.ErrorCode()
		for _, code := range errorCodes {
			if errorCode == code {
				return true
			}
		}
	}
	return false
}

// KMS-specific error checking functions using proper AWS SDK error types
func isKMSKeyNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific KMS exception types
	var notFoundErr *kmstypes.NotFoundException
	var keyUnavailableErr *kmstypes.KeyUnavailableException

	if errors.As(err, &notFoundErr) || errors.As(err, &keyUnavailableErr) {
		return true
	}

	// Check for generic API error with specific error codes
	return checkAPIErrorCode(err, []string{
		"NotFoundException",
		"InvalidKeyId.NotFound",
		"InvalidKeyId.Malformed",
		"KeyUnavailableException",
	})
}

func isKMSAccessDeniedError(err error) bool {
	if err == nil {
		return false
	}

	// Check for KMS access denied exception types using errors.As with smithy.APIError
	// This provides consistent error handling across different AWS SDK error types
	return checkAPIErrorCode(err, []string{
		"AccessDeniedException",
		"UnauthorizedOperation",
		"Forbidden",
		"AccessDenied",
	})
}

func isInvalidLogGroupError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific CloudWatch Logs exception types
	var resourceNotFoundErr *cloudwatchlogstypes.ResourceNotFoundException
	var invalidParameterErr *cloudwatchlogstypes.InvalidParameterException

	if errors.As(err, &resourceNotFoundErr) || errors.As(err, &invalidParameterErr) {
		return true
	}

	// Check for generic API error with specific error codes
	return checkAPIErrorCode(err, []string{
		"ResourceNotFoundException",
		"InvalidLogGroupName",
		"InvalidParameterValue",
	})
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

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
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

// convertToComplianceResultForRule converts a NonCompliantResource to ComplianceResult based on specific Config rule
func (s *ComplianceService) convertToComplianceResultForRule(configRuleName string, resource types.NonCompliantResource) types.ComplianceResult {
	result := types.ComplianceResult{
		LogGroupName: resource.ResourceName,
		Region:       resource.Region,
		AccountId:    resource.AccountId,
	}

	// Each Config rule evaluates ONLY its specific compliance requirement
	// This ensures each rule evaluates ALL resources for its requirement independently
	ruleType := s.ruleClassifier.ClassifyRule(configRuleName)

	switch ruleType {
	case types.RuleTypeEncryption:
		// Encryption rule: ONLY evaluate encryption compliance
		result.MissingEncryption = true // Resource is non-compliant for encryption
		result.MissingRetention = false // Not this rule's concern

		slog.Info("Encryption rule batch evaluation",
			"log_group", resource.ResourceName,
			"config_rule", configRuleName,
			"compliance_type", resource.ComplianceType,
			"rule_type", ruleType.String(),
			"audit_action", "encryption_batch_compliance_check")

	case types.RuleTypeRetention:
		// Retention rule: ONLY evaluate retention compliance
		result.MissingRetention = true   // Resource is non-compliant for retention
		result.MissingEncryption = false // Not this rule's concern

		slog.Info("Retention rule batch evaluation",
			"log_group", resource.ResourceName,
			"config_rule", configRuleName,
			"compliance_type", resource.ComplianceType,
			"rule_type", ruleType.String(),
			"audit_action", "retention_batch_compliance_check")

	default:
		// Unknown rule - log and skip
		slog.Warn("Unsupported Config rule in batch - no compliance evaluation performed",
			"config_rule", configRuleName,
			"log_group", resource.ResourceName,
			"rule_type", "unknown",
			"audit_action", "unsupported_rule_batch_skip")
		result.MissingEncryption = false
		result.MissingRetention = false
	}

	return result
}
