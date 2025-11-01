output "cluster_name" {
  description = "ECS cluster name for running tasks"
  value       = aws_ecs_cluster.logguardian.name
}

output "cluster_arn" {
  description = "ECS cluster ARN"
  value       = aws_ecs_cluster.logguardian.arn
}

output "task_definition_arn" {
  description = "Full ARN of the task definition"
  value       = aws_ecs_task_definition.logguardian.arn
}

output "task_definition_family" {
  description = "Task definition family name"
  value       = aws_ecs_task_definition.logguardian.family
}

output "task_role_arn" {
  description = "IAM role ARN for task permissions"
  value       = aws_iam_role.task_role.arn
}

output "execution_role_arn" {
  description = "IAM role ARN for ECS agent"
  value       = aws_iam_role.execution_role.arn
}

output "security_group_id" {
  description = "Security group ID for ECS tasks"
  value       = aws_security_group.ecs_tasks.id
}

output "log_group_name" {
  description = "CloudWatch log group name"
  value       = aws_cloudwatch_log_group.logguardian.name
}

output "subnet_ids" {
  description = "Subnet IDs for task execution"
  value       = var.subnet_ids
}

output "vpc_id" {
  description = "VPC ID where tasks run"
  value       = local.vpc_id
}

output "eventbridge_role_arn" {
  description = "IAM role ARN for EventBridge to run ECS tasks"
  value       = aws_iam_role.eventbridge_ecs_role.arn
}

output "encryption_schedule_rule_name" {
  description = "EventBridge rule name for encryption compliance checks"
  value       = var.enable_scheduling ? aws_cloudwatch_event_rule.encryption_schedule[0].name : null
}

output "retention_schedule_rule_name" {
  description = "EventBridge rule name for retention compliance checks"
  value       = var.enable_scheduling ? aws_cloudwatch_event_rule.retention_schedule[0].name : null
}

output "scheduling_enabled" {
  description = "Whether automated scheduling is enabled"
  value       = var.enable_scheduling
}
