# LogGuardian Terraform Root Module
# Orchestrates all sub-modules for complete deployment

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = local.common_tags
  }
}

# KMS Module - Encryption key management
module "kms" {
  source = "./modules/kms"

  environment          = var.environment
  create_kms_key       = var.create_kms_key
  existing_kms_key_arn = var.existing_kms_key_arn
  kms_key_alias        = var.kms_key_alias
  enable_key_rotation  = true
  product_name         = var.product_name
  tags                 = var.tags
}

# Storage Module - S3 buckets for Config
module "storage" {
  source = "./modules/storage"

  environment            = var.environment
  create_config_bucket   = var.create_config_service
  enable_lifecycle_rules = var.enable_s3_lifecycle_rules
  s3_expiration_days     = var.s3_expiration_days
  product_name           = var.product_name
  tags                   = var.tags
}

# IAM Module - Lambda execution role
module "iam" {
  source = "./modules/iam"

  environment  = var.environment
  kms_key_arn  = module.kms.kms_key_arn
  product_name = var.product_name
  owner        = var.owner
  managed_by   = var.managed_by
  tags         = var.tags

  depends_on = [module.kms]
}

# Config Module - AWS Config service and rules
module "config" {
  source = "./modules/config"

  environment                = var.environment
  create_config_service      = var.create_config_service
  config_bucket_name         = module.storage.config_bucket_name
  existing_config_role_arn   = var.existing_config_role_arn
  create_encryption_rule     = var.create_encryption_config_rule
  existing_encryption_rule   = var.existing_encryption_config_rule
  create_retention_rule      = var.create_retention_config_rule
  existing_retention_rule    = var.existing_retention_config_rule
  default_retention_days     = var.default_retention_days
  product_name               = var.product_name
  tags                       = var.tags

  depends_on = [module.storage]
}

# Lambda Module - LogGuardian function
module "lambda" {
  source = "./modules/lambda"

  environment               = var.environment
  lambda_role_arn           = module.iam.lambda_execution_role_arn
  kms_key_arn               = module.kms.kms_key_arn
  kms_key_alias             = var.kms_key_alias
  encryption_config_rule    = local.encryption_config_rule
  retention_config_rule     = local.retention_config_rule
  default_retention_days    = var.default_retention_days
  lambda_log_retention_days = var.lambda_log_retention_days
  lambda_memory_size        = var.lambda_memory_size
  lambda_timeout            = var.lambda_timeout
  log_level                 = var.log_level
  dry_run                   = var.dry_run
  lambda_code_path          = var.lambda_code_path
  product_name              = var.product_name
  owner                     = var.owner
  managed_by                = var.managed_by
  tags                      = var.tags

  depends_on = [module.iam, module.config]
}

# EventBridge Module - Scheduled triggers
module "eventbridge" {
  source = "./modules/eventbridge"

  environment              = var.environment
  create_eventbridge_rules = var.create_eventbridge_rules
  lambda_function_arn      = module.lambda.function_arn
  lambda_function_name     = module.lambda.function_name
  encryption_config_rule   = local.encryption_config_rule
  retention_config_rule    = local.retention_config_rule
  encryption_schedule      = var.encryption_schedule_expression
  retention_schedule       = var.retention_schedule_expression
  tags                     = var.tags

  depends_on = [module.lambda]
}

# Monitoring Module - CloudWatch dashboard
module "monitoring" {
  source = "./modules/monitoring"

  environment          = var.environment
  create_dashboard     = var.create_monitoring_dashboard
  lambda_function_name = module.lambda.function_name
  log_group_name       = module.lambda.log_group_name
  product_name         = var.product_name
  tags                 = var.tags

  depends_on = [module.lambda]
}
