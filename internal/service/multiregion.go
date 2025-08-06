package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/zsoftly/logguardian/internal/types"
)

// MultiRegionComplianceService handles compliance across multiple AWS regions
type MultiRegionComplianceService struct {
	baseConfig     aws.Config
	serviceConfigs map[string]ServiceConfig      // region -> config
	services       map[string]*ComplianceService // region -> service
	mu             sync.RWMutex
}

// validationResult represents the result of KMS key validation for a specific region
type validationResult struct {
	region string
	report *types.KMSValidationReport
}

// regionJob represents a KMS validation job for a specific region
type regionJob struct {
	region  string
	service *ComplianceService
}

// NewMultiRegionComplianceService creates a new multi-region compliance service
func NewMultiRegionComplianceService(baseConfig aws.Config) *MultiRegionComplianceService {
	return &MultiRegionComplianceService{
		baseConfig:     baseConfig,
		serviceConfigs: make(map[string]ServiceConfig),
		services:       make(map[string]*ComplianceService),
	}
}

// AddRegion adds support for a specific region with custom configuration
func (mrs *MultiRegionComplianceService) AddRegion(region string, serviceConfig ServiceConfig) error {
	mrs.mu.Lock()
	defer mrs.mu.Unlock()

	// Create region-specific AWS config
	regionConfig := mrs.baseConfig.Copy()
	regionConfig.Region = region

	// Create CloudWatch Logs and KMS clients for this region
	logsClient := cloudwatchlogs.NewFromConfig(regionConfig)
	kmsClient := kms.NewFromConfig(regionConfig)

	// Create compliance service for this region
	service := &ComplianceService{
		logsClient:     logsClient,
		kmsClient:      kmsClient,
		ruleClassifier: types.NewRuleClassifier(), // Initialize rule classifier
		config:         serviceConfig,
	}

	mrs.serviceConfigs[region] = serviceConfig
	mrs.services[region] = service

	slog.Info("Added region support",
		"region", region,
		"kms_key_alias", serviceConfig.DefaultKMSKeyAlias,
		"retention_days", serviceConfig.DefaultRetentionDays)

	return nil
}

// RemediateLogGroup applies remediation to a log group in the appropriate region
func (mrs *MultiRegionComplianceService) RemediateLogGroup(ctx context.Context, compliance types.ComplianceResult) (*types.RemediationResult, error) {
	mrs.mu.RLock()
	service, exists := mrs.services[compliance.Region]
	mrs.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no service configured for region %s", compliance.Region)
	}

	slog.Info("Using region-specific service",
		"region", compliance.Region,
		"log_group", compliance.LogGroupName)

	return service.RemediateLogGroup(ctx, compliance)
}

// GetSupportedRegions returns the list of configured regions
func (mrs *MultiRegionComplianceService) GetSupportedRegions() []string {
	mrs.mu.RLock()
	defer mrs.mu.RUnlock()

	regions := make([]string, 0, len(mrs.services))
	for region := range mrs.services {
		regions = append(regions, region)
	}
	return regions
}

// LoadRegionsFromConfig loads multiple regions from environment configuration
func (mrs *MultiRegionComplianceService) LoadRegionsFromConfig(ctx context.Context, regions []string) error {
	for _, region := range regions {
		// Create region-specific configuration
		// You could customize this per region by reading region-specific env vars
		serviceConfig := ServiceConfig{
			DefaultKMSKeyAlias:   getEnvOrDefault(fmt.Sprintf("KMS_KEY_ALIAS_%s", region), getEnvOrDefault("KMS_KEY_ALIAS", "alias/cloudwatch-logs-compliance")),
			DefaultRetentionDays: getEnvAsInt32OrDefault(fmt.Sprintf("DEFAULT_RETENTION_DAYS_%s", region), getEnvAsInt32OrDefault("DEFAULT_RETENTION_DAYS", 365)),
			DryRun:               getEnvAsBoolOrDefault("DRY_RUN", false),
		}

		if err := mrs.AddRegion(region, serviceConfig); err != nil {
			return fmt.Errorf("failed to add region %s: %w", region, err)
		}
	}

	slog.Info("Loaded multi-region configuration", "regions", regions)
	return nil
}

// ValidateRegionAccess validates that we can access required services in each region
func (mrs *MultiRegionComplianceService) ValidateRegionAccess(ctx context.Context) error {
	mrs.mu.RLock()
	defer mrs.mu.RUnlock()

	for region, service := range mrs.services {
		slog.Info("Validating region access", "region", region)

		// Test CloudWatch Logs access by listing log groups (limit 1)
		logsInput := &cloudwatchlogs.DescribeLogGroupsInput{
			Limit: aws.Int32(1),
		}
		_, err := service.logsClient.DescribeLogGroups(ctx, logsInput)
		if err != nil {
			slog.Error("Failed to access CloudWatch Logs in region",
				"region", region,
				"error", err,
				"audit_action", "region_validation_failed",
				"service", "cloudwatch_logs")
			return fmt.Errorf("failed to access CloudWatch Logs in region %s: %w", region, err)
		}

		// Test KMS access by validating the key alias
		keyInfo, err := service.validateKMSKeyAccessibility(ctx, service.config.DefaultKMSKeyAlias)
		if err != nil {
			slog.Warn("KMS key validation failed during region validation",
				"region", region,
				"key_alias", service.config.DefaultKMSKeyAlias,
				"error", err,
				"audit_action", "region_validation_warning",
				"service", "kms",
				"note", "KMS key may not exist yet - this is acceptable if keys will be created later")
			// Don't fail validation if KMS key doesn't exist - it might be created later
		} else {
			// If key exists, perform comprehensive validation
			slog.Info("KMS key validation successful during region access check",
				"region", region,
				"key_alias", service.config.DefaultKMSKeyAlias,
				"kms_key_id", keyInfo.KeyId,
				"kms_key_arn", keyInfo.Arn,
				"key_state", keyInfo.KeyState,
				"key_region", keyInfo.Region,
				"audit_action", "region_kms_validation_success")

			// Also validate the key policy for CloudWatch Logs
			if err := service.validateKMSKeyPolicyForCloudWatchLogs(ctx, keyInfo.KeyId); err != nil {
				slog.Warn("KMS key policy validation failed during region validation",
					"region", region,
					"kms_key_id", keyInfo.KeyId,
					"error", err,
					"audit_action", "region_policy_validation_warning",
					"note", "Key policy may need adjustment for CloudWatch Logs access")
			}
		}

		slog.Info("Region validation passed",
			"region", region,
			"audit_action", "region_validation_success")
	}

	return nil
}

// ValidateKMSKeysAcrossRegions validates KMS keys in all configured regions
// This provides comprehensive cross-region KMS key validation with concurrent processing
func (mrs *MultiRegionComplianceService) ValidateKMSKeysAcrossRegions(ctx context.Context) (map[string]*types.KMSValidationReport, error) {
	mrs.mu.RLock()
	defer mrs.mu.RUnlock()

	reports := make(map[string]*types.KMSValidationReport)
	var mu sync.Mutex
	var wg sync.WaitGroup

	slog.Info("Starting cross-region KMS key validation",
		"regions", len(mrs.services),
		"audit_action", "multi_region_kms_validation_start")

	// Use worker pool pattern to control concurrency and prevent resource exhaustion
	// Default to 10 workers based on AWS API rate limiting considerations:
	// - AWS APIs typically allow 10-20 requests per second per service
	// - Each region validation involves multiple API calls (KMS DescribeKey, GetKeyPolicy, etc.)
	// - This balances performance with avoiding throttling across multiple AWS services
	maxWorkers := getEnvAsIntOrDefault("MAX_REGION_WORKERS", 10)

	// Create channels for work distribution
	jobChan := make(chan regionJob, len(mrs.services))
	resultChan := make(chan validationResult, len(mrs.services))

	// Start worker goroutines with controlled concurrency
	numWorkers := maxWorkers
	if numWorkers > len(mrs.services) {
		numWorkers = len(mrs.services)
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobChan {
				slog.Info("Validating KMS key in region",
					"region", job.region,
					"key_alias", job.service.config.DefaultKMSKeyAlias)

				report, err := job.service.ValidateKMSKeyComprehensively(ctx, job.service.config.DefaultKMSKeyAlias)
				if err != nil {
					slog.Error("Failed to validate KMS key in region",
						"region", job.region,
						"key_alias", job.service.config.DefaultKMSKeyAlias,
						"error", err,
						"audit_action", "region_kms_validation_error")

					// Create a minimal report with error information
					report = &types.KMSValidationReport{
						KeyAlias:            job.service.config.DefaultKMSKeyAlias,
						CurrentRegion:       job.region,
						ValidationTimestamp: time.Now().UTC(),
						KeyExists:           false,
						KeyAccessible:       false,
						ValidationErrors:    []string{err.Error()},
					}
				}

				resultChan <- validationResult{
					region: job.region,
					report: report,
				}

				slog.Info("Completed KMS key validation for region",
					"region", job.region,
					"key_alias", job.service.config.DefaultKMSKeyAlias,
					"key_exists", report.KeyExists,
					"key_accessible", report.KeyAccessible,
					"cloudwatch_logs_access", report.CloudWatchLogsAccess,
					"validation_errors", len(report.ValidationErrors),
					"validation_warnings", len(report.ValidationWarnings))
			}
		}()
	}

	// Send all jobs to workers
	for region, service := range mrs.services {
		jobChan <- regionJob{region, service}
	}
	close(jobChan)

	// Collect results
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Gather all reports with thread-safe access
	for result := range resultChan {
		mu.Lock()
		reports[result.region] = result.report
		mu.Unlock()
	}

	// Log summary
	totalRegions := len(reports)
	successfulRegions := 0
	warningRegions := 0
	errorRegions := 0

	for _, report := range reports {
		if len(report.ValidationErrors) == 0 {
			if len(report.ValidationWarnings) == 0 {
				successfulRegions++
			} else {
				warningRegions++
			}
		} else {
			errorRegions++
		}
	}

	slog.Info("Cross-region KMS key validation summary",
		"total_regions", totalRegions,
		"successful_regions", successfulRegions,
		"warning_regions", warningRegions,
		"error_regions", errorRegions,
		"audit_action", "multi_region_kms_validation_complete")

	return reports, nil
}

// NewMultiRegionFromEnvironment creates a multi-region service from environment variables
func NewMultiRegionFromEnvironment(ctx context.Context) (*MultiRegionComplianceService, error) {
	// Load base AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create multi-region service
	mrs := NewMultiRegionComplianceService(cfg)

	// Load regions from environment (comma-separated list)
	regionsEnv := getEnvOrDefault("SUPPORTED_REGIONS", "ca-central-1,ca-west-1")
	regions := parseCommaDelimitedString(regionsEnv)

	if len(regions) == 0 {
		regions = []string{"ca-central-1"} // Default fallback
	}

	// Load configuration for each region
	if err := mrs.LoadRegionsFromConfig(ctx, regions); err != nil {
		return nil, fmt.Errorf("failed to load regions configuration: %w", err)
	}

	return mrs, nil
}

// parseCommaDelimitedString splits a comma-delimited string into a slice
func parseCommaDelimitedString(s string) []string {
	if s == "" {
		return nil
	}

	var result []string
	for _, item := range strings.Split(s, ",") {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
