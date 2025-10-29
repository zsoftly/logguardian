# IAM roles and policies for Lambda execution

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# Lambda execution role
resource "aws_iam_role" "lambda_execution" {
  name = "LogGuardian-LambdaRole-${var.environment}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }]
  })

  tags = merge(
    var.tags,
    {
      Name        = "LogGuardian-LambdaRole-${var.environment}"
      Environment = var.environment
      ManagedBy   = "Terraform"
      Module      = "iam"
    }
  )
}

# Lambda execution policy
resource "aws_iam_role_policy" "lambda_execution" {
  name = "LogGuardianExecutionPolicy"
  role = aws_iam_role.lambda_execution.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      # AWS Config permissions
      {
        Sid    = "ConfigReadAccess"
        Effect = "Allow"
        Action = [
          "config:GetComplianceDetailsByConfigRule",
          "config:GetComplianceDetailsByResource",
          "config:DescribeConfigRules",
          "config:DescribeComplianceByConfigRule"
        ]
        Resource = "*"
      },
      # CloudWatch Logs permissions
      {
        Sid    = "CloudWatchLogsAccess"
        Effect = "Allow"
        Action = [
          "logs:AssociateKmsKey",
          "logs:PutRetentionPolicy",
          "logs:DescribeLogGroups",
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "*"
      },
      # KMS permissions
      {
        Sid    = "KMSAccess"
        Effect = "Allow"
        Action = [
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:ReEncrypt*",
          "kms:GenerateDataKey*",
          "kms:DescribeKey",
          "kms:GetKeyPolicy",
          "kms:ListAliases"
        ]
        Resource = var.kms_key_arn
      },
      # CloudWatch Metrics permissions
      {
        Sid    = "CloudWatchMetricsAccess"
        Effect = "Allow"
        Action = [
          "cloudwatch:PutMetricData"
        ]
        Resource = "*"
        Condition = {
          StringEquals = {
            "cloudwatch:namespace" = "LogGuardian"
          }
        }
      }
    ]
  })
}
