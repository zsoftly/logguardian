# CloudWatch dashboard for LogGuardian monitoring

data "aws_region" "current" {}

resource "aws_cloudwatch_dashboard" "logguardian" {
  count = var.create_dashboard ? 1 : 0

  dashboard_name = "${var.product_name}-${var.environment}"

  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        x      = 0
        y      = 0
        width  = 12
        height = 6
        properties = {
          metrics = [
            ["AWS/Lambda", "Duration", { stat = "Average", label = "Avg Duration" }, { "FunctionName" = var.lambda_function_name }],
            ["...", { stat = "Maximum", label = "Max Duration" }],
            [".", "Errors", { stat = "Sum", label = "Errors" }, { "." = "." }],
            [".", "Invocations", { stat = "Sum", label = "Invocations" }, { "." = "." }],
            [".", "Throttles", { stat = "Sum", label = "Throttles" }, { "." = "." }]
          ]
          period = 300
          stat   = "Average"
          region = data.aws_region.current.id
          title  = "Lambda Function Metrics"
          yAxis = {
            left = {
              label = "Count / Duration (ms)"
            }
          }
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 0
        width  = 12
        height = 6
        properties = {
          metrics = [
            ["LogGuardian", "LogGroupsProcessed", { stat = "Sum", label = "Processed" }, { "Environment" = var.environment }],
            [".", "LogGroupsRemediated", { stat = "Sum", label = "Remediated" }, { "." = "." }],
            [".", "RemediationErrors", { stat = "Sum", label = "Errors" }, { "." = "." }]
          ]
          period = 300
          stat   = "Sum"
          region = data.aws_region.current.id
          title  = "Compliance Metrics"
          yAxis = {
            left = {
              label = "Count"
            }
          }
        }
      },
      {
        type   = "log"
        x      = 0
        y      = 6
        width  = 24
        height = 6
        properties = {
          query   = "SOURCE '${var.log_group_name}' | fields @timestamp, @message | filter @message like /ERROR/ or @message like /WARN/ | sort @timestamp desc | limit 20"
          region  = data.aws_region.current.id
          title   = "Recent Errors and Warnings"
          stacked = false
        }
      }
    ]
  })
}
