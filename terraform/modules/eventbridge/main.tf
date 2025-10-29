# EventBridge rules for scheduled Lambda invocation

data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

# EventBridge rule for encryption compliance checks
resource "aws_cloudwatch_event_rule" "encryption" {
  count = var.create_eventbridge_rules ? 1 : 0

  name                = "logguardian-encryption-schedule-${var.environment}"
  description         = "Scheduled trigger for LogGuardian encryption compliance checks"
  schedule_expression = var.encryption_schedule
  state               = "ENABLED"

  tags = merge(
    var.tags,
    {
      Name        = "logguardian-encryption-schedule-${var.environment}"
      Environment = var.environment
      Purpose     = "Scheduled encryption compliance checks"
      ManagedBy   = "Terraform"
      Module      = "eventbridge"
    }
  )
}

# EventBridge target for encryption rule
resource "aws_cloudwatch_event_target" "encryption" {
  count = var.create_eventbridge_rules ? 1 : 0

  rule      = aws_cloudwatch_event_rule.encryption[0].name
  target_id = "LogGuardianEncryptionTarget"
  arn       = var.lambda_function_arn

  input = jsonencode({
    type           = "config-rule-evaluation"
    configRuleName = var.encryption_config_rule
    region         = data.aws_region.current.id
    account        = data.aws_caller_identity.current.account_id
    environment    = var.environment
  })
}

# Lambda permission for encryption EventBridge rule
resource "aws_lambda_permission" "encryption" {
  count = var.create_eventbridge_rules ? 1 : 0

  statement_id  = "AllowExecutionFromEventBridgeEncryption"
  action        = "lambda:InvokeFunction"
  function_name = var.lambda_function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.encryption[0].arn
}

# EventBridge rule for retention compliance checks
resource "aws_cloudwatch_event_rule" "retention" {
  count = var.create_eventbridge_rules ? 1 : 0

  name                = "logguardian-retention-schedule-${var.environment}"
  description         = "Scheduled trigger for LogGuardian retention compliance checks"
  schedule_expression = var.retention_schedule
  state               = "ENABLED"

  tags = merge(
    var.tags,
    {
      Name        = "logguardian-retention-schedule-${var.environment}"
      Environment = var.environment
      Purpose     = "Scheduled retention compliance checks"
      ManagedBy   = "Terraform"
      Module      = "eventbridge"
    }
  )
}

# EventBridge target for retention rule
resource "aws_cloudwatch_event_target" "retention" {
  count = var.create_eventbridge_rules ? 1 : 0

  rule      = aws_cloudwatch_event_rule.retention[0].name
  target_id = "LogGuardianRetentionTarget"
  arn       = var.lambda_function_arn

  input = jsonencode({
    type           = "config-rule-evaluation"
    configRuleName = var.retention_config_rule
    region         = data.aws_region.current.id
    account        = data.aws_caller_identity.current.account_id
    environment    = var.environment
  })
}

# Lambda permission for retention EventBridge rule
resource "aws_lambda_permission" "retention" {
  count = var.create_eventbridge_rules ? 1 : 0

  statement_id  = "AllowExecutionFromEventBridgeRetention"
  action        = "lambda:InvokeFunction"
  function_name = var.lambda_function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.retention[0].arn
}
