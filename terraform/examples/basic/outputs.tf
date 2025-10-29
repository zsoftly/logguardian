output "lambda_function_name" {
  description = "Name of the Lambda function"
  value       = module.logguardian.lambda_function_name
}

output "dashboard_url" {
  description = "CloudWatch dashboard URL"
  value       = module.logguardian.dashboard_url
}

output "deployment_summary" {
  description = "Deployment summary"
  value       = module.logguardian.deployment_summary
}

output "manual_invocation_command" {
  description = "Command to manually invoke Lambda"
  value       = module.logguardian.manual_invocation_command
}
