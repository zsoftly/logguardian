# ============================================
# Local Variables for DRY Principle
# ============================================

locals {
  # Naming convention: <product>-<component>-<environment>
  name_prefix = lower("${var.product_name}-${var.environment}")

  # Resource names
  lambda_function_name = "${local.name_prefix}-compliance"
  config_bucket_name   = "${local.name_prefix}-config-${data.aws_caller_identity.current.account_id}"

  # IAM role names - centralized naming for IAM resources
  iam_lambda_role_name = "${local.name_prefix}-lambda-role"
  iam_config_role_name = "${local.name_prefix}-config-role"

  # Config rule names
  encryption_rule_name = var.create_config_rules ? "${local.name_prefix}-encryption-rule" : var.existing_encryption_config_rule
  retention_rule_name  = var.create_config_rules ? "${local.name_prefix}-retention-rule" : var.existing_retention_config_rule

  # KMS configuration
  kms_key_arn   = var.create_kms_key ? aws_kms_key.logs[0].arn : var.existing_kms_key_arn
  kms_key_alias = var.kms_key_alias != null ? var.kms_key_alias : "${local.name_prefix}-logs-key"

  # Config service configuration
  config_bucket_name_final = var.create_config_service ? local.config_bucket_name : var.existing_config_bucket
  config_role_arn_final    = var.create_config_service ? aws_iam_role.config[0].arn : var.existing_config_service_role_arn

  # IAM role references - single source of truth from iam.tf
  lambda_role_arn  = aws_iam_role.lambda.arn
  lambda_role_name = aws_iam_role.lambda.name
  config_role_arn  = var.create_config_service ? aws_iam_role.config[0].arn : var.existing_config_service_role_arn
  config_role_name = var.create_config_service ? aws_iam_role.config[0].name : null

  # Lambda environment variables
  lambda_env_vars = merge(
    {
      LOG_LEVEL              = var.lambda_log_level
      DEFAULT_RETENTION_DAYS = tostring(var.default_retention_days)
      BATCH_SIZE             = tostring(var.batch_size)
      KMS_KEY_ARN            = local.kms_key_arn
      ENVIRONMENT            = var.environment
    },
    length(var.supported_regions) > 0 ? {
      SUPPORTED_REGIONS = join(",", var.supported_regions)
    } : {},
    var.additional_lambda_env_vars
  )

  # Consolidated tags
  common_tags = merge(
    {
      Product     = var.product_name
      Environment = var.environment
      Owner       = var.owner
      ManagedBy   = "Terraform"
    },
    var.additional_tags
  )

  # CloudWatch dashboard name
  dashboard_name = "${local.name_prefix}-dashboard"

  # EventBridge rule names
  encryption_schedule_rule_name = "${local.name_prefix}-encryption-schedule"
  retention_schedule_rule_name  = "${local.name_prefix}-retention-schedule"

  # Current AWS account and region
  account_id = data.aws_caller_identity.current.account_id
  region     = data.aws_region.current.name
}

# ============================================
# Data Sources
# ============================================

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}
