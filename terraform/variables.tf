# Environment Configuration
variable "environment" {
  description = "Deployment environment (dev, staging, prod, test)"
  type        = string
  default     = "prod"

  validation {
    condition     = contains(["dev", "staging", "prod", "test"], var.environment)
    error_message = "Environment must be one of: dev, staging, prod, test."
  }
}

variable "aws_region" {
  description = "AWS region for deployment"
  type        = string
  default     = "ca-central-1"
}

# KMS Configuration
variable "create_kms_key" {
  description = "Create new KMS key or use existing"
  type        = bool
  default     = true
}

variable "existing_kms_key_arn" {
  description = "ARN of existing KMS key (required if create_kms_key=false)"
  type        = string
  default     = null
}

variable "kms_key_alias" {
  description = "KMS key alias"
  type        = string
  default     = "alias/logguardian-cloudwatch-logs"
}

# AWS Config Configuration
variable "create_config_service" {
  description = "Create AWS Config service resources (recorder, delivery channel, bucket)"
  type        = bool
  default     = false
}

variable "existing_config_bucket" {
  description = "Name of existing S3 bucket for Config (required if create_config_service=false)"
  type        = string
  default     = null
}

variable "existing_config_role_arn" {
  description = "ARN of existing Config service role (required if create_config_service=false)"
  type        = string
  default     = null
}

# Config Rules Configuration
variable "create_encryption_config_rule" {
  description = "Create new encryption Config rule"
  type        = bool
  default     = true
}

variable "existing_encryption_config_rule" {
  description = "Name of existing encryption Config rule (required if create_encryption_config_rule=false)"
  type        = string
  default     = null
}

variable "create_retention_config_rule" {
  description = "Create new retention Config rule"
  type        = bool
  default     = true
}

variable "existing_retention_config_rule" {
  description = "Name of existing retention Config rule (required if create_retention_config_rule=false)"
  type        = string
  default     = null
}

# EventBridge Configuration
variable "create_eventbridge_rules" {
  description = "Create EventBridge scheduling rules (disable for manual invocation only)"
  type        = bool
  default     = true
}

variable "encryption_schedule_expression" {
  description = "Schedule for encryption compliance checks (cron or rate format)"
  type        = string
  default     = "cron(0 3 ? * SUN *)"
}

variable "retention_schedule_expression" {
  description = "Schedule for retention compliance checks (cron or rate format)"
  type        = string
  default     = "cron(0 4 ? * SUN *)"
}

# Lambda Configuration
variable "default_retention_days" {
  description = "Default retention period for remediated log groups (compliance threshold)"
  type        = number
  default     = 1

  validation {
    condition     = contains([1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 3653], var.default_retention_days)
    error_message = "Must be a valid CloudWatch Logs retention value."
  }
}

variable "lambda_log_retention_days" {
  description = "Retention period for LogGuardian Lambda function logs"
  type        = number
  default     = 30

  validation {
    condition     = contains([1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 3653], var.lambda_log_retention_days)
    error_message = "Must be a valid CloudWatch Logs retention value."
  }
}

variable "lambda_memory_size" {
  description = "Memory allocation for Lambda function (MB)"
  type        = number
  default     = 128

  validation {
    condition     = var.lambda_memory_size >= 128 && var.lambda_memory_size <= 10240
    error_message = "Memory must be between 128 and 10240 MB."
  }
}

variable "lambda_timeout" {
  description = "Lambda function timeout in seconds"
  type        = number
  default     = 60

  validation {
    condition     = var.lambda_timeout >= 1 && var.lambda_timeout <= 900
    error_message = "Timeout must be between 1 and 900 seconds."
  }
}

variable "log_level" {
  description = "Logging level for Lambda function (ERROR, WARN, INFO, DEBUG)"
  type        = string
  default     = "INFO"

  validation {
    condition     = contains(["ERROR", "WARN", "INFO", "DEBUG"], var.log_level)
    error_message = "Log level must be ERROR, WARN, INFO, or DEBUG."
  }
}

variable "dry_run" {
  description = "Enable dry-run mode (preview changes without applying)"
  type        = bool
  default     = false
}

variable "lambda_code_path" {
  description = "Path to Lambda build directory containing bootstrap binary"
  type        = string
  default     = "../build"
}

# S3 Lifecycle Configuration
variable "s3_expiration_days" {
  description = "Days after which Config data is permanently deleted (only applies to new buckets)"
  type        = number
  default     = 90

  validation {
    condition     = var.s3_expiration_days >= 1 && var.s3_expiration_days <= 3653
    error_message = "Expiration days must be between 1 and 3653."
  }
}

variable "enable_s3_lifecycle_rules" {
  description = "Enable S3 lifecycle rules for cost optimization (only applies to new buckets)"
  type        = bool
  default     = true
}

# Monitoring Configuration
variable "create_monitoring_dashboard" {
  description = "Create CloudWatch dashboard for monitoring"
  type        = bool
  default     = true
}

# Resource Tagging
variable "product_name" {
  description = "Product name for resource tagging"
  type        = string
  default     = "LogGuardian"
}

variable "owner" {
  description = "Owner/Team responsible for this deployment"
  type        = string
  default     = "DevOps"
}

variable "managed_by" {
  description = "How this stack is managed"
  type        = string
  default     = "Terraform"

  validation {
    condition     = contains(["Terraform", "SAM", "CloudFormation", "Manual"], var.managed_by)
    error_message = "Must be Terraform, SAM, CloudFormation, or Manual."
  }
}

variable "tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}
