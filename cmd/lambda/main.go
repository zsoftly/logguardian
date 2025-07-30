package main

import (
	"context"
	"encoding/json"
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

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		slog.Error("Failed to load AWS config", "error", err)
		panic(err)
	}

	// Create services
	complianceService := service.NewComplianceService(cfg)

	// Create handler
	h := handler.NewComplianceHandler(complianceService)

	// Start Lambda with unified handler
	lambda.Start(func(ctx context.Context, request types.LambdaRequest) error {
		return handleUnifiedRequest(ctx, h, request)
	})
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

	case "":
		// For backward compatibility, try to parse as a direct Config event
		slog.Info("No type specified, attempting to parse as Config event for backward compatibility")

		// Try to marshal the entire request as a Config event
		eventBytes, err := json.Marshal(request)
		if err != nil {
			return fmt.Errorf("failed to parse request as Config event: %w", err)
		}

		// Check if it looks like a Config event by trying to unmarshal it
		var testEvent types.ConfigEvent
		if err := json.Unmarshal(eventBytes, &testEvent); err == nil && testEvent.ConfigRuleName != "" {
			return h.HandleConfigEvent(ctx, eventBytes)
		}

		return fmt.Errorf("unable to determine request type - please specify 'type' field")

	default:
		return fmt.Errorf("unsupported request type: %s", request.Type)
	}
}
