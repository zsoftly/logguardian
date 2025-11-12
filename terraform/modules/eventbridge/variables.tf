variable "environment" {
  description = "Environment name (dev, staging, prod, test)"
  type        = string

  validation {
    condition     = can(regex("^(dev|staging|prod|test)$", var.environment))
    error_message = "Environment must be one of: dev, staging, prod, test."
  }
}

variable "create_eventbridge_rules" {
  description = "Whether to create EventBridge scheduling rules"
  type        = bool
  default     = true
}

variable "lambda_function_arn" {
  description = "ARN of Lambda function to invoke"
  type        = string

  validation {
    condition     = can(regex("^arn:aws:lambda:[a-z0-9-]+:[0-9]{12}:function:.+$", var.lambda_function_arn))
    error_message = "Lambda function ARN must be valid format."
  }
}

variable "lambda_function_name" {
  description = "Name of Lambda function"
  type        = string

  validation {
    condition     = can(regex("^[a-zA-Z0-9-_]+$", var.lambda_function_name))
    error_message = "Lambda function name must contain only alphanumeric characters, hyphens, or underscores."
  }
}

variable "encryption_config_rule" {
  description = "Name of encryption Config rule"
  type        = string

  validation {
    condition     = can(regex("^[a-zA-Z0-9._-]+$", var.encryption_config_rule))
    error_message = "Config rule name must be valid format."
  }
}

variable "retention_config_rule" {
  description = "Name of retention Config rule"
  type        = string

  validation {
    condition     = can(regex("^[a-zA-Z0-9._-]+$", var.retention_config_rule))
    error_message = "Config rule name must be valid format."
  }
}

variable "encryption_schedule" {
  description = "Schedule expression for encryption checks (cron or rate)"
  type        = string
  default     = "cron(0 3 ? * SUN *)"

  validation {
    condition     = can(regex("^(rate\\([0-9]+ (minute|minutes|hour|hours|day|days)\\)|cron\\(.+\\))$", var.encryption_schedule))
    error_message = "Must be a valid EventBridge schedule expression (rate or cron)."
  }
}

variable "retention_schedule" {
  description = "Schedule expression for retention checks (cron or rate)"
  type        = string
  default     = "cron(0 4 ? * SUN *)"

  validation {
    condition     = can(regex("^(rate\\([0-9]+ (minute|minutes|hour|hours|day|days)\\)|cron\\(.+\\))$", var.retention_schedule))
    error_message = "Must be a valid EventBridge schedule expression (rate or cron)."
  }
}

variable "tags" {
  description = "Additional tags to apply to resources"
  type        = map(string)
  default     = {}
}
