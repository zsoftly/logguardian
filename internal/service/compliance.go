package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/zsoftly/logguardian/internal/types"
)

// ComplianceService handles log group compliance remediation
type ComplianceService struct {
	logsClient CloudWatchLogsClientInterface
	kmsClient  KMSClientInterface
	config     ServiceConfig
}

// ServiceConfig holds configuration for the compliance service
type ServiceConfig struct {
	DefaultKMSKeyAlias   string
	DefaultRetentionDays int32
	DryRun               bool
}

// NewComplianceService creates a new compliance service
func NewComplianceService(cfg aws.Config) *ComplianceService {
	// Load configuration from environment variables
	config := ServiceConfig{
		DefaultKMSKeyAlias:   getEnvOrDefault("KMS_KEY_ALIAS", "alias/cloudwatch-logs-compliance"),
		DefaultRetentionDays: getEnvAsInt32OrDefault("DEFAULT_RETENTION_DAYS", 365),
		DryRun:               getEnvAsBoolOrDefault("DRY_RUN", false),
	}

	return &ComplianceService{
		logsClient: cloudwatchlogs.NewFromConfig(cfg),
		kmsClient:  kms.NewFromConfig(cfg),
		config:     config,
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
