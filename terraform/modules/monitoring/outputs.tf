output "dashboard_name" {
  description = "Name of the CloudWatch dashboard"
  value       = var.create_dashboard ? aws_cloudwatch_dashboard.logguardian[0].dashboard_name : null
}

output "dashboard_arn" {
  description = "ARN of the CloudWatch dashboard"
  value       = var.create_dashboard ? aws_cloudwatch_dashboard.logguardian[0].dashboard_arn : null
}

output "dashboard_url" {
  description = "URL to view the CloudWatch dashboard"
  value       = var.create_dashboard ? "https://${data.aws_region.current.id}.console.aws.amazon.com/cloudwatch/home?region=${data.aws_region.current.id}#dashboards:name=${aws_cloudwatch_dashboard.logguardian[0].dashboard_name}" : null
}

output "dashboard_created" {
  description = "Whether dashboard was created"
  value       = var.create_dashboard
}
