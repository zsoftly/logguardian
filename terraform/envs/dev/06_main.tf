resource "aws_security_group" "ecs_tasks" {
  name        = "${local.name_prefix}-ecs-tasks"
  description = "Security group for LogGuardian ECS tasks"
  vpc_id      = local.vpc_id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound traffic for AWS API access"
  }
}

resource "aws_cloudwatch_log_group" "logguardian" {
  name              = local.log_group_name
  retention_in_days = var.log_retention_days
}

resource "aws_ecs_cluster" "logguardian" {
  name = local.name_prefix

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

resource "aws_ecs_cluster_capacity_providers" "logguardian" {
  cluster_name       = aws_ecs_cluster.logguardian.name
  capacity_providers = var.enable_spot ? ["FARGATE", "FARGATE_SPOT"] : ["FARGATE"]

  default_capacity_provider_strategy {
    capacity_provider = var.enable_spot ? "FARGATE_SPOT" : "FARGATE"
    weight            = var.enable_spot ? 80 : 100
    base              = 0
  }

  dynamic "default_capacity_provider_strategy" {
    for_each = var.enable_spot ? [1] : []
    content {
      capacity_provider = "FARGATE"
      weight            = 20
      base              = 0
    }
  }
}

resource "aws_ecs_task_definition" "logguardian" {
  family                   = local.name_prefix
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.cpu
  memory                   = var.memory
  execution_role_arn       = aws_iam_role.execution_role.arn
  task_role_arn            = aws_iam_role.task_role.arn

  container_definitions = templatefile("${path.module}/container_definition.json.tpl", {
    container_image = local.container_image
    region          = var.region
    log_group_name  = aws_cloudwatch_log_group.logguardian.name
  })
}
