variable "environment" {
  description = "Environment name (dev, staging, prod, test)"
  type        = string

  validation {
    condition     = can(regex("^(dev|staging|prod|test)$", var.environment))
    error_message = "Environment must be one of: dev, staging, prod, test."
  }
}

variable "create_dashboard" {
  description = "Whether to create CloudWatch dashboard"
  type        = bool
  default     = true
}

variable "lambda_function_name" {
  description = "Name of Lambda function to monitor"
  type        = string

  validation {
    condition     = can(regex("^[a-zA-Z0-9-_]+$", var.lambda_function_name))
    error_message = "Lambda function name must be valid format."
  }
}

variable "log_group_name" {
  description = "CloudWatch Log Group name for Lambda logs"
  type        = string

  validation {
    condition     = can(regex("^/aws/lambda/.+$", var.log_group_name))
    error_message = "Log group name must be valid Lambda log group format."
  }
}

variable "product_name" {
  description = "Product name for dashboard naming"
  type        = string
  default     = "LogGuardian"
}

variable "tags" {
  description = "Additional tags to apply to resources"
  type        = map(string)
  default     = {}
}
