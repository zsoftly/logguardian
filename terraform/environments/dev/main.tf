terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "account_id" {
  description = "AWS Account ID"
  type        = string
}

variable "environment" {
  description = "Environment (dev/prod)"
  type        = string
}

variable "container_image" {
  description = "Container image URL"
  type        = string
}

variable "cpu" {
  description = "Fargate CPU units"
  type        = string
  default     = "256"
}

variable "memory" {
  description = "Fargate memory MB"
  type        = string
  default     = "512"
}

variable "enable_spot" {
  description = "Use Fargate Spot"
  type        = bool
  default     = true
}

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "subnet_ids" {
  description = "Subnet IDs"
  type        = list(string)
}

variable "assign_public_ip" {
  description = "Assign public IP"
  type        = bool
  default     = false
}

variable "tags" {
  description = "Resource tags"
  type        = map(string)
  default     = {}
}

# Security Group
resource "aws_security_group" "ecs_tasks" {
  name        = "logguardian-${var.environment}-ecs-tasks"
  description = "Security group for LogGuardian ECS tasks"
  vpc_id      = var.vpc_id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound traffic"
  }

  tags = merge(var.tags, {
    Name = "logguardian-${var.environment}-ecs-tasks"
  })
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "logguardian" {
  name              = "/ecs/logguardian"
  retention_in_days = 30
  tags              = var.tags
}

# ECS Cluster
resource "aws_ecs_cluster" "logguardian" {
  name = "logguardian-${var.environment}"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = var.tags
}

# ECS Cluster Capacity Providers
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

# Task Definition
resource "aws_ecs_task_definition" "logguardian" {
  family                   = "logguardian-${var.environment}"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.cpu
  memory                   = var.memory
  execution_role_arn       = aws_iam_role.execution_role.arn
  task_role_arn            = aws_iam_role.task_role.arn

  container_definitions = jsonencode([{
    name      = "logguardian"
    image     = var.container_image
    essential = true
    
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.logguardian.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "ecs"
      }
    }

    healthCheck = {
      command     = ["CMD-SHELL", "exit 0"]
      interval    = 30
      timeout     = 5
      retries     = 3
      startPeriod = 10
    }
  }])

  tags = var.tags
}

# Outputs
output "cluster_name" {
  value = aws_ecs_cluster.logguardian.name
}

output "task_definition_arn" {
  value = aws_ecs_task_definition.logguardian.arn
}

output "task_definition_family" {
  value = aws_ecs_task_definition.logguardian.family
}

output "security_group_id" {
  value = aws_security_group.ecs_tasks.id
}

output "log_group_name" {
  value = aws_cloudwatch_log_group.logguardian.name
}

output "task_role_arn" {
  value = aws_iam_role.task_role.arn
}

output "execution_role_arn" {
  value = aws_iam_role.execution_role.arn
}
