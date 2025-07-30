package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

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
		logsClient: logsClient,
		kmsClient:  kmsClient,
		config:     serviceConfig,
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
			return fmt.Errorf("failed to access CloudWatch Logs in region %s: %w", region, err)
		}

		// Test KMS access by attempting to resolve the key alias
		_, err = service.resolveKMSKeyAlias(ctx, service.config.DefaultKMSKeyAlias)
		if err != nil {
			slog.Warn("KMS key validation failed",
				"region", region,
				"key_alias", service.config.DefaultKMSKeyAlias,
				"error", err)
			// Don't fail validation if KMS key doesn't exist - it might be created later
		}

		slog.Info("Region validation passed", "region", region)
	}

	return nil
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
	regionsEnv := getEnvOrDefault("SUPPORTED_REGIONS", "us-east-1,us-west-2")
	regions := parseCommaDelimitedString(regionsEnv)

	if len(regions) == 0 {
		regions = []string{"us-east-1"} // Default fallback
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
