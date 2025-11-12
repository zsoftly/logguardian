output "config_bucket_name" {
  description = "Name of the Config S3 bucket"
  value       = var.create_config_bucket ? aws_s3_bucket.config[0].id : null
}

output "config_bucket_arn" {
  description = "ARN of the Config S3 bucket"
  value       = var.create_config_bucket ? aws_s3_bucket.config[0].arn : null
}

output "config_bucket_region" {
  description = "Region of the Config S3 bucket"
  value       = var.create_config_bucket ? aws_s3_bucket.config[0].region : null
}

output "config_bucket_created" {
  description = "Whether a new Config bucket was created"
  value       = var.create_config_bucket
}
