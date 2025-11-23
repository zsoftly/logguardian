# ============================================
# Core Configuration
# ============================================

variable "environment" {
  description = "Deployment environment (dev, staging, prod, sandbox)"
  type        = string

  validation {
    condition     = contains(["dev", "staging", "prod", "sandbox"], var.environment)
    error_message = "Environment must be one of: dev, staging, prod, sandbox"
  }
}

variable "product_name" {
  description = "Product name for resource tagging and naming"
  type        = string
  default     = "LogGuardian"
}

variable "owner" {
  description = "Owner/Team responsible for the resources"
  type        = string
  default     = "Platform-Engineering"
}

# ============================================
# Lambda Configuration
# ============================================

variable "lambda_s3_bucket" {
  description = "S3 bucket containing the Lambda deployment package"
  type        = string
}

variable "lambda_s3_key" {
  description = "S3 key for the Lambda deployment package"
  type        = string
  default     = "logguardian-compliance.zip"
}

variable "lambda_memory_size" {
  description = "Lambda memory allocation in MB (Go is efficient, 128MB is usually sufficient)"
  type        = number
  default     = 128

  validation {
    condition     = var.lambda_memory_size >= 128 && var.lambda_memory_size <= 3008
    error_message = "Lambda memory must be between 128 MB and 3008 MB"
  }
}

variable "lambda_timeout" {
  description = "Lambda timeout in seconds"
  type        = number
  default     = 300

  validation {
    condition     = var.lambda_timeout >= 1 && var.lambda_timeout <= 900
    error_message = "Lambda timeout must be between 1 and 900 seconds"
  }
}

variable "lambda_log_level" {
  description = "Lambda logging level (ERROR, WARN, INFO, DEBUG)"
  type        = string
  default     = "INFO"

  validation {
    condition     = contains(["ERROR", "WARN", "INFO", "DEBUG"], var.lambda_log_level)
    error_message = "Log level must be one of: ERROR, WARN, INFO, DEBUG"
  }
}

variable "lambda_log_retention_days" {
  description = "Retention period for LogGuardian Lambda's own logs"
  type        = number
  default     = 7

  validation {
    condition     = contains([1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 3653], var.lambda_log_retention_days)
    error_message = "Lambda log retention must be a valid CloudWatch Logs retention period"
  }
}

variable "default_retention_days" {
  description = "Default retention period for log groups managed by LogGuardian"
  type        = number
  default     = 30

  validation {
    condition     = var.default_retention_days >= 1 && var.default_retention_days <= 3653
    error_message = "Default retention days must be between 1 and 3653 days"
  }
}

variable "batch_size" {
  description = "Number of log groups to process in parallel batches"
  type        = number
  default     = 10

  validation {
    condition     = var.batch_size > 0 && var.batch_size <= 100
    error_message = "Batch size must be between 1 and 100"
  }
}

# ============================================
# KMS Configuration
# ============================================

variable "create_kms_key" {
  description = "Create a new KMS key for CloudWatch Logs encryption. If false, existing_kms_key_arn must be provided."
  type        = bool
  default     = true
}

variable "existing_kms_key_arn" {
  description = "ARN of existing KMS key (required if create_kms_key = false)"
  type        = string
  default     = null

  validation {
    condition     = var.existing_kms_key_arn == null || can(regex("^arn:(aws|aws-cn|aws-us-gov):kms:[a-z0-9-]+:[0-9]{12}:key/[A-Fa-f0-9-]+$", var.existing_kms_key_arn))
    error_message = "KMS key ARN must be a valid ARN format (supports aws, aws-cn, aws-us-gov partitions)"
  }
}

variable "kms_key_alias" {
  description = "Alias for the KMS key (without 'alias/' prefix)"
  type        = string
  default     = null
}

variable "kms_deletion_window_days" {
  description = "KMS key deletion window in days"
  type        = number
  default     = 30

  validation {
    condition     = var.kms_deletion_window_days >= 7 && var.kms_deletion_window_days <= 30
    error_message = "KMS deletion window must be between 7 and 30 days"
  }
}

# ============================================
# AWS Config Configuration
# ============================================

variable "create_config_service" {
  description = "Create AWS Config service resources (recorder, delivery channel)"
  type        = bool
  default     = false
}

variable "existing_config_bucket" {
  description = "Existing S3 bucket for Config snapshots (required if create_config_service = false)"
  type        = string
  default     = null
}

variable "existing_config_service_role_arn" {
  description = "Existing IAM role ARN for Config service (required if create_config_service = false)"
  type        = string
  default     = null

  validation {
    condition     = var.existing_config_service_role_arn == null || can(regex("^arn:(aws|aws-cn|aws-us-gov):iam::[0-9]{12}:role/[^/]+$", var.existing_config_service_role_arn))
    error_message = "Config service role ARN must be a valid IAM role ARN (supports aws, aws-cn, aws-us-gov partitions)"
  }
}

variable "config_bucket_expiration_days" {
  description = "S3 lifecycle expiration for Config snapshots (cost optimization)"
  type        = number
  default     = 90

  validation {
    condition     = var.config_bucket_expiration_days >= 1 && var.config_bucket_expiration_days <= 3653
    error_message = "Config bucket expiration must be between 1 and 3653 days"
  }
}

# ============================================
# Config Rules Configuration
# ============================================

variable "create_config_rules" {
  description = "Create AWS Config rules for compliance monitoring"
  type        = bool
  default     = true
}

variable "existing_encryption_config_rule" {
  description = "Name of existing encryption Config rule (if create_config_rules = false)"
  type        = string
  default     = null
}

variable "existing_retention_config_rule" {
  description = "Name of existing retention Config rule (if create_config_rules = false)"
  type        = string
  default     = null
}

# ============================================
# EventBridge Scheduling Configuration
# ============================================

variable "create_eventbridge_rules" {
  description = "Create EventBridge scheduled rules for automated compliance checks"
  type        = bool
  default     = true
}

variable "encryption_schedule_expression" {
  description = "EventBridge schedule expression for encryption compliance checks"
  type        = string
  default     = "cron(0 2 ? * SUN *)" # Weekly Sunday at 2 AM UTC

  validation {
    condition     = can(regex("^(cron\\(.*\\)|rate\\(.*\\))$", var.encryption_schedule_expression))
    error_message = "Schedule expression must be a valid cron() or rate() expression"
  }
}

variable "retention_schedule_expression" {
  description = "EventBridge schedule expression for retention compliance checks"
  type        = string
  default     = "cron(0 3 ? * SUN *)" # Weekly Sunday at 3 AM UTC

  validation {
    condition     = can(regex("^(cron\\(.*\\)|rate\\(.*\\))$", var.retention_schedule_expression))
    error_message = "Schedule expression must be a valid cron() or rate() expression"
  }
}

# ============================================
# Monitoring Configuration
# ============================================

variable "create_monitoring_dashboard" {
  description = "Create CloudWatch dashboard for monitoring LogGuardian"
  type        = bool
  default     = true
}

variable "enable_cloudwatch_alarms" {
  description = "Enable CloudWatch alarms for Lambda errors and throttling"
  type        = bool
  default     = true
}

variable "alarm_sns_topic_arn" {
  description = "SNS topic ARN for CloudWatch alarms and Config remediation notifications (optional)"
  type        = string
  default     = null

  validation {
    condition     = var.alarm_sns_topic_arn == null || can(regex("^arn:(aws|aws-cn|aws-us-gov):sns:[A-Za-z0-9-]+:[0-9]{12}:.*$", var.alarm_sns_topic_arn))
    error_message = "SNS topic ARN must be a valid ARN format (supports aws, aws-cn, aws-us-gov partitions)"
  }
}

# ============================================
# Advanced Configuration
# ============================================

variable "supported_regions" {
  description = "List of AWS regions where log groups will be managed"
  type        = list(string)
  default     = []

  validation {
    condition     = alltrue([for r in var.supported_regions : can(regex("^[a-z]{2}-[a-z]+-[0-9]+$", r))])
    error_message = "All regions must be valid AWS region codes (e.g., ca-central-1)"
  }
}

variable "additional_lambda_env_vars" {
  description = "Additional environment variables for the Lambda function"
  type        = map(string)
  default     = {}
}

variable "additional_tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}

variable "config_recorder_resource_types" {
  description = "List of resource types to record in AWS Config. Empty list records all supported types."
  type        = list(string)
  default = [
    "AWS::Logs::LogGroup"
  ]

  validation {
    condition = (
      length(var.config_recorder_resource_types) == 0 ||
      alltrue([
        for rt in var.config_recorder_resource_types :
        can(regex("^AWS::[A-Za-z0-9]+::[A-Za-z0-9]+$", rt))
      ])
    )
    error_message = "Resource types must follow AWS Config format (e.g., 'AWS::Logs::LogGroup'). Use empty list to record all types."
  }
}
