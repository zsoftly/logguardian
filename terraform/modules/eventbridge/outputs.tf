output "encryption_rule_arn" {
  description = "ARN of encryption EventBridge rule"
  value       = var.create_eventbridge_rules ? aws_cloudwatch_event_rule.encryption[0].arn : null
}

output "encryption_rule_name" {
  description = "Name of encryption EventBridge rule"
  value       = var.create_eventbridge_rules ? aws_cloudwatch_event_rule.encryption[0].name : null
}

output "retention_rule_arn" {
  description = "ARN of retention EventBridge rule"
  value       = var.create_eventbridge_rules ? aws_cloudwatch_event_rule.retention[0].arn : null
}

output "retention_rule_name" {
  description = "Name of retention EventBridge rule"
  value       = var.create_eventbridge_rules ? aws_cloudwatch_event_rule.retention[0].name : null
}

output "eventbridge_rules_created" {
  description = "Whether EventBridge rules were created"
  value       = var.create_eventbridge_rules
}
