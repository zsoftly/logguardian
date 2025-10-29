# Lambda Outputs
output "lambda_function_name" {
  description = "Name of the LogGuardian Lambda function"
  value       = module.lambda.function_name
}

output "lambda_function_arn" {
  description = "ARN of the LogGuardian Lambda function"
  value       = module.lambda.function_arn
}

output "lambda_log_group" {
  description = "CloudWatch Log Group for Lambda logs"
  value       = module.lambda.log_group_name
}

# KMS Outputs
output "kms_key_arn" {
  description = "ARN of the KMS key for log encryption"
  value       = module.kms.kms_key_arn
}

output "kms_key_alias" {
  description = "Alias of the KMS key"
  value       = module.kms.kms_key_alias
}

# Config Outputs
output "config_bucket_name" {
  description = "Name of the S3 bucket for AWS Config data"
  value       = module.storage.config_bucket_name
}

output "encryption_config_rule_name" {
  description = "Name of the encryption Config rule"
  value       = local.encryption_config_rule
}

output "retention_config_rule_name" {
  description = "Name of the retention Config rule"
  value       = local.retention_config_rule
}

# EventBridge Outputs
output "encryption_schedule_rule_name" {
  description = "Name of the encryption schedule EventBridge rule"
  value       = module.eventbridge.encryption_rule_name
}

output "retention_schedule_rule_name" {
  description = "Name of the retention schedule EventBridge rule"
  value       = module.eventbridge.retention_rule_name
}

# Monitoring Outputs
output "dashboard_url" {
  description = "URL to the CloudWatch dashboard"
  value       = module.monitoring.dashboard_url
}

output "dashboard_name" {
  description = "Name of the CloudWatch dashboard"
  value       = module.monitoring.dashboard_name
}

# Deployment Summary
output "deployment_summary" {
  description = "Summary of what was created vs what was reused"
  value = {
    kms_key         = var.create_kms_key ? "Created" : "Using Existing"
    config_service  = var.create_config_service ? "Created" : "Using Existing"
    encryption_rule = var.create_encryption_config_rule ? "Created" : "Using Existing"
    retention_rule  = var.create_retention_config_rule ? "Created" : "Using Existing"
    eventbridge     = var.create_eventbridge_rules ? "Created" : "Disabled"
    dashboard       = var.create_monitoring_dashboard ? "Created" : "Disabled"
  }
}

# Manual Invocation Command
output "manual_invocation_command" {
  description = "AWS CLI command to manually invoke the Lambda function"
  value       = <<-EOT
    aws lambda invoke \
      --function-name ${module.lambda.function_name} \
      --payload '{"type":"config-rule-evaluation","configRuleName":"${local.encryption_config_rule}","region":"${var.aws_region}","batchSize":10}' \
      response.json \
      --region ${var.aws_region}
  EOT
}
