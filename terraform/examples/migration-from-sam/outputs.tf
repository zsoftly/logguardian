output "migration_status" {
  description = "Migration status summary"
  value = {
    environment = var.environment
    message     = "Migration complete - Terraform now manages LogGuardian resources"
    next_steps  = "Delete SAM stack: aws cloudformation delete-stack --stack-name logguardian-${var.environment}"
  }
}

output "lambda_function_name" {
  description = "Lambda function name"
  value       = module.logguardian.lambda_function_name
}

output "dashboard_url" {
  description = "Dashboard URL"
  value       = module.logguardian.dashboard_url
}
