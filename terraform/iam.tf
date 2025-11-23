# ============================================
# IAM Module - Single Source of Truth (DRY Principle)
# ============================================
#
# This file is the ONLY place where IAM resources are defined.
# All other modules reference IAM resources through locals.tf
#
# Architecture:
#   iam.tf (this file)  →  defines IAM roles & policies
#   locals.tf           →  exposes IAM role references
#   other modules       →  consume via locals.lambda_role_arn, etc.
#
# IAM Resources Defined Here:
#   - Lambda execution role (aws_iam_role.lambda)
#   - Lambda policies (CloudWatch, KMS, Config, S3)
#   - Config service role (aws_iam_role.config)
#   - Config S3 policy
#
# Referenced By:
#   - lambda.tf    (via local.lambda_role_arn)
#   - config.tf    (via local.config_role_arn)
#   - kms.tf       (via local.lambda_role_arn for key policy)
#   - monitoring.tf (implicitly via Lambda)
#   - eventbridge.tf (implicitly via Lambda)
#
# ============================================

# ============================================
# Lambda IAM Role
# ============================================

resource "aws_iam_role" "lambda" {
  name               = local.iam_lambda_role_name
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json

  tags = merge(
    local.common_tags,
    {
      Name        = local.iam_lambda_role_name
      Component   = "Lambda"
      Description = "Execution role for LogGuardian Lambda function"
    }
  )
}

data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    sid    = "LambdaAssumeRole"
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
    actions = ["sts:AssumeRole"]
  }
}

# ============================================
# Lambda Managed Policy Attachments
# ============================================

resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = local.lambda_role_name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# ============================================
# Lambda Custom Policy
# ============================================

resource "aws_iam_role_policy" "lambda_logguardian" {
  name   = "${local.iam_lambda_role_name}-policy"
  role   = local.lambda_role_name
  policy = data.aws_iam_policy_document.lambda_logguardian.json
}

data "aws_iam_policy_document" "lambda_logguardian" {
  # CloudWatch Logs permissions
  statement {
    sid    = "CloudWatchLogsRead"
    effect = "Allow"
    actions = [
      "logs:DescribeLogGroups",
      "logs:ListTagsLogGroup",
      "logs:DescribeSubscriptionFilters"
    ]
    resources = ["*"]
  }

  statement {
    sid    = "CloudWatchLogsWrite"
    effect = "Allow"
    actions = [
      "logs:PutRetentionPolicy",
      "logs:AssociateKmsKey",
      "logs:DisassociateKmsKey"
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:logs:${local.region}:${local.account_id}:log-group:*"
    ]
  }

  # KMS permissions
  statement {
    sid    = "KMSAccess"
    effect = "Allow"
    actions = [
      "kms:DescribeKey",
      "kms:GetKeyPolicy",
      "kms:Decrypt",
      "kms:CreateGrant"
    ]
    resources = [local.kms_key_arn]
  }

  # AWS Config permissions
  statement {
    sid    = "ConfigRead"
    effect = "Allow"
    actions = [
      "config:GetComplianceDetailsByConfigRule",
      "config:DescribeConfigRules",
      "config:PutEvaluations"
    ]
    resources = ["*"]
  }

  # S3 permissions for Config bucket
  statement {
    sid    = "S3ConfigAccess"
    effect = "Allow"
    actions = [
      "s3:GetObject",
      "s3:ListBucket"
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::${local.config_bucket_name_final}",
      "arn:${data.aws_partition.current.partition}:s3:::${local.config_bucket_name_final}/*"
    ]
  }
}

# ============================================
# AWS Config IAM Role
# ============================================

resource "aws_iam_role" "config" {
  count = var.create_config_service ? 1 : 0

  name               = local.iam_config_role_name
  assume_role_policy = data.aws_iam_policy_document.config_assume_role.json

  tags = merge(
    local.common_tags,
    {
      Name        = local.iam_config_role_name
      Component   = "Config"
      Description = "Service role for AWS Config"
    }
  )
}

data "aws_iam_policy_document" "config_assume_role" {
  statement {
    sid    = "ConfigAssumeRole"
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["config.amazonaws.com"]
    }
    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role_policy_attachment" "config" {
  count = var.create_config_service ? 1 : 0

  role       = aws_iam_role.config[0].name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/ConfigRole"
}

resource "aws_iam_role_policy" "config_s3" {
  count = var.create_config_service ? 1 : 0

  name   = "${local.iam_config_role_name}-s3-policy"
  role   = aws_iam_role.config[0].id
  policy = data.aws_iam_policy_document.config_s3.json
}

data "aws_iam_policy_document" "config_s3" {
  statement {
    effect = "Allow"
    actions = [
      "s3:GetBucketVersioning",
      "s3:PutObject",
      "s3:GetObject"
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::${local.config_bucket_name_final}",
      "arn:${data.aws_partition.current.partition}:s3:::${local.config_bucket_name_final}/*"
    ]
  }
}
