# Migrate existing SAM deployment to Terraform

terraform {
  required_version = ">= 1.5.0"

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

# Import existing SAM resources into Terraform management
module "logguardian" {
  source = "../../"

  environment = var.environment
  aws_region  = var.aws_region

  # Match SAM deployment settings
  create_kms_key                 = var.sam_created_kms
  existing_kms_key_arn           = var.sam_created_kms ? null : var.existing_kms_key_arn
  create_config_service          = var.sam_created_config
  create_encryption_config_rule  = var.sam_created_encryption_rule
  create_retention_config_rule   = var.sam_created_retention_rule
  create_eventbridge_rules       = var.sam_created_eventbridge
  create_monitoring_dashboard    = var.sam_created_dashboard

  # Lambda settings
  default_retention_days    = var.default_retention_days
  lambda_log_retention_days = 30
  lambda_memory_size        = 128
  lambda_timeout            = 60
  log_level                 = "INFO"

  product_name = "LogGuardian"
  owner        = var.owner
}
