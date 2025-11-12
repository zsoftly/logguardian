variable "environment" {
  description = "Environment name"
  type        = string
  default     = "prod"
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "ca-central-1"
}

# Existing Infrastructure
variable "existing_kms_key_arn" {
  description = "ARN of existing KMS key"
  type        = string
}

variable "kms_key_alias" {
  description = "Alias of existing KMS key"
  type        = string
  default     = "alias/existing-cloudwatch-logs-key"
}

variable "existing_config_bucket" {
  description = "Name of existing Config S3 bucket"
  type        = string
}

variable "existing_config_role_arn" {
  description = "ARN of existing Config service role"
  type        = string
}

variable "existing_encryption_config_rule" {
  description = "Name of existing encryption Config rule"
  type        = string
}

variable "existing_retention_config_rule" {
  description = "Name of existing retention Config rule"
  type        = string
}

# Lambda Configuration
variable "default_retention_days" {
  description = "Default retention period"
  type        = number
  default     = 90
}

variable "log_level" {
  description = "Lambda log level"
  type        = string
  default     = "INFO"
}

variable "encryption_schedule" {
  description = "Encryption check schedule"
  type        = string
  default     = "cron(0 3 ? * SUN *)"
}

variable "retention_schedule" {
  description = "Retention check schedule"
  type        = string
  default     = "cron(0 4 ? * SUN *)"
}

variable "owner" {
  description = "Owner/Team"
  type        = string
  default     = "DevOps"
}

variable "additional_tags" {
  description = "Additional tags"
  type        = map(string)
  default     = {}
}
