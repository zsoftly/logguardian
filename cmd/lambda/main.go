package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/zsoftly/logguardian/internal/handler"
	"github.com/zsoftly/logguardian/internal/service"
	"github.com/zsoftly/logguardian/internal/types"
)

func main() {
	// Set up structured logging with JSON output for Lambda
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Initialize Lambda handler
	h, err := initializeHandler(context.TODO())
	if err != nil {
		slog.Error("Failed to initialize handler", "error", err)
		panic(err)
	}

	// Start Lambda with unified handler
	lambda.Start(func(ctx context.Context, request types.LambdaRequest) error {
		return handleUnifiedRequest(ctx, h, request)
	})
}

// initializeHandler creates and returns a ComplianceHandler with AWS configuration.
// This function is extracted from main() to allow testing of initialization logic
// without requiring the Lambda runtime. This is the only production code change made
// for the unit test suite implementation - it's a safe refactoring that improves
// testability without changing behavior.
func initializeHandler(ctx context.Context) (*handler.ComplianceHandler, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create services
	complianceService := service.NewComplianceService(cfg)

	// Create handler
	h := handler.NewComplianceHandler(complianceService)

	return h, nil
}

// handleUnifiedRequest routes requests to the appropriate handler based on request type
func handleUnifiedRequest(ctx context.Context, h *handler.ComplianceHandler, request types.LambdaRequest) error {
	slog.Info("Received Lambda request", "type", request.Type)

	switch request.Type {
	case "config-event":
		// Handle individual Config rule evaluation events
		if request.ConfigEvent == nil {
			return fmt.Errorf("configEvent is required for type 'config-event'")
		}
		return h.HandleConfigEvent(ctx, request.ConfigEvent)

	case "config-rule-evaluation":
		// Handle batch Config rule evaluation requests
		if request.ConfigRuleName == "" {
			return fmt.Errorf("configRuleName is required for type 'config-rule-evaluation'")
		}
		if request.Region == "" {
			return fmt.Errorf("region is required for type 'config-rule-evaluation'")
		}

		batchSize := request.BatchSize
		if batchSize <= 0 {
			batchSize = 10 // Default batch size
		}

		return h.HandleConfigRuleEvaluationRequest(ctx, request.ConfigRuleName, request.Region, batchSize)

	default:
		return fmt.Errorf("unsupported request type: %s (supported types: 'config-event', 'config-rule-evaluation')", request.Type)
	}
}
