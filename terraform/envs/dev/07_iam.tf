resource "aws_iam_role" "execution_role" {
  name = "${local.name_prefix}-execution"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ecs-tasks.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "execution_role_policy" {
  role       = aws_iam_role.execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role" "task_role" {
  name = "${local.name_prefix}-task"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ecs-tasks.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "task_policy" {
  name = "${local.name_prefix}-task-policy"
  role = aws_iam_role.task_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "ConfigAccess"
        Effect = "Allow"
        Action = [
          "config:DescribeConfigRules",
          "config:DescribeConfigRuleEvaluationStatus",
          "config:GetComplianceDetailsByConfigRule",
          "config:GetComplianceDetailsByResource",
          "config:ListDiscoveredResources",
          "config:PutEvaluations"
        ]
        Resource = [
          "arn:aws:config:${var.region}:${local.account_id}:config-rule/*"
        ]
      },
      {
        Sid    = "ConfigRuleManagement"
        Effect = "Allow"
        Action = [
          "config:PutConfigRule",
          "config:DescribeConfigRules"
        ]
        Resource = "arn:aws:config:${var.region}:${local.account_id}:config-rule/cw-lg-*"
      },
      {
        Sid    = "LogsManagement"
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:PutRetentionPolicy",
          "logs:AssociateKmsKey",
          "logs:DescribeLogGroups",
          "logs:ListTagsLogGroup",
          "logs:ListTagsForResource"
        ]
        Resource = [
          "arn:aws:logs:${var.region}:${local.account_id}:log-group:*"
        ]
      },
      {
        Sid      = "CloudWatchMetrics"
        Effect   = "Allow"
        Action   = ["cloudwatch:PutMetricData"]
        Resource = "*"
      },
      {
        Sid    = "KMSAccess"
        Effect = "Allow"
        Action = [
          "kms:DescribeKey",
          "kms:ListKeys",
          "kms:ListAliases"
        ]
        Resource = "arn:aws:kms:${var.region}:${local.account_id}:key/*"
      }
    ]
  })
}

resource "aws_iam_role" "eventbridge_ecs_role" {
  name = "${local.name_prefix}-eventbridge-ecs"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "events.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "eventbridge_ecs_policy" {
  name = "${local.name_prefix}-eventbridge-ecs-policy"
  role = aws_iam_role.eventbridge_ecs_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "ECSTaskExecution"
        Effect = "Allow"
        Action = [
          "ecs:RunTask"
        ]
        Resource = [
          aws_ecs_task_definition.logguardian.arn
        ]
      },
      {
        Sid    = "PassRoleToECS"
        Effect = "Allow"
        Action = [
          "iam:PassRole"
        ]
        Resource = [
          aws_iam_role.task_role.arn,
          aws_iam_role.execution_role.arn
        ]
      }
    ]
  })
}
