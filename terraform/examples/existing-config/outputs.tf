output "lambda_function_name" {
  description = "Lambda function name"
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
