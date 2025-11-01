resource "aws_cloudwatch_event_rule" "encryption_schedule" {
  count = var.enable_scheduling ? 1 : 0

  name                = "${local.name_prefix}-encryption-schedule"
  description         = "Scheduled trigger for LogGuardian encryption compliance checks"
  schedule_expression = var.encryption_schedule_expression
  state               = "ENABLED"
}

resource "aws_cloudwatch_event_target" "encryption_ecs_target" {
  count = var.enable_scheduling ? 1 : 0

  rule     = aws_cloudwatch_event_rule.encryption_schedule[0].name
  arn      = aws_ecs_cluster.logguardian.arn
  role_arn = aws_iam_role.eventbridge_ecs_role.arn

  ecs_target {
    task_count          = 1
    task_definition_arn = aws_ecs_task_definition.logguardian.arn
    launch_type         = "FARGATE"
    platform_version    = "LATEST"

    network_configuration {
      subnets          = local.subnet_ids
      security_groups  = [aws_security_group.ecs_tasks.id]
      assign_public_ip = var.assign_public_ip
    }
  }

  input = jsonencode({
    containerOverrides = [{
      name = "logguardian"
      command = concat(
        ["--config-rule", local.encryption_config_rule, "--region", var.region],
        var.dry_run ? ["--dry-run"] : []
      )
    }]
  })
}

resource "aws_cloudwatch_event_rule" "retention_schedule" {
  count = var.enable_scheduling ? 1 : 0

  name                = "${local.name_prefix}-retention-schedule"
  description         = "Scheduled trigger for LogGuardian retention compliance checks"
  schedule_expression = var.retention_schedule_expression
  state               = "ENABLED"
}

resource "aws_cloudwatch_event_target" "retention_ecs_target" {
  count = var.enable_scheduling ? 1 : 0

  rule     = aws_cloudwatch_event_rule.retention_schedule[0].name
  arn      = aws_ecs_cluster.logguardian.arn
  role_arn = aws_iam_role.eventbridge_ecs_role.arn

  ecs_target {
    task_count          = 1
    task_definition_arn = aws_ecs_task_definition.logguardian.arn
    launch_type         = "FARGATE"
    platform_version    = "LATEST"

    network_configuration {
      subnets          = local.subnet_ids
      security_groups  = [aws_security_group.ecs_tasks.id]
      assign_public_ip = var.assign_public_ip
    }
  }

  input = jsonencode({
    containerOverrides = [{
      name = "logguardian"
      command = concat(
        ["--config-rule", local.retention_config_rule, "--region", var.region],
        var.dry_run ? ["--dry-run"] : []
      )
    }]
  })
}
