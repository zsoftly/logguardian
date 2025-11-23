# ============================================
# Lambda Function
# ============================================

resource "aws_lambda_function" "compliance" {
  function_name = local.lambda_function_name
  description   = "LogGuardian compliance automation for CloudWatch Logs (${var.environment})"

  # Deployment package from S3
  s3_bucket = var.lambda_s3_bucket
  s3_key    = var.lambda_s3_key

  # Runtime configuration
  runtime       = "provided.al2023"
  handler       = "bootstrap"
  architectures = ["x86_64"]

  # Resource allocation
  memory_size = var.lambda_memory_size
  timeout     = var.lambda_timeout

  # IAM role - referenced from iam.tf via locals
  role = local.lambda_role_arn

  # Environment variables
  environment {
    variables = local.lambda_env_vars
  }

  # Logging configuration
  logging_config {
    log_format = "JSON"
    log_group  = aws_cloudwatch_log_group.lambda.name
  }

  tags = merge(
    local.common_tags,
    {
      Name = local.lambda_function_name
    }
  )

  depends_on = [
    aws_cloudwatch_log_group.lambda,
    aws_iam_role_policy_attachment.lambda_basic
  ]
}

# ============================================
# Lambda CloudWatch Log Group
# ============================================

resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${local.lambda_function_name}"
  retention_in_days = var.lambda_log_retention_days

  tags = merge(
    local.common_tags,
    {
      Name = "/aws/lambda/${local.lambda_function_name}"
    }
  )
}

# ============================================
# Lambda Permissions for EventBridge
# ============================================

resource "aws_lambda_permission" "eventbridge_encryption" {
  count = var.create_eventbridge_rules ? 1 : 0

  statement_id  = "AllowExecutionFromEventBridgeEncryption"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.compliance.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.encryption_schedule[0].arn
}

resource "aws_lambda_permission" "eventbridge_retention" {
  count = var.create_eventbridge_rules ? 1 : 0

  statement_id  = "AllowExecutionFromEventBridgeRetention"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.compliance.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.retention_schedule[0].arn
}

# ============================================
# Lambda Permissions for Config
# ============================================

resource "aws_lambda_permission" "config" {
  count = var.create_config_rules ? 1 : 0

  statement_id   = "AllowExecutionFromConfig"
  action         = "lambda:InvokeFunction"
  function_name  = aws_lambda_function.compliance.function_name
  principal      = "config.amazonaws.com"
  source_account = local.account_id
}
