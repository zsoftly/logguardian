package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/zsoftly/logguardian/internal/handler"
	"github.com/zsoftly/logguardian/internal/service"
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

	// Start Lambda
	lambda.Start(h.HandleConfigEvent)
}