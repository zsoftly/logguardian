package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

// MetricsService handles CloudWatch metrics publishing
type MetricsService struct {
	cloudwatchClient *cloudwatch.Client
	environment      string
	region           string
	namespace        string
}

// NewMetricsService creates a new metrics service
func NewMetricsService(cfg aws.Config) *MetricsService {
	return &MetricsService{
		cloudwatchClient: cloudwatch.NewFromConfig(cfg),
		environment:      getEnvOrDefault("ENVIRONMENT", "unknown"),
		region:           cfg.Region,
		namespace:        "LogGuardian",
	}
}

// MetricsData holds metrics for batch publishing
type MetricsData struct {
	LogGroupsProcessed int
	LogGroupsRemediated int
	RemediationErrors   int
}

// PublishBatchMetrics publishes all metrics from a batch operation
func (m *MetricsService) PublishBatchMetrics(ctx context.Context, metrics MetricsData) error {
	timestamp := time.Now()
	
	var metricData []types.MetricDatum

	// Add LogGroupsProcessed metric
	if metrics.LogGroupsProcessed > 0 {
		metricData = append(metricData, types.MetricDatum{
			MetricName: aws.String("LogGroupsProcessed"),
			Value:      aws.Float64(float64(metrics.LogGroupsProcessed)),
			Unit:       types.StandardUnitCount,
			Timestamp:  &timestamp,
			Dimensions: []types.Dimension{
				{
					Name:  aws.String("Environment"),
					Value: aws.String(m.environment),
				},
			},
		})
	}

	// Add LogGroupsRemediated metric
	if metrics.LogGroupsRemediated > 0 {
		metricData = append(metricData, types.MetricDatum{
			MetricName: aws.String("LogGroupsRemediated"),
			Value:      aws.Float64(float64(metrics.LogGroupsRemediated)),
			Unit:       types.StandardUnitCount,
			Timestamp:  &timestamp,
			Dimensions: []types.Dimension{
				{
					Name:  aws.String("Environment"),
					Value: aws.String(m.environment),
				},
			},
		})
	}

	// Add RemediationErrors metric
	if metrics.RemediationErrors > 0 {
		metricData = append(metricData, types.MetricDatum{
			MetricName: aws.String("RemediationErrors"),
			Value:      aws.Float64(float64(metrics.RemediationErrors)),
			Unit:       types.StandardUnitCount,
			Timestamp:  &timestamp,
			Dimensions: []types.Dimension{
				{
					Name:  aws.String("Environment"),
					Value: aws.String(m.environment),
				},
			},
		})
	}

	// Publish metrics if we have any
	if len(metricData) > 0 {
		input := &cloudwatch.PutMetricDataInput{
			Namespace:  aws.String(m.namespace),
			MetricData: metricData,
		}

		_, err := m.cloudwatchClient.PutMetricData(ctx, input)
		if err != nil {
			slog.Error("Failed to publish CloudWatch metrics",
				"namespace", m.namespace,
				"environment", m.environment,
				"metrics_count", len(metricData),
				"error", err)
			return err
		}

		slog.Info("Successfully published CloudWatch metrics",
			"namespace", m.namespace,
			"environment", m.environment,
			"metrics_published", len(metricData),
			"processed", metrics.LogGroupsProcessed,
			"remediated", metrics.LogGroupsRemediated,
			"errors", metrics.RemediationErrors)
	}

	return nil
}

// PublishSingleMetric publishes a single metric
func (m *MetricsService) PublishSingleMetric(ctx context.Context, metricName string, value float64, unit types.StandardUnit) error {
	timestamp := time.Now()
	
	metricData := []types.MetricDatum{
		{
			MetricName: aws.String(metricName),
			Value:      aws.Float64(value),
			Unit:       unit,
			Timestamp:  &timestamp,
			Dimensions: []types.Dimension{
				{
					Name:  aws.String("Environment"),
					Value: aws.String(m.environment),
				},
			},
		},
	}

	input := &cloudwatch.PutMetricDataInput{
		Namespace:  aws.String(m.namespace),
		MetricData: metricData,
	}

	_, err := m.cloudwatchClient.PutMetricData(ctx, input)
	if err != nil {
		slog.Error("Failed to publish single CloudWatch metric",
			"namespace", m.namespace,
			"metric_name", metricName,
			"value", value,
			"error", err)
		return err
	}

	slog.Debug("Published single CloudWatch metric",
		"namespace", m.namespace,
		"metric_name", metricName,
		"value", value,
		"environment", m.environment)

	return nil
}

