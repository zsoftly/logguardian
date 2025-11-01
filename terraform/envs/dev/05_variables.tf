variable "tags" {
  description = "A map of tags to add to all resources"
  type        = map(string)
}

variable "region" {
  description = "AWS region to deploy resources"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "container_image" {
  description = "Container image URL for LogGuardian"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID where ECS tasks will run (defaults to default VPC if not specified)"
  type        = string
  default     = null
}

variable "subnet_ids" {
  description = "List of subnet IDs for ECS tasks (defaults to public subnets in selected VPC if not specified)"
  type        = list(string)
  default     = null
}

variable "assign_public_ip" {
  description = "Assign public IP to ECS tasks (required for public subnets without NAT)"
  type        = bool
  default     = true
}

variable "cpu" {
  description = "Fargate CPU units (256, 512, 1024, 2048, 4096)"
  type        = string
  default     = "256"
}

variable "memory" {
  description = "Fargate memory in MB (512, 1024, 2048, etc.)"
  type        = string
  default     = "512"
}

variable "enable_spot" {
  description = "Enable Fargate Spot for cost savings"
  type        = bool
  default     = true
}

variable "log_retention_days" {
  description = "CloudWatch log retention in days"
  type        = number
  default     = 30
}

variable "config_rule_name" {
  description = "Default AWS Config rule name to evaluate"
  type        = string
  default     = "cw-lg-retention-min"
}

variable "enable_scheduling" {
  description = "Enable EventBridge scheduled triggers for automated compliance checks"
  type        = bool
  default     = true
}

variable "encryption_schedule_expression" {
  description = <<-EOT
    Schedule expression for encryption compliance checks.

    AWS Config evaluates compliance every 24 hours minimum. Stagger encryption
    and retention checks by 1 hour to distribute load.

    Common patterns (all times in UTC):

    DAILY (fastest option - runs every 24 hours):
      - "rate(1 day)" or "cron(0 2 * * ? *)" - Every day at 2 AM UTC

    WEEKLY (recommended - lower cost):
      - "cron(0 3 ? * SUN *)" - Sunday at 3 AM UTC (Saturday 10 PM EST)
      - "cron(0 2 ? * MON *)" - Monday at 2 AM UTC (Sunday 9 PM EST)

    MONTHLY (lowest cost):
      - "cron(0 2 1 * ? *)" - 1st of month at 2 AM UTC
      - "cron(0 2 15 * ? *)" - 15th of month at 2 AM UTC

    BUSINESS HOURS (weekdays only):
      - "cron(0 14 ? * MON-FRI *)" - Weekdays at 2 PM UTC (9 AM EST)

    Format: cron(Minutes Hours Day-of-month Month Day-of-week Year)
    Note: Use '?' for either Day-of-month or Day-of-week (not both)
  EOT
  type        = string
  default     = "cron(0 3 ? * SUN *)"
}

variable "retention_schedule_expression" {
  description = <<-EOT
    Schedule expression for retention compliance checks.

    IMPORTANT: Stagger this 1 hour after encryption checks to avoid concurrent runs.

    Common patterns (staggered examples):

    DAILY (if encryption runs at 2 AM):
      - "cron(0 3 * * ? *)" - Every day at 3 AM UTC (1 hour after encryption)

    WEEKLY (if encryption runs Sunday 3 AM):
      - "cron(0 4 ? * SUN *)" - Sunday at 4 AM UTC (1 hour after encryption)
      - Or different day: "cron(0 2 ? * MON *)" - Monday at 2 AM UTC

    MONTHLY (if encryption runs 1st at 2 AM):
      - "cron(0 3 1 * ? *)" - 1st of month at 3 AM UTC (1 hour after encryption)
      - Or mid-month: "cron(0 2 15 * ? *)" - 15th of month at 2 AM UTC

    Format: cron(Minutes Hours Day-of-month Month Day-of-week Year)
  EOT
  type        = string
  default     = "cron(0 4 ? * SUN *)"
}

variable "encryption_config_rule" {
  description = "AWS Config rule name for encryption compliance"
  type        = string
  default     = "logguardian-encryption-dev"
}

variable "retention_config_rule" {
  description = "AWS Config rule name for retention compliance"
  type        = string
  default     = "logguardian-retention-dev"
}

variable "dry_run" {
  description = <<-EOT
    Enable dry-run mode for scheduled compliance checks.

    When true: Tasks will only report compliance violations without making changes.
    When false: Tasks will automatically remediate non-compliant log groups.

    Recommended: Start with true for testing, then set to false for production automation.
  EOT
  type        = bool
  default     = true
}
