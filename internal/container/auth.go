package container

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type AuthenticationStrategy struct {
	strategies []AuthStrategy
}

type AuthStrategy interface {
	Name() string
	IsAvailable(ctx context.Context, options AuthOptions) bool
	GetConfig(ctx context.Context, options AuthOptions) (aws.Config, error)
	Priority() int
}

type AuthOptions struct {
	Profile    string
	AssumeRole string
	Region     string
}

func NewAuthenticationStrategy() *AuthenticationStrategy {
	return &AuthenticationStrategy{
		strategies: []AuthStrategy{
			&ExplicitCredentialsStrategy{},
			&ProfileStrategy{},
			&AssumeRoleStrategy{},
			&TaskRoleStrategy{},
			&EnvironmentStrategy{},
			&InstanceProfileStrategy{},
			&DefaultStrategy{},
		},
	}
}

func (a *AuthenticationStrategy) GetAWSConfig(ctx context.Context, options AuthOptions) (aws.Config, error) {
	slog.Info("Resolving AWS credentials",
		"profile", options.Profile,
		"assume_role", options.AssumeRole,
		"region", options.Region)

	// Try strategies in order of priority
	for _, strategy := range a.strategies {
		if strategy.IsAvailable(ctx, options) {
			slog.Info("Attempting authentication strategy",
				"strategy", strategy.Name(),
				"priority", strategy.Priority())

			cfg, err := strategy.GetConfig(ctx, options)
			if err == nil {
				slog.Info("Successfully authenticated",
					"strategy", strategy.Name(),
					"region", options.Region)
				return cfg, nil
			}

			slog.Warn("Strategy failed",
				"strategy", strategy.Name(),
				"error", err)
		}
	}

	return aws.Config{}, fmt.Errorf("no valid authentication method found")
}

// ExplicitCredentialsStrategy handles explicitly provided credentials
type ExplicitCredentialsStrategy struct{}

func (s *ExplicitCredentialsStrategy) Name() string  { return "explicit-credentials" }
func (s *ExplicitCredentialsStrategy) Priority() int { return 1 }

func (s *ExplicitCredentialsStrategy) IsAvailable(ctx context.Context, options AuthOptions) bool {
	// Check if explicit credentials are provided via environment
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	return accessKey != "" && secretKey != ""
}

func (s *ExplicitCredentialsStrategy) GetConfig(ctx context.Context, options AuthOptions) (aws.Config, error) {
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	sessionToken := os.Getenv("AWS_SESSION_TOKEN")

	opts := []func(*config.LoadOptions) error{
		config.WithRegion(options.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, sessionToken),
		),
	}

	return config.LoadDefaultConfig(ctx, opts...)
}

// ProfileStrategy handles AWS profile-based authentication
type ProfileStrategy struct{}

func (s *ProfileStrategy) Name() string  { return "profile" }
func (s *ProfileStrategy) Priority() int { return 2 }

func (s *ProfileStrategy) IsAvailable(ctx context.Context, options AuthOptions) bool {
	return options.Profile != ""
}

func (s *ProfileStrategy) GetConfig(ctx context.Context, options AuthOptions) (aws.Config, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(options.Region),
		config.WithSharedConfigProfile(options.Profile),
	}

	return config.LoadDefaultConfig(ctx, opts...)
}

// AssumeRoleStrategy handles IAM role assumption
type AssumeRoleStrategy struct{}

func (s *AssumeRoleStrategy) Name() string  { return "assume-role" }
func (s *AssumeRoleStrategy) Priority() int { return 3 }

func (s *AssumeRoleStrategy) IsAvailable(ctx context.Context, options AuthOptions) bool {
	return options.AssumeRole != ""
}

func (s *AssumeRoleStrategy) GetConfig(ctx context.Context, options AuthOptions) (aws.Config, error) {
	// First get base config
	baseCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(options.Region))
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load base config: %w", err)
	}

	// Create STS client
	stsClient := sts.NewFromConfig(baseCfg)

	// Create assume role provider
	roleProvider := stscreds.NewAssumeRoleProvider(stsClient, options.AssumeRole,
		func(o *stscreds.AssumeRoleOptions) {
			o.RoleSessionName = "logguardian-container"
		})

	// Return config with assumed role credentials
	return config.LoadDefaultConfig(ctx,
		config.WithRegion(options.Region),
		config.WithCredentialsProvider(roleProvider),
	)
}

// TaskRoleStrategy handles ECS/Fargate task role authentication
type TaskRoleStrategy struct{}

func (s *TaskRoleStrategy) Name() string  { return "ecs-task-role" }
func (s *TaskRoleStrategy) Priority() int { return 4 }

func (s *TaskRoleStrategy) IsAvailable(ctx context.Context, options AuthOptions) bool {
	// Check if running in ECS by looking for ECS metadata URI
	metadataURI := os.Getenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI")
	if metadataURI != "" {
		return true
	}

	// Check for ECS full URI (Fargate)
	fullURI := os.Getenv("AWS_CONTAINER_CREDENTIALS_FULL_URI")
	return fullURI != ""
}

func (s *TaskRoleStrategy) GetConfig(ctx context.Context, options AuthOptions) (aws.Config, error) {
	// AWS SDK automatically handles ECS task role credentials
	return config.LoadDefaultConfig(ctx, config.WithRegion(options.Region))
}

// EnvironmentStrategy handles environment variable authentication
type EnvironmentStrategy struct{}

func (s *EnvironmentStrategy) Name() string  { return "environment" }
func (s *EnvironmentStrategy) Priority() int { return 5 }

func (s *EnvironmentStrategy) IsAvailable(ctx context.Context, options AuthOptions) bool {
	// Check if AWS credentials are in environment (but not explicit)
	// This catches cases where credentials are set but not via AWS_ACCESS_KEY_ID
	_, hasProfile := os.LookupEnv("AWS_PROFILE")
	return hasProfile && options.Profile == ""
}

func (s *EnvironmentStrategy) GetConfig(ctx context.Context, options AuthOptions) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx, config.WithRegion(options.Region))
}

// InstanceProfileStrategy handles EC2 instance profile authentication
type InstanceProfileStrategy struct{}

func (s *InstanceProfileStrategy) Name() string  { return "ec2-instance-profile" }
func (s *InstanceProfileStrategy) Priority() int { return 6 }

func (s *InstanceProfileStrategy) IsAvailable(ctx context.Context, options AuthOptions) bool {
	// Try to detect if running on EC2
	// This is a heuristic - the SDK will handle the actual detection
	hostname, _ := os.Hostname()
	return hostname != "" && isEC2Instance()
}

func (s *InstanceProfileStrategy) GetConfig(ctx context.Context, options AuthOptions) (aws.Config, error) {
	// AWS SDK automatically handles EC2 instance profile credentials
	return config.LoadDefaultConfig(ctx, config.WithRegion(options.Region))
}

// DefaultStrategy is the fallback authentication method
type DefaultStrategy struct{}

func (s *DefaultStrategy) Name() string  { return "default" }
func (s *DefaultStrategy) Priority() int { return 99 }

func (s *DefaultStrategy) IsAvailable(ctx context.Context, options AuthOptions) bool {
	// Always available as last resort
	return true
}

func (s *DefaultStrategy) GetConfig(ctx context.Context, options AuthOptions) (aws.Config, error) {
	// Try default credential chain
	return config.LoadDefaultConfig(ctx, config.WithRegion(options.Region))
}

// isEC2Instance attempts to detect if running on EC2
func isEC2Instance() bool {
	// Check for EC2 metadata service
	// This is a simple heuristic - the SDK does more robust detection
	if _, err := os.Stat("/sys/devices/virtual/dmi/id/product_uuid"); err == nil {
		return true
	}

	// Check for AWS_EXECUTION_ENV
	if env := os.Getenv("AWS_EXECUTION_ENV"); env != "" {
		return true
	}

	// Check IMDSv2 endpoint availability would be done by SDK
	return false
}
