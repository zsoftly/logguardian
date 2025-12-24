# ============================================
# EventBridge Rule - Encryption Compliance Check
# ============================================

resource "aws_cloudwatch_event_rule" "encryption_schedule" {
  count = var.create_eventbridge_rules ? 1 : 0

  name                = local.encryption_schedule_rule_name
  description         = "Scheduled trigger for CloudWatch Logs encryption compliance check"
  schedule_expression = var.encryption_schedule_expression

  tags = merge(
    local.common_tags,
    {
      Name = local.encryption_schedule_rule_name
      Type = "EncryptionCompliance"
    }
  )
}

resource "aws_cloudwatch_event_target" "encryption_schedule" {
  count = var.create_eventbridge_rules ? 1 : 0

  rule      = aws_cloudwatch_event_rule.encryption_schedule[0].name
  target_id = "LogGuardianEncryptionCheck"
  arn       = aws_lambda_function.compliance.arn

  input = jsonencode({
    type           = "config-rule-evaluation"
    configRuleName = local.encryption_rule_name
    region         = local.region
    batchSize      = var.batch_size
  })
}

# ============================================
# EventBridge Rule - Retention Compliance Check
# ============================================

resource "aws_cloudwatch_event_rule" "retention_schedule" {
  count = var.create_eventbridge_rules ? 1 : 0

  name                = local.retention_schedule_rule_name
  description         = "Scheduled trigger for CloudWatch Logs retention compliance check"
  schedule_expression = var.retention_schedule_expression

  tags = merge(
    local.common_tags,
    {
      Name = local.retention_schedule_rule_name
      Type = "RetentionCompliance"
    }
  )
}

resource "aws_cloudwatch_event_target" "retention_schedule" {
  count = var.create_eventbridge_rules ? 1 : 0

  rule      = aws_cloudwatch_event_rule.retention_schedule[0].name
  target_id = "LogGuardianRetentionCheck"
  arn       = aws_lambda_function.compliance.arn

  input = jsonencode({
    type           = "config-rule-evaluation"
    configRuleName = local.retention_rule_name
    region         = local.region
    batchSize      = var.batch_size
  })
}
