# ============================================
# Lambda Outputs
# ============================================

output "lambda_function_name" {
  description = "Name of the LogGuardian Lambda function"
  value       = aws_lambda_function.compliance.function_name
}

output "lambda_function_arn" {
  description = "ARN of the LogGuardian Lambda function"
  value       = aws_lambda_function.compliance.arn
}

output "lambda_role_arn" {
  description = "ARN of the Lambda execution role"
  value       = local.lambda_role_arn
}

output "lambda_log_group_name" {
  description = "CloudWatch Log Group name for Lambda function logs"
  value       = aws_cloudwatch_log_group.lambda.name
}

# ============================================
# KMS Outputs
# ============================================

output "kms_key_id" {
  description = "ID of the KMS key used for CloudWatch Logs encryption"
  value       = local.kms_key_arn
}

output "kms_key_alias" {
  description = "Alias of the KMS key"
  value       = var.create_kms_key ? aws_kms_alias.logs[0].name : null
}

# ============================================
# Config Outputs
# ============================================

output "config_bucket_name" {
  description = "S3 bucket name for AWS Config snapshots"
  value       = local.config_bucket_name_final
}

output "config_role_arn" {
  description = "IAM role ARN for AWS Config service"
  value       = local.config_role_arn_final
}

output "encryption_config_rule_name" {
  description = "Name of the encryption compliance Config rule"
  value       = local.encryption_rule_name
}

output "retention_config_rule_name" {
  description = "Name of the retention compliance Config rule"
  value       = local.retention_rule_name
}

# ============================================
# EventBridge Outputs
# ============================================

output "encryption_schedule_rule_name" {
  description = "Name of the EventBridge rule for encryption checks"
  value       = var.create_eventbridge_rules ? aws_cloudwatch_event_rule.encryption_schedule[0].name : null
}

output "retention_schedule_rule_name" {
  description = "Name of the EventBridge rule for retention checks"
  value       = var.create_eventbridge_rules ? aws_cloudwatch_event_rule.retention_schedule[0].name : null
}

# ============================================
# Monitoring Outputs
# ============================================

output "dashboard_name" {
  description = "Name of the CloudWatch dashboard"
  value       = var.create_monitoring_dashboard ? aws_cloudwatch_dashboard.main[0].dashboard_name : null
}

output "dashboard_url" {
  description = "URL to the CloudWatch dashboard"
  value       = var.create_monitoring_dashboard ? "https://console.aws.amazon.com/cloudwatch/home?region=${local.region}#dashboards:name=${local.dashboard_name}" : null
}

# ============================================
# Invocation Outputs
# ============================================

output "manual_invocation_command" {
  description = "AWS CLI command to manually invoke LogGuardian for encryption compliance"
  value       = <<-EOT
    aws lambda invoke \
      --function-name ${aws_lambda_function.compliance.function_name} \
      --payload '{"type":"config-rule-evaluation","configRuleName":"${local.encryption_rule_name}","region":"${local.region}","batchSize":${var.batch_size}}' \
      --cli-binary-format raw-in-base64-out \
      response.json
  EOT
}

output "test_invocation_payload" {
  description = "Sample payload for testing LogGuardian Lambda function"
  value = {
    encryption_check = {
      type           = "config-rule-evaluation"
      configRuleName = local.encryption_rule_name
      region         = local.region
      batchSize      = var.batch_size
    }
    retention_check = {
      type           = "config-rule-evaluation"
      configRuleName = local.retention_rule_name
      region         = local.region
      batchSize      = var.batch_size
    }
  }
}

# ============================================
# Summary Outputs
# ============================================

output "deployment_summary" {
  description = "Summary of the LogGuardian deployment"
  value = {
    environment            = var.environment
    region                 = local.region
    kms_key_created        = var.create_kms_key
    config_service_created = var.create_config_service
    config_rules_created   = var.create_config_rules
    eventbridge_enabled    = var.create_eventbridge_rules
    monitoring_enabled     = var.create_monitoring_dashboard
    lambda_memory_mb       = var.lambda_memory_size
    default_retention_days = var.default_retention_days
  }
}
