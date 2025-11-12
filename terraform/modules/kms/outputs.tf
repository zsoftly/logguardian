output "kms_key_arn" {
  description = "ARN of the KMS key (created or existing)"
  value       = var.create_kms_key ? aws_kms_key.logguardian[0].arn : var.existing_kms_key_arn
}

output "kms_key_id" {
  description = "ID of the KMS key (null if using existing)"
  value       = var.create_kms_key ? aws_kms_key.logguardian[0].key_id : null
}

output "kms_key_alias" {
  description = "Alias of the KMS key"
  value       = var.kms_key_alias
}

output "kms_key_rotation_enabled" {
  description = "Whether automatic key rotation is enabled"
  value       = var.create_kms_key ? aws_kms_key.logguardian[0].enable_key_rotation : null
}

output "kms_key_created" {
  description = "Whether a new KMS key was created"
  value       = var.create_kms_key
}
