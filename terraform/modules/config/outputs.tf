output "encryption_config_rule_name" {
  description = "Name of the encryption Config rule (created or existing)"
  value       = var.create_encryption_rule ? aws_config_config_rule.encryption[0].name : var.existing_encryption_rule
}

output "retention_config_rule_name" {
  description = "Name of the retention Config rule (created or existing)"
  value       = var.create_retention_rule ? aws_config_config_rule.retention[0].name : var.existing_retention_rule
}

output "encryption_config_rule_arn" {
  description = "ARN of the encryption Config rule"
  value       = var.create_encryption_rule ? aws_config_config_rule.encryption[0].arn : null
}

output "retention_config_rule_arn" {
  description = "ARN of the retention Config rule"
  value       = var.create_retention_rule ? aws_config_config_rule.retention[0].arn : null
}

output "config_recorder_name" {
  description = "Name of the Config recorder"
  value       = var.create_config_service ? aws_config_configuration_recorder.main[0].name : null
}

output "config_recorder_id" {
  description = "ID of the Config recorder"
  value       = var.create_config_service ? aws_config_configuration_recorder.main[0].id : null
}

output "config_service_created" {
  description = "Whether Config service was created by this module"
  value       = var.create_config_service
}

output "encryption_rule_created" {
  description = "Whether encryption rule was created by this module"
  value       = var.create_encryption_rule
}

output "retention_rule_created" {
  description = "Whether retention rule was created by this module"
  value       = var.create_retention_rule
}
