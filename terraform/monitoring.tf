# ============================================
# CloudWatch Dashboard
# ============================================

resource "aws_cloudwatch_dashboard" "main" {
  count = var.create_monitoring_dashboard ? 1 : 0

  dashboard_name = local.dashboard_name

  dashboard_body = jsonencode({
    widgets = [
      {
        type = "metric"
        properties = {
          metrics = [
            ["AWS/Lambda", "Invocations", { stat = "Sum", label = "Lambda Invocations" }],
            [".", "Errors", { stat = "Sum", label = "Lambda Errors" }],
            [".", "Throttles", { stat = "Sum", label = "Lambda Throttles" }],
            [".", "Duration", { stat = "Average", label = "Avg Duration (ms)" }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = local.region
          title   = "Lambda Performance Metrics"
          period  = 300
          dimensions = {
            FunctionName = [aws_lambda_function.compliance.function_name]
          }
        }
      },
      {
        type = "metric"
        properties = {
          metrics = [
            ["LogGuardian", "LogGroupsProcessed", { stat = "Sum" }],
            [".", "RemediationSuccess", { stat = "Sum" }],
            [".", "RemediationFailure", { stat = "Sum" }],
            [".", "RateLimitHits", { stat = "Sum" }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = local.region
          title   = "LogGuardian Compliance Metrics"
          period  = 300
        }
      },
      {
        type = "log"
        properties = {
          query   = "SOURCE '/aws/lambda/${aws_lambda_function.compliance.function_name}' | fields @timestamp, @message | filter @message like /ERROR/ | sort @timestamp desc | limit 20"
          region  = local.region
          stacked = false
          title   = "Recent Lambda Errors"
          view    = "table"
        }
      },
      {
        type = "metric"
        properties = {
          metrics = [
            ["AWS/Lambda", "ConcurrentExecutions", { stat = "Maximum" }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = local.region
          title   = "Lambda Concurrency"
          period  = 60
          dimensions = {
            FunctionName = [aws_lambda_function.compliance.function_name]
          }
        }
      }
    ]
  })
}

# ============================================
# CloudWatch Alarms
# ============================================

resource "aws_cloudwatch_metric_alarm" "lambda_errors" {
  count = var.enable_cloudwatch_alarms ? 1 : 0

  alarm_name          = "${local.name_prefix}-lambda-errors"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "Errors"
  namespace           = "AWS/Lambda"
  period              = 300
  statistic           = "Sum"
  threshold           = 5
  alarm_description   = "Alert when Lambda function has more than 5 errors in 5 minutes"
  treat_missing_data  = "notBreaching"

  dimensions = {
    FunctionName = aws_lambda_function.compliance.function_name
  }

  alarm_actions = var.alarm_sns_topic_arn != null ? [var.alarm_sns_topic_arn] : []

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-lambda-errors"
    }
  )
}

resource "aws_cloudwatch_metric_alarm" "lambda_throttles" {
  count = var.enable_cloudwatch_alarms ? 1 : 0

  alarm_name          = "${local.name_prefix}-lambda-throttles"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "Throttles"
  namespace           = "AWS/Lambda"
  period              = 300
  statistic           = "Sum"
  threshold           = 3
  alarm_description   = "Alert when Lambda function is throttled"
  treat_missing_data  = "notBreaching"

  dimensions = {
    FunctionName = aws_lambda_function.compliance.function_name
  }

  alarm_actions = var.alarm_sns_topic_arn != null ? [var.alarm_sns_topic_arn] : []

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-lambda-throttles"
    }
  )
}

resource "aws_cloudwatch_metric_alarm" "lambda_duration" {
  count = var.enable_cloudwatch_alarms ? 1 : 0

  alarm_name          = "${local.name_prefix}-lambda-duration"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "Duration"
  namespace           = "AWS/Lambda"
  period              = 300
  statistic           = "Average"
  threshold           = var.lambda_timeout * 1000 * 0.8 # 80% of timeout in milliseconds
  alarm_description   = "Alert when Lambda function duration approaches timeout"
  treat_missing_data  = "notBreaching"

  dimensions = {
    FunctionName = aws_lambda_function.compliance.function_name
  }

  alarm_actions = var.alarm_sns_topic_arn != null ? [var.alarm_sns_topic_arn] : []

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-lambda-duration"
    }
  )
}

resource "aws_cloudwatch_metric_alarm" "remediation_failures" {
  count = var.enable_cloudwatch_alarms ? 1 : 0

  alarm_name          = "${local.name_prefix}-remediation-failures"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "RemediationErrors"
  namespace           = "LogGuardian"
  period              = 300
  statistic           = "Sum"
  threshold           = 10
  alarm_description   = "Alert when remediation failures exceed threshold"
  treat_missing_data  = "notBreaching"

  alarm_actions = var.alarm_sns_topic_arn != null ? [var.alarm_sns_topic_arn] : []

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-remediation-failures"
    }
  )
}