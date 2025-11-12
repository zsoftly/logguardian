variable "environment" {
  description = "Environment name (dev, staging, prod, test)"
  type        = string

  validation {
    condition     = can(regex("^(dev|staging|prod|test)$", var.environment))
    error_message = "Environment must be one of: dev, staging, prod, test."
  }
}

variable "create_config_service" {
  description = "Whether to create AWS Config service resources (recorder, delivery channel)"
  type        = bool
  default     = false
}

variable "config_bucket_name" {
  description = "S3 bucket name for Config delivery channel (required if create_config_service=true)"
  type        = string
  default     = null
}

variable "existing_config_role_arn" {
  description = "ARN of existing Config service role (used if create_config_service=false)"
  type        = string
  default     = null
}

variable "create_encryption_rule" {
  description = "Whether to create encryption Config rule"
  type        = bool
  default     = true
}

variable "existing_encryption_rule" {
  description = "Name of existing encryption Config rule (used if create_encryption_rule=false)"
  type        = string
  default     = null

  validation {
    condition     = var.existing_encryption_rule == null || can(regex("^[a-zA-Z0-9._-]+$", var.existing_encryption_rule))
    error_message = "Config rule name must contain only alphanumeric characters, dots, underscores, or hyphens."
  }
}

variable "create_retention_rule" {
  description = "Whether to create retention Config rule"
  type        = bool
  default     = true
}

variable "existing_retention_rule" {
  description = "Name of existing retention Config rule (used if create_retention_rule=false)"
  type        = string
  default     = null

  validation {
    condition     = var.existing_retention_rule == null || can(regex("^[a-zA-Z0-9._-]+$", var.existing_retention_rule))
    error_message = "Config rule name must contain only alphanumeric characters, dots, underscores, or hyphens."
  }
}

variable "default_retention_days" {
  description = "Minimum retention days for log groups (compliance threshold)"
  type        = number
  default     = 1

  validation {
    condition     = contains([1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 3653], var.default_retention_days)
    error_message = "Retention days must be a valid CloudWatch Logs retention value."
  }
}

variable "product_name" {
  description = "Product name for resource tagging"
  type        = string
  default     = "LogGuardian"
}

variable "tags" {
  description = "Additional tags to apply to resources"
  type        = map(string)
  default     = {}
}
