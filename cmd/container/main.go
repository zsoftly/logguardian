package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/zsoftly/logguardian/internal/container"
)

const (
	ExitSuccess = 0
	ExitError   = 1
	ExitUsage   = 2
)

type CommandInput struct {
	Type           string
	ConfigRuleName string
	Region         string
	BatchSize      int
	DryRun         bool
	Profile        string
	AssumeRole     string
	Verbose        bool
	OutputFormat   string
}

func main() {
	input := parseCommandLineArgs()

	logLevel := slog.LevelInfo
	if input.Verbose {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	executionID := fmt.Sprintf("exec-%d", time.Now().Unix())
	startTime := time.Now()
	slog.Info("Starting LogGuardian container execution",
		"execution_id", executionID,
		"version", getVersion(),
		"mode", getExecutionMode(input.DryRun))

	ctx := context.Background()
	exitCode := execute(ctx, input, executionID)

	slog.Info("Execution completed",
		"execution_id", executionID,
		"exit_code", exitCode,
		"duration", time.Since(startTime).String())

	os.Exit(exitCode)
}

func parseCommandLineArgs() CommandInput {
	input := CommandInput{}

	flag.StringVar(&input.Type, "type", "config-rule-evaluation", "Request type: config-rule-evaluation")
	flag.StringVar(&input.ConfigRuleName, "config-rule", "", "AWS Config rule name to evaluate")
	flag.StringVar(&input.Region, "region", os.Getenv("AWS_REGION"), "AWS region")
	flag.IntVar(&input.BatchSize, "batch-size", 10, "Batch size for processing resources")
	flag.BoolVar(&input.DryRun, "dry-run", false, "Preview changes without applying them")
	flag.StringVar(&input.Profile, "profile", os.Getenv("AWS_PROFILE"), "AWS profile to use")
	flag.StringVar(&input.AssumeRole, "assume-role", os.Getenv("AWS_ASSUME_ROLE_ARN"), "IAM role ARN to assume")
	flag.BoolVar(&input.Verbose, "verbose", false, "Enable verbose logging")
	flag.StringVar(&input.OutputFormat, "output", "json", "Output format: json or text")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "LogGuardian Container - AWS Config Compliance Automation\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nEvaluate and remediate AWS Config compliance for CloudWatch Log Groups.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  AWS_REGION              Default AWS region\n")
		fmt.Fprintf(os.Stderr, "  AWS_PROFILE             AWS profile to use\n")
		fmt.Fprintf(os.Stderr, "  AWS_ASSUME_ROLE_ARN     IAM role ARN to assume\n")
		fmt.Fprintf(os.Stderr, "  CONFIG_RULE_NAME        Config rule name (alternative to --config-rule)\n")
		fmt.Fprintf(os.Stderr, "  BATCH_SIZE              Batch size for processing\n")
		fmt.Fprintf(os.Stderr, "  DRY_RUN                 Set to 'true' for dry-run mode\n")
	}

	flag.Parse()

	// Check environment variables as fallback
	if input.ConfigRuleName == "" {
		input.ConfigRuleName = os.Getenv("CONFIG_RULE_NAME")
	}
	if input.Region == "" {
		input.Region = os.Getenv("AWS_DEFAULT_REGION")
	}
	if envBatchSize := os.Getenv("BATCH_SIZE"); envBatchSize != "" {
		var batchSize int
		if _, err := fmt.Sscanf(envBatchSize, "%d", &batchSize); err == nil && batchSize > 0 {
			input.BatchSize = batchSize
		}
	}
	if strings.ToLower(os.Getenv("DRY_RUN")) == "true" {
		input.DryRun = true
	}

	return input
}

func execute(ctx context.Context, input CommandInput, executionID string) int {
	if err := validateInput(input); err != nil {
		slog.Error("Invalid input", "error", err, "execution_id", executionID)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		return ExitUsage
	}

	// Create AWS config with authentication strategy
	awsCfg, err := createAWSConfig(ctx, input)
	if err != nil {
		slog.Error("Failed to create AWS config", "error", err, "execution_id", executionID)
		outputError(input.OutputFormat, executionID, "Authentication failed", err)
		return ExitError
	}

	// Create the command processor
	processor := container.NewCommandProcessor(awsCfg, container.ProcessorOptions{
		DryRun:       input.DryRun,
		ExecutionID:  executionID,
		OutputFormat: input.OutputFormat,
	})

	// Execute the command
	result, err := processor.Execute(ctx, container.CommandRequest{
		Type:           input.Type,
		ConfigRuleName: input.ConfigRuleName,
		Region:         input.Region,
		BatchSize:      input.BatchSize,
	})

	if err != nil {
		slog.Error("Command execution failed", "error", err, "execution_id", executionID)
		outputError(input.OutputFormat, executionID, "Execution failed", err)
		return ExitError
	}

	// Output the result
	if err := outputResult(input.OutputFormat, result); err != nil {
		slog.Error("Failed to output result", "error", err, "execution_id", executionID)
		return ExitError
	}

	return ExitSuccess
}

func validateInput(input CommandInput) error {
	if input.Type != "config-rule-evaluation" {
		return fmt.Errorf("unsupported request type: %s", input.Type)
	}

	if input.ConfigRuleName == "" {
		return fmt.Errorf("config rule name is required (use --config-rule or CONFIG_RULE_NAME env var)")
	}

	if input.Region == "" {
		return fmt.Errorf("region is required (use --region, AWS_REGION, or AWS_DEFAULT_REGION env var)")
	}

	if input.BatchSize <= 0 || input.BatchSize > 100 {
		return fmt.Errorf("batch size must be between 1 and 100")
	}

	return nil
}

func createAWSConfig(ctx context.Context, input CommandInput) (aws.Config, error) {
	authStrategy := container.NewAuthenticationStrategy()

	options := container.AuthOptions{
		Profile:    input.Profile,
		AssumeRole: input.AssumeRole,
		Region:     input.Region,
	}

	return authStrategy.GetAWSConfig(ctx, options)
}

func outputResult(format string, result *container.ExecutionResult) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	case "text":
		fmt.Printf("Execution ID: %s\n", result.ExecutionID)
		fmt.Printf("Status: %s\n", result.Status)
		fmt.Printf("Mode: %s\n", result.Mode)
		fmt.Printf("Config Rule: %s\n", result.ConfigRuleName)
		fmt.Printf("Region: %s\n", result.Region)
		fmt.Printf("Total Processed: %d\n", result.TotalProcessed)
		fmt.Printf("Success Count: %d\n", result.SuccessCount)
		fmt.Printf("Failure Count: %d\n", result.FailureCount)
		fmt.Printf("Duration: %s\n", result.Duration)
		if result.DryRunSummary != nil {
			fmt.Printf("\nDry Run Summary:\n")
			fmt.Printf("  Would Apply Encryption: %d\n", result.DryRunSummary.WouldApplyEncryption)
			fmt.Printf("  Would Apply Retention: %d\n", result.DryRunSummary.WouldApplyRetention)
			fmt.Printf("  Already Compliant: %d\n", result.DryRunSummary.AlreadyCompliant)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func outputError(format, executionID, message string, err error) {
	result := &container.ExecutionResult{
		ExecutionID: executionID,
		Status:      "failed",
		Error:       fmt.Sprintf("%s: %v", message, err),
		Timestamp:   time.Now(),
	}

	switch format {
	case "json":
		encoder := json.NewEncoder(os.Stderr)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(result)
	case "text":
		fmt.Fprintf(os.Stderr, "Execution ID: %s\n", executionID)
		fmt.Fprintf(os.Stderr, "Status: failed\n")
		fmt.Fprintf(os.Stderr, "Error: %s: %v\n", message, err)
	}
}

func getVersion() string {
	if version := os.Getenv("APP_VERSION"); version != "" {
		return version
	}
	return "unknown"
}

func getExecutionMode(dryRun bool) string {
	if dryRun {
		return "dry-run"
	}
	return "apply"
}
