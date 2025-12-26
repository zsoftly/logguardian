output "cluster_name" {
  description = "ECS cluster name"
  value       = aws_ecs_cluster.logguardian.name
}

output "task_definition_family" {
  description = "Task definition family"
  value       = aws_ecs_task_definition.logguardian.family
}

output "log_group_name" {
  description = "CloudWatch log group"
  value       = aws_cloudwatch_log_group.logguardian.name
}

output "log_tail_command" {
  description = "Command to tail logs in real-time"
  value       = "aws logs tail ${aws_cloudwatch_log_group.logguardian.name} --since 5m --follow"
}
